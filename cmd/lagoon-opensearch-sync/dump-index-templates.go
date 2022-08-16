package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/uselagoon/lagoon-opensearch-sync/internal/opensearch"
)

// DumpIndexTemplatesCmd represents the `dump-index-templates` command.
type DumpIndexTemplatesCmd struct {
	OpensearchUsername      string `kong:"default='admin',env='OPENSEARCH_ADMIN_USERNAME',help='Opensearch admin user'"`
	OpensearchPassword      string `kong:"required,env='OPENSEARCH_ADMIN_PASSWORD',help='Opensearch admin password'"`
	OpensearchBaseURL       string `kong:"required,env='OPENSEARCH_BASE_URL',help='Opensearch Base URL'"`
	OpensearchCACertificate string `kong:"required,env='OPENSEARCH_CA_CERTIFICATE',help='Opensearch CA Certificate'"`
}

// Run the dump-index-templates command.
func (cmd *DumpIndexTemplatesCmd) Run() error {
	// get main process context, which cancels on SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM)
	defer stop()
	// init the opensearch client
	o, err := opensearch.NewClient(ctx, cmd.OpensearchBaseURL,
		cmd.OpensearchUsername, cmd.OpensearchPassword, cmd.OpensearchCACertificate)
	if err != nil {
		return fmt.Errorf("couldn't init opensearch client: %v", err)
	}
	// get the index templates
	it, err := o.IndexTemplates(ctx)
	if err != nil {
		return fmt.Errorf("couldn't get opensearch index templates: %v", err)
	}
	// marshal and dump
	data, err := json.Marshal(it)
	if err != nil {
		return fmt.Errorf("couldn't marshal index templates: %v", err)
	}
	_, err = fmt.Println(string(data))
	return err
}
