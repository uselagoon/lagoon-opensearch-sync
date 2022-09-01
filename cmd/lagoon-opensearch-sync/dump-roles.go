package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/uselagoon/lagoon-opensearch-sync/internal/opensearch"
	"go.uber.org/zap"
)

// DumpRolesCmd represents the `dump-roles` command.
type DumpRolesCmd struct {
	OpensearchUsername      string `kong:"default='admin',env='OPENSEARCH_ADMIN_USERNAME',help='Opensearch admin user'"`
	OpensearchPassword      string `kong:"required,env='OPENSEARCH_ADMIN_PASSWORD',help='Opensearch admin password'"`
	OpensearchBaseURL       string `kong:"required,env='OPENSEARCH_BASE_URL',help='Opensearch Base URL'"`
	OpensearchCACertificate string `kong:"required,env='OPENSEARCH_CA_CERTIFICATE',help='Opensearch CA Certificate'"`
	Raw                     bool   `kong:"help='Dump the raw JSON recevied from the backend service.'"`
}

// Run the dump-roles command.
func (cmd *DumpRolesCmd) Run(log *zap.Logger) error {
	// get main process context, which cancels on SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM)
	defer stop()
	// init the opensearch client
	o, err := opensearch.NewClient(ctx, log, cmd.OpensearchBaseURL,
		cmd.OpensearchUsername, cmd.OpensearchPassword, cmd.OpensearchCACertificate)
	if err != nil {
		return fmt.Errorf("couldn't init opensearch client: %v", err)
	}
	if cmd.Raw {
		data, err := o.RawRoles(ctx)
		fmt.Println(string(data))
		return err
	}
	// get the roles
	roles, err := o.Roles(ctx)
	if err != nil {
		return fmt.Errorf("couldn't get opensearch roles: %v", err)
	}
	// marshal and dump
	j, err := json.Marshal(roles)
	if err != nil {
		return fmt.Errorf("couldn't marshal roles: %v", err)
	}
	_, err = fmt.Println(string(j))
	return err
}
