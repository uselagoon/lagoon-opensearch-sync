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
	Roles(context.Context) (map[string]opensearch.Role, error)
	CreateRole(context.Context, string, *opensearch.Role) error
	DeleteRole(context.Context, string) error
}

// Sync will read the Lagoon state from the LagoonDBService and KeycloakService,
// and then configure the OpensearchService as required.
func Sync(ctx context.Context, log *zap.Logger, l LagoonDBService,
	k KeycloakService, o OpensearchService, dryRun bool) error {
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
	// get roles from Opensearch
	roles, err := o.Roles(ctx)
	if err != nil {
		return fmt.Errorf("couldn't get roles: %v", err)
	}
	syncRoles(ctx, log, groups, projectNames, roles, o, dryRun)
	return nil
}