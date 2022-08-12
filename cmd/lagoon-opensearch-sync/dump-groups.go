package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/uselagoon/lagoon-opensearch-sync/internal/keycloak"
)

// DumpGroupsCmd represents the `dump-groups` command.
type DumpGroupsCmd struct {
	KeycloakClientID     string `kong:"default='lagoon-opensearch-sync',env='KEYCLOAK_CLIENT_ID',help='Keycloak OAuth2 Client ID'"`
	KeycloakClientSecret string `kong:"required,env='KEYCLOAK_CLIENT_SECRET',help='Keycloak OAuth2 Client Secret'"`
	KeycloakBaseURL      string `kong:"required,env='KEYCLOAK_BASE_URL',help='Keycloak Base URL'"`
}

// Run the dump-groups command.
func (cmd *DumpGroupsCmd) Run() error {
	// get main process context, which cancels on SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM)
	defer stop()
	// init the keycloak client
	k, err := keycloak.NewClient(ctx, cmd.KeycloakBaseURL, cmd.KeycloakClientID,
		cmd.KeycloakClientSecret)
	if err != nil {
		return fmt.Errorf("couldn't init keycloak client: %v", err)
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
