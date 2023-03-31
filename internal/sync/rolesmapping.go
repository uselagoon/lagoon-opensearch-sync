package sync

import (
	"context"

	"github.com/uselagoon/lagoon-opensearch-sync/internal/keycloak"
	"github.com/uselagoon/lagoon-opensearch-sync/internal/opensearch"
	"go.uber.org/zap"
)

// rolesMappingEqual checks the fields Lagoon cares about for functional
// equality.
func rolesMappingEqual(a, b opensearch.RoleMapping) bool {
	if !stringSliceEqual(a.BackendRoles, b.BackendRoles) {
		return false
	}
	if a.Hidden != b.Hidden {
		return false
	}
	if a.Reserved != b.Reserved {
		return false
	}
	return true
}

// calculateRoleMappingDiff returns a map of opensearch rolesmapping which
// should be created, and a slice of rolemapping names which should be deleted,
// in order to reconcile existing with required.
func calculateRoleMappingDiff(
	existing, required map[string]opensearch.RoleMapping) (
	map[string]opensearch.RoleMapping, []string) {
	// calculate rolesmapping to create
	toCreate := map[string]opensearch.RoleMapping{}
	for name, rRoleMapping := range required {
		eRoleMapping, ok := existing[name]
		if !ok || !rolesMappingEqual(eRoleMapping, rRoleMapping) {
			toCreate[name] = rRoleMapping
		}
	}
	// calculate rolesmapping to delete
	var toDelete []string
	for name, eRoleMapping := range existing {
		rRoleMapping, ok := required[name]
		if !ok || !rolesMappingEqual(rRoleMapping, eRoleMapping) {
			// don't delete unnecessarily. create action in opensearch is actually
			// create/replace.
			// https://opensearch.org/docs/2.2/security-plugin/access-control
			// 	/api#create-role-mapping
			if _, ok := toCreate[name]; !ok {
				toDelete = append(toDelete, name)
			}
		}
	}
	return toCreate, toDelete
}

// generateRolesMapping returns a slice of rolesmapping generated from the
// given slice of keycloak Groups.
//
// Any groups which are not recognized as either project groups or regular
// Lagoon groups are ignored.
func generateRolesMapping(log *zap.Logger,
	groups []keycloak.Group) map[string]opensearch.RoleMapping {
	rolesmapping := map[string]opensearch.RoleMapping{}
	for _, group := range groups {
		// figure out if this is a regular group or project group
		if isProjectGroup(log, group) {
			name, err := projectGroupRoleName(group)
			if err != nil {
				log.Warn("couldn't generate project group role name", zap.Error(err),
					zap.String("group name", group.Name))
				continue
			}
			rolesmapping[name] = opensearch.RoleMapping{
				RoleMappingPermissions: opensearch.RoleMappingPermissions{
					BackendRoles:    []string{name},
					AndBackendRoles: []string{},
					Hosts:           []string{},
					Users:           []string{},
				},
			}
		} else if isLagoonGroup(group) {
			rolesmapping[group.Name] = opensearch.RoleMapping{
				RoleMappingPermissions: opensearch.RoleMappingPermissions{
					BackendRoles:    []string{group.Name},
					AndBackendRoles: []string{},
					Hosts:           []string{},
					Users:           []string{},
				},
			}
		}
	}
	return rolesmapping
}

// given a map of opensearch rolesmapping, return those that are not reserved
// or hidden.
func filterRolesMapping(rolesmapping map[string]opensearch.RoleMapping,
	roles map[string]opensearch.Role) map[string]opensearch.RoleMapping {
	valid := map[string]opensearch.RoleMapping{}
	for name, rolemapping := range rolesmapping {
		if rolemapping.Reserved || rolemapping.Hidden {
			continue
		}
		// for some reason even a "reserved" RoleMapping can have reserved=false,
		// so we need to inspect the corresponding Role
		if role, ok := roles[name]; ok {
			if role.Reserved || role.Static {
				continue
			}
		}
		valid[name] = rolemapping
	}
	return valid
}

// syncRolesmapping reconciles Opensearch rolesmapping with Lagoon keycloak
// groups.
func syncRolesMapping(ctx context.Context, log *zap.Logger, groups []keycloak.Group,
	projectNames map[int]string, roles map[string]opensearch.Role,
	o OpensearchService, dryRun bool) {
	// get rolesmapping from Opensearch
	existing, err := o.RolesMapping(ctx)
	if err != nil {
		log.Error("couldn't get rolesmapping from Opensearch", zap.Error(err))
		return
	}
	// ignore non-lagoon rolesmapping
	existing = filterRolesMapping(existing, roles)
	// generate the rolesmapping required by Lagoon
	required := generateRolesMapping(log, groups)
	// calculate rolesmapping to add/remove
	toCreate, toDelete := calculateRoleMappingDiff(existing, required)
	for _, name := range toDelete {
		if dryRun {
			log.Info("dry run mode: not deleting rolemapping",
				zap.String("name", name))
			continue
		}
		err = o.DeleteRoleMapping(ctx, name)
		if err != nil {
			log.Warn("couldn't delete rolemapping", zap.Error(err))
			continue
		}
		log.Info("deleted rolemapping", zap.String("name", name))
	}
	for name, rolemapping := range toCreate {
		if dryRun {
			log.Info("dry run mode: not creating rolemapping",
				zap.String("name", name))
			continue
		}
		err = o.CreateRoleMapping(ctx, name, &rolemapping)
		if err != nil {
			log.Warn("couldn't create rolemapping", zap.Error(err))
			continue
		}
		log.Info("created rolemapping", zap.String("name", name))
	}
}
