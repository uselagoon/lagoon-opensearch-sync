package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	"github.com/uselagoon/lagoon-opensearch-sync/internal/opensearch"
	"go.uber.org/zap"
)

// DumpIndexTemplatesCmd represents the `dump-index-templates` command.
type DumpIndexTemplatesCmd struct {
	OpensearchUsername      string        `kong:"default='admin',env='OPENSEARCH_ADMIN_USERNAME',help='Opensearch admin user'"`
	OpensearchPassword      string        `kong:"required,env='OPENSEARCH_ADMIN_PASSWORD',help='Opensearch admin password'"`
	OpensearchBaseURL       string        `kong:"required,env='OPENSEARCH_BASE_URL',help='Opensearch Base URL'"`
	OpensearchCACertificate string        `kong:"required,env='OPENSEARCH_CA_CERTIFICATE',help='Opensearch CA Certificate'"`
	HTTPClientTimeout       time.Duration `kong:"default='30s',env='HTTP_CLIENT_TIMEOUT',help='HTTP client timeout for API requests'"`
	Raw                     bool          `kong:"help='Dump the raw JSON recevied from the backend service.'"`
}

// Run the dump-index-templates command.
func (cmd *DumpIndexTemplatesCmd) Run(log *zap.Logger) error {
	// get main process context, which cancels on SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM)
	defer stop()
	// init the opensearch client
	o, err := opensearch.NewClient(log, cmd.OpensearchBaseURL,
		cmd.OpensearchUsername, cmd.OpensearchPassword, cmd.OpensearchCACertificate, cmd.HTTPClientTimeout)
	if err != nil {
		return fmt.Errorf("couldn't init opensearch client: %v", err)
	}
	if cmd.Raw {
		data, err := o.RawIndexTemplates(ctx)
		fmt.Println(string(data))
		return err
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
