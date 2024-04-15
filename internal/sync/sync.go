// Package sync implements synchronization of state from Lagoon to Opensearch.
package sync

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/uselagoon/lagoon-opensearch-sync/internal/keycloak"
	"github.com/uselagoon/lagoon-opensearch-sync/internal/lagoondb"
	"github.com/uselagoon/lagoon-opensearch-sync/internal/opensearch"
	"go.uber.org/zap"
)

// KeycloakService defines the Keycloak service interface.
type KeycloakService interface {
	Groups(context.Context) ([]keycloak.Group, error)
}

// LagoonDBService defines the Lagoon database service interface.
type LagoonDBService interface {
	Projects(context.Context) ([]lagoondb.Project, error)
}

// OpensearchService defines the Opensearch service interface.
type OpensearchService interface {
	Tenants(context.Context) (map[string]opensearch.Tenant, error)
	CreateTenant(context.Context, string, *opensearch.Tenant) error
	DeleteTenant(context.Context, string) error

	Roles(context.Context) (map[string]opensearch.Role, error)
	CreateRole(context.Context, string, *opensearch.Role) error
	DeleteRole(context.Context, string) error

	RolesMapping(context.Context) (map[string]opensearch.RoleMapping, error)
	CreateRoleMapping(context.Context, string, *opensearch.RoleMapping) error
	DeleteRoleMapping(context.Context, string) error

	IndexTemplates(context.Context) (map[string]opensearch.IndexTemplate, error)
	CreateIndexTemplate(context.Context, string, *opensearch.IndexTemplate) error
	DeleteIndexTemplate(context.Context, string) error

	IndexPatterns(context.Context) (map[string]map[string][]string, error)
}

// DashboardsService defines the Opensearch Dashboards service interface.
type DashboardsService interface {
	DeleteIndexPattern(context.Context, string, string) error
	CreateIndexPattern(context.Context, string, string) error
}

// Sync will read the Lagoon state from the LagoonDBService and KeycloakService,
// and then configure the OpensearchService as required.
func Sync(ctx context.Context, log *zap.Logger, l LagoonDBService,
	k KeycloakService, o OpensearchService, d DashboardsService, dryRun bool,
	objects []string, legacyIndexPatternDelimiter bool) error {
	// get projects from Lagoon
	projects, err := l.Projects(ctx)
	if err != nil {
		return fmt.Errorf("couldn't get projects: %v", err)
	}
	// https://github.com/uselagoon/lagoon/blob/
	// 	7dd4eb3b695bd507f25de5d7ea49d6601a229b87/services/api/src/resources/
	// 	group/opendistroSecurity.ts#L31-L34
	lagoonName := regexp.MustCompile(`[^0-9a-z-]`)
	// generate project ID -> name map
	projectNames := map[int]string{}
	for _, project := range projects {
		// munge the project name in a Lagoon-compatible manner
		projectNames[project.ID] =
			lagoonName.ReplaceAllLiteralString(strings.ToLower(project.Name), `-`)
	}
	// get groups from Keycloak
	groups, err := k.Groups(ctx)
	if err != nil {
		return fmt.Errorf("couldn't get groups: %v", err)
	}
	// Work around security-dashboards-plugin bug by ignoring "global" group when
	// creating tenants and index patterns:
	// * Users in the "global" group will have to use the reserved Global Tenant.
	// * Users will still have access to their logs, but not their
	//   project-specific index patterns.
	// https://github.com/opensearch-project/security-dashboards-plugin/issues/1411
	//
	// TODO: once the upstream bug is closed, target the global tenant correctly
	// (however they fix it) and remove this workaround.
	var groupsSansGlobal []keycloak.Group
	for i := range groups {
		if groups[i].Name == "global" {
			continue
		}
		groupsSansGlobal = append(groupsSansGlobal, groups[i])
	}
	// Get roles from Opensearch. Getting this data here is an optimisation
	// because both syncRoles and syncRolesMapping use this data and this way we
	// only need to request it from Opensearch once.
	roles, err := o.Roles(ctx)
	if err != nil {
		return fmt.Errorf("couldn't get roles: %v", err)
	}
	for _, object := range objects {
		switch object {
		case "tenants":
			syncTenants(ctx, log, groupsSansGlobal, o, dryRun)
		case "roles":
			syncRoles(ctx, log, groups, projectNames, roles, o, dryRun)
		case "rolesmapping":
			syncRolesMapping(ctx, log, groups, roles, o, dryRun)
		case "indexpatterns":
			syncIndexPatterns(ctx, log, groupsSansGlobal, projectNames, o, d, dryRun,
				legacyIndexPatternDelimiter)
		case "indextemplates":
			syncIndexTemplates(ctx, log, o, dryRun)
		default:
			log.Warn("sync object not implemented", zap.String("object", object))
		}
	}
	return nil
}
