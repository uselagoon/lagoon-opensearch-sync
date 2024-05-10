package sync

import (
	"context"
	"fmt"
	"strings"

	"github.com/uselagoon/lagoon-opensearch-sync/internal/keycloak"
	"github.com/uselagoon/lagoon-opensearch-sync/internal/opensearch"
	"go.uber.org/zap"
)

// generateIndexPermissionPatterns returns a slice of index pattern strings
// in regular expressions format generated from the given slice of project IDs.
// https://www.elastic.co/guide/en/elasticsearch/reference/7.10/defining-roles.html#roles-indices-priv
func generateIndexPermissionPatterns(
	log *zap.Logger,
	pids []int,
	projectNames map[int]string,
) []string {
	var patterns []string
	for _, pid := range pids {
		name, ok := projectNames[pid]
		if !ok {
			// If you see this warning it means that a project ID appears in the
			// kc_group_projects table that does not appear in the projects table in
			// the Lagoon API DB.
			// This is likely a bug in Lagoon causing loss of referential integrity,
			// as there is no foreign key constraint to enforce valid project IDs in
			// the group mapping.
			log.Warn("invalid project ID when generating index permission patterns",
				zap.Int("projectID", pid))
			continue
		}
		patterns = append(patterns,
			fmt.Sprintf(`/^(application|container|lagoon|router)-logs-%s-_-.+/`, name))
	}
	return patterns
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

// isLagoonGroup inspects the given group to determine if it is a Lagoon group.
//
// It checks if the group ID appears in the groupProjectsMap.
func isLagoonGroup(
	group keycloak.Group,
	groupProjectsMap map[string][]int,
) bool {
	_, ok := groupProjectsMap[group.ID]
	return ok
}

// projectGroupRoleName generates the name of a project group role from the
// ID of the group's project.
func projectGroupRoleName(
	group keycloak.Group,
	groupProjectsMap map[string][]int,
) (string, error) {
	projectIDs, ok := groupProjectsMap[group.ID]
	if !ok {
		return "", fmt.Errorf("missing project group ID %s in groupProjectsMap",
			group.ID)
	}
	if len(projectIDs) != 1 {
		return "", fmt.Errorf("too many projects in group ID %s: %d", group.ID,
			len(projectIDs))
	}
	if projectIDs[0] < 0 {
		return "", fmt.Errorf("invalid project ID in group ID %s: %d", group.ID,
			projectIDs[0])
	}
	return fmt.Sprintf("p%d", projectIDs[0]), nil
}

// generateProjectGroupRole constructs an opensearch.Role from the given
// keycloak group corresponding to a Lagoon project group.
func generateProjectGroupRole(
	group keycloak.Group,
	groupProjectsMap map[string][]int,
) (string, *opensearch.Role, error) {
	name, err := projectGroupRoleName(group, groupProjectsMap)
	if err != nil {
		return "", nil,
			fmt.Errorf("couldn't generate project group role name: %v", err)
	}
	return name, &opensearch.Role{
		RolePermissions: opensearch.RolePermissions{
			// use an empty slice instead of omitting this entirely because the
			// Opensearch API errors if this field is omitted.
			ClusterPermissions: []string{},
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
		},
	}, nil
}

// generateRegularGroupRole constructs an opensearch.Role from the given
// keycloak group corresponding to a Lagoon group.
func generateRegularGroupRole(
	log *zap.Logger,
	group keycloak.Group,
	projectNames map[int]string,
	groupProjectsMap map[string][]int,
) (string, *opensearch.Role, error) {
	pids, ok := groupProjectsMap[group.ID]
	if !ok {
		return "", nil, fmt.Errorf("missing project group ID %s in groupProjectsMap",
			group.ID)
	}
	// calculate index patterns from project IDs
	indexPatterns := generateIndexPermissionPatterns(log, pids, projectNames)
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
		RolePermissions: opensearch.RolePermissions{
			// Allow users to read and download Reports
			// https://github.com/opensearch-project/security/blob/2.7.0.0/config/
			// 		roles.yml#L126-L132
			ClusterPermissions: []string{
				"cluster:admin/opendistro/reports/instance/list",
				"cluster:admin/opendistro/reports/instance/get",
				"cluster:admin/opendistro/reports/menu/download",
			},
			IndexPermissions: indexPermissions,
			TenantPermissions: []opensearch.TenantPermission{
				{
					AllowedActions: []string{"kibana_all_write"},
					TenantPatterns: []string{group.Name},
				},
			},
		},
	}, nil
}

// generateRoles returns a slice of roles generated from the given slice of
// keycloak Groups.
//
// Any groups which are not recognized as either project groups or regular
// Lagoon groups are ignored.
func generateRoles(
	log *zap.Logger,
	groups []keycloak.Group,
	projectNames map[int]string,
	groupProjectsMap map[string][]int,
) map[string]opensearch.Role {
	roles := map[string]opensearch.Role{}
	var name string
	var role *opensearch.Role
	var err error
	for _, group := range groups {
		if isProjectGroup(log, group) {
			name, role, err = generateProjectGroupRole(group, groupProjectsMap)
			if err != nil {
				log.Warn("couldn't generate role for project group",
					zap.String("group name", group.Name), zap.Error(err))
				continue
			}
		} else if isLagoonGroup(group, groupProjectsMap) {
			name, role, err =
				generateRegularGroupRole(log, group, projectNames, groupProjectsMap)
			if err != nil {
				log.Warn("couldn't generate role for regular group",
					zap.String("group name", group.Name), zap.Error(err))
				continue
			}
		}
		if role != nil {
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
func syncRoles(
	ctx context.Context,
	log *zap.Logger,
	groups []keycloak.Group,
	projectNames map[int]string,
	roles map[string]opensearch.Role,
	groupProjectsMap map[string][]int,
	o OpensearchService,
	dryRun bool,
) {
	// ignore non-lagoon roles
	existing := filterRoles(roles)
	// generate the roles required by Lagoon
	required := generateRoles(log, groups, projectNames, groupProjectsMap)
	// calculate roles to add/remove
	toCreate, toDelete := calculateRoleDiff(existing, required)
	var err error
	for _, name := range toDelete {
		if dryRun {
			log.Info("dry run mode: not deleting role", zap.String("name", name))
			continue
		}
		err = o.DeleteRole(ctx, name)
		if err != nil {
			log.Warn("couldn't delete role", zap.Error(err))
			continue
		}
		log.Info("deleted role", zap.String("name", name))
	}
	for name, role := range toCreate {
		if dryRun {
			log.Info("dry run mode: not creating role", zap.String("name", name))
			continue
		}
		err = o.CreateRole(ctx, name, &role)
		if err != nil {
			log.Warn("couldn't create role", zap.Error(err))
			continue
		}
		log.Info("created role", zap.String("name", name))
	}
}
