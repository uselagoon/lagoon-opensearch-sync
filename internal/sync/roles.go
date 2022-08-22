package sync

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/uselagoon/lagoon-opensearch-sync/internal/keycloak"
	"github.com/uselagoon/lagoon-opensearch-sync/internal/opensearch"
	"go.uber.org/zap"
)

// generateIndexPatterns returns a slice of index pattern strings generated
// from the given lagoon-projects keycloak group attribute. This attribute
// contains a comma-separated list of Lagoon project IDs.
func generateIndexPatterns(log *zap.Logger, lpa string,
	projectNames map[int]string) ([]string, error) {
	var buf bytes.Buffer
	// get the project IDs
	if _, err := fmt.Fprintf(&buf, "[%s]", lpa); err != nil {
		return nil,
			fmt.Errorf("couldn't format lagoon-projects attribute: %v", err)
	}
	var pids []int
	if err := json.Unmarshal(buf.Bytes(), &pids); err != nil {
		return nil,
			fmt.Errorf("couldn't unmarshal lagoon-projects attribute: %v", err)
	}
	var patterns []string
	for _, pid := range pids {
		name, ok := projectNames[pid]
		if !ok {
			log.Debug("invalid project ID in lagoon-projects group attribute",
				zap.Int("projectID", pid))
			continue
		}
		patterns = append(patterns,
			fmt.Sprintf(`/^(application|container|lagoon|router)-logs-%s-_-.+/`, name))
	}
	return patterns, nil
}

// isProjectGroup inspects the given group to determine if it is a
// project-default-group type.
func isProjectGroup(log *zap.Logger, group keycloak.Group) bool {
	t, ok := group.Attributes["type"]
	if !ok {
		return false
	}
	if len(t) != 1 {
		log.Debug(`group attribute "type" invalid`,
			zap.String("group name", group.Name), zap.Int("attribute length", len(t)))
		return false
	}
	if t[0] != "project-default-group" {
		log.Debug(`group attribute "type" unknown`,
			zap.String("group name", group.Name), zap.String("type", t[0]))
		return false
	}
	return true
}

