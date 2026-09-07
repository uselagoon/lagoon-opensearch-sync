package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	"github.com/uselagoon/lagoon-opensearch-sync/internal/keycloak"
)

// DumpGroupsCmd represents the `dump-groups` command.
type DumpGroupsCmd struct {
	KeycloakClientID       string        `kong:"default='lagoon-opensearch-sync',env='KEYCLOAK_CLIENT_ID',help='Keycloak OAuth2 Client ID'"`
	KeycloakClientSecret   string        `kong:"required,env='KEYCLOAK_CLIENT_SECRET',help='Keycloak OAuth2 Client Secret'"`
	KeycloakBaseURL        string        `kong:"required,env='KEYCLOAK_BASE_URL',help='Keycloak Base URL'"`
	KeycloakClientTimeout  time.Duration `kong:"default='30s',env='KEYCLOAK_CLIENT_TIMEOUT',help='Keycloak HTTP client request timeout'"`
	KeycloakGroupsPageSize uint          `kong:"default='100',env='KEYCLOAK_GROUPS_PAGE_SIZE',help='Number of groups to fetch per page from the Keycloak Admin API'"`
	Raw                    bool          `kong:"help='Dump the raw JSON recevied from the backend service.'"`
	RawFirst               uint          `kong:"help='Set the first field of the raw groups query, which controls the pagination offset.'"`
	RawMax                 uint          `kong:"default='100',help='Set the max field of the raw groups query, which controls the number of results returned.'"`
}

// Run the dump-groups command.
func (cmd *DumpGroupsCmd) Run() error {
	// get main process context, which cancels on SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM)
	defer stop()
	// init the keycloak client
	k, err := keycloak.NewClientCredentialsClient(ctx, cmd.KeycloakBaseURL,
		cmd.KeycloakClientID, cmd.KeycloakClientSecret,
		cmd.KeycloakClientTimeout, cmd.KeycloakGroupsPageSize)
	if err != nil {
		return fmt.Errorf("couldn't init keycloak client: %v", err)
	}
	if cmd.Raw {
		data, err := k.RawGroups(ctx, cmd.RawFirst, cmd.RawMax)
		fmt.Println(string(data))
		return err
	}
	groups, err := k.Groups(ctx)
	if err != nil {
		return fmt.Errorf("couldn't get keycloak groups: %v", err)
	}
	j, err := json.Marshal(groups)
	if err != nil {
		return fmt.Errorf("couldn't marshal groups: %v", err)
	}
	_, err = fmt.Println(string(j))
	return err
}
