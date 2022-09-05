package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/uselagoon/lagoon-opensearch-sync/internal/dashboards"
	"github.com/uselagoon/lagoon-opensearch-sync/internal/keycloak"
	"github.com/uselagoon/lagoon-opensearch-sync/internal/lagoondb"
	"github.com/uselagoon/lagoon-opensearch-sync/internal/opensearch"
	"github.com/uselagoon/lagoon-opensearch-sync/internal/sync"
	"go.uber.org/zap"
)

// WatchCmd represents the `watch` command.
type WatchCmd struct {
	SyncCmd
	Period time.Duration `kong:"default='8m',help='Period between synchronisation checks'"`
}

// Run the watch command.
func (cmd *WatchCmd) Run(log *zap.Logger) error {
	// get main process context, which cancels on SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM)
	defer stop()
	// init lagoon DB client
	dbConf := mysql.NewConfig()
	dbConf.Addr = cmd.APIDBAddress
	dbConf.DBName = cmd.APIDBDatabase
	dbConf.Net = "tcp"
	dbConf.Passwd = cmd.APIDBPassword
	dbConf.User = cmd.APIDBUsername
	l, err := lagoondb.NewClient(ctx, dbConf.FormatDSN())
	if err != nil {
		return fmt.Errorf("couldn't init lagoon DBClient: %v", err)
	}
	// init the keycloak client
	k, err := keycloak.NewClientCredentialsClient(ctx, cmd.KeycloakBaseURL, cmd.KeycloakClientID,
		cmd.KeycloakClientSecret)
	if err != nil {
		return fmt.Errorf("couldn't init keycloak client: %v", err)
	}
	// init the opensearch client
	o, err := opensearch.NewClient(ctx, log, cmd.OpensearchBaseURL,
		cmd.OpensearchUsername, cmd.OpensearchPassword, cmd.OpensearchCACertificate)
	if err != nil {
		return fmt.Errorf("couldn't init opensearch client: %v", err)
	}
	// init the opensearch dashboards client
	d, err := dashboards.NewClient(ctx, cmd.OpensearchDashboardsBaseURL,
		cmd.OpensearchUsername, cmd.OpensearchPassword)
	if err != nil {
		return fmt.Errorf("couldn't init opensearch client: %v", err)
	}
	// run sync immediately
	err = sync.Sync(ctx, log, l, k, o, d, cmd.DryRun, cmd.Objects)
	if err != nil {
		return err
	}
	// continue running in a loop
	tick := time.NewTicker(cmd.Period)
	for range tick.C {
		err = sync.Sync(ctx, log, l, k, o, d, cmd.DryRun, cmd.Objects)
		if err != nil {
			return err
		}
	}
	return nil
}