// isInt returns true if the given string looks like a base-10 integer.
func isInt(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

// generateProjectGroupRole constructs an opensearch.Role from the given
// keycloak group corresponding to a Lagoon project group.
func generateProjectGroupRole(
	group keycloak.Group) (string, *opensearch.Role, error) {
	pAttr, ok := group.Attributes["lagoon-projects"]
	if !ok {
		return "", nil, fmt.Errorf("missing lagoon-projects attribute")
	}
	if len(pAttr) != 1 || !isInt(pAttr[0]) {
		return "", nil, fmt.Errorf("invalid lagoon-projects attribute")
	}
	return fmt.Sprintf("p%s", pAttr[0]), &opensearch.Role{
		ClusterPermissions: []string{
			"cluster:admin/opendistro/reports/menu/download",
		},
		IndexPermissions: []opensearch.IndexPermission{
			{
				AllowedActions: []string{
					"read",
					"indices:monitor/settings/get",
				},
				IndexPatterns: []string{
					fmt.Sprintf(
						`/^(application|container|lagoon|router)-logs-%s-_-.+/`,
						strings.TrimPrefix(group.Name, "project-")),
				},
			},
		},
		TenantPermissions: []opensearch.TenantPermission{
			{
				AllowedActions: []string{"kibana_all_read"},
				TenantPatterns: []string{"global_tenant"},
			},
		},
	}, nil
}

// generateRegularGroupRole constructs an opensearch.Role from the given
// keycloak group corresponding to a Lagoon group.
func generateRegularGroupRole(log *zap.Logger, projectNames map[int]string,
	group keycloak.Group) (string, *opensearch.Role, error) {
	// get lagoon-projects attribute
	lpa, ok := group.Attributes["lagoon-projects"]
	if !ok {
		return "", nil, fmt.Errorf("missing lagoon-projects attribute")
	}
	if len(lpa) != 1 {
		return "", nil, fmt.Errorf("invalid lagoon-projects attribute")
	}
	// calculate index patterns from lagoon-projects attribute
	indexPatterns, err := generateIndexPatterns(log, lpa[0], projectNames)
	if err != nil {
		return "", nil, fmt.Errorf("couldn't generate index patterns: %v", err)
	}
	// the Opensearch API is picky about the structure of create group requests,
	// so ensure that the index_permissions field is only set if there are any
	// index patterns. Also it cannot be omitted, so can't be nil.
	var indexPermissions []opensearch.IndexPermission
	if len(indexPatterns) == 0 {
		indexPermissions = []opensearch.IndexPermission{}
	} else {
		indexPermissions = []opensearch.IndexPermission{
			{
				AllowedActions: []string{
					"read",
					"indices:monitor/settings/get",
				},
				IndexPatterns: indexPatterns,
			},
		}
	}
	return group.Name, &opensearch.Role{
		ClusterPermissions: []string{
			"cluster:admin/opendistro/reports/menu/download",
		},
		IndexPermissions: indexPermissions,
		TenantPermissions: []opensearch.TenantPermission{
			{
				AllowedActions: []string{"kibana_all_write"},
				TenantPatterns: []string{group.Name},
			},
		},
	}, nil
}

// generateRoles returns a slice of roles generated from the given slice of
// keycloak Groups.
func generateRoles(log *zap.Logger, groups []keycloak.Group,
	projectNames map[int]string) map[string]opensearch.Role {
	roles := map[string]opensearch.Role{}
	for _, group := range groups {
		// figure out if this is a regular group or project group
		switch {
		case group.Name == "lagoonadmin":
			// lagoonadmin is a special role used by Lagoon
			roles[group.Name] = opensearch.Role{
				ClusterPermissions: []string{
					"cluster:admin/opendistro/reports/menu/download",
				},
				IndexPermissions: []opensearch.IndexPermission{},
				TenantPermissions: []opensearch.TenantPermission{
					{
						AllowedActions: []string{"kibana_all_write"},
						TenantPatterns: []string{"lagoonadmin"},
					},
				},
			}
		case isProjectGroup(log, group):
			name, role, err := generateProjectGroupRole(group)
			if err != nil {
				log.Warn("couldn't generate role for project group",
					zap.String("group name", group.Name), zap.Error(err))
				continue
			}
			roles[name] = *role
		default:
			name, role, err := generateRegularGroupRole(log, projectNames, group)
			if err != nil {
				log.Warn("couldn't generate role for regular group",
					zap.String("group name", group.Name), zap.Error(err))
				continue
			}
			roles[name] = *role
		}
	}
	return roles
}

// calculateRoleDiff returns a map of opensearch roles which should be created,
// and a slice of role names which should be deleted, in order to reconcile
// existing with required.
func calculateRoleDiff(existing, required map[string]opensearch.Role) (
	map[string]opensearch.Role, []string) {
	// calculate roles to create
	toCreate := map[string]opensearch.Role{}
	for name, rRole := range required {
		eRole, ok := existing[name]
		if !ok || !rolesEqual(eRole, rRole) {
			toCreate[name] = rRole
		}
	}
	// calculate roles to delete
	var toDelete []string
	for name, eRole := range existing {
		rRole, ok := required[name]
		if !ok || !rolesEqual(rRole, eRole) {
			// don't delete unnecessarily. create action in opensearch is actually
			// create/replace.
			// https://opensearch.org/docs/2.2/security-plugin/access-control
			// 	/api#create-role
			if _, ok := toCreate[name]; !ok {
				toDelete = append(toDelete, name)
			}
		}
	}
	return toCreate, toDelete
}

// given a map of opensearch roles, return those that are not static or
// reserved.
func filterRoles(
	roles map[string]opensearch.Role) map[string]opensearch.Role {
	valid := map[string]opensearch.Role{}
	for name, role := range roles {
		if role.Static || role.Reserved {
			continue
		}
		valid[name] = role
	}
	return valid
}

// syncRoles reconciles Opensearch roles with Lagoon keycloak and projects.
func syncRoles(ctx context.Context, log *zap.Logger, groups []keycloak.Group,
	projectNames map[int]string, o OpensearchService, dryRun bool) {
	existing, err := o.Roles(ctx)
	if err != nil {
		log.Error("couldn't get roles from Opensearch", zap.Error(err))
		return
	}
	// ignore non-lagoon roles
	existing = filterRoles(existing)
	// generate the roles required by Lagoon
	required := generateRoles(log, groups, projectNames)
	// calculate roles to add/remove
	toCreate, toDelete := calculateRoleDiff(existing, required)
	// remove any extraneous roles
	for _, name := range toDelete {
		if dryRun {
			log.Info("dry run mode: not deleting role",
				zap.String("name", name))
			continue
		}
		err = o.DeleteRole(ctx, name)
		if err != nil {
			log.Warn("couldn't delete role", zap.Error(err))
		}
	}
	for name, role := range toCreate {
		if dryRun {
			log.Info("dry run mode: not creating role",
				zap.String("name", name))
			continue
		}
		err = o.CreateRole(ctx, name, &role)
		if err != nil {
			log.Warn("couldn't create role", zap.Error(err))
		}
	}
}
