package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/go-sql-driver/mysql"
	"github.com/uselagoon/lagoon-opensearch-sync/internal/dashboards"
	"github.com/uselagoon/lagoon-opensearch-sync/internal/keycloak"
	"github.com/uselagoon/lagoon-opensearch-sync/internal/lagoondb"
	"github.com/uselagoon/lagoon-opensearch-sync/internal/opensearch"
	"github.com/uselagoon/lagoon-opensearch-sync/internal/sync"
	"go.uber.org/zap"
)

// SyncCmd represents the `sync` command.
type SyncCmd struct {
	DryRun  bool     `kong:"env='DRY_RUN',help='Print actions that will be taken but do not persist any changes to Opensearch'"`
	Objects []string `kong:"enum='tenants,roles,rolesmapping,indexpatterns,indextemplates',default='tenants,roles,rolesmapping,indexpatterns,indextemplates',help='Opensearch objects which will be synchronized'"`
	// lagoon DB client fields
	APIDBAddress  string `kong:"required,env='API_DB_ADDRESS',help='Lagoon API DB Address (host[:port])'"`
	APIDBDatabase string `kong:"default='infrastructure',env='API_DB_DATABASE',help='Lagoon API DB Database Name'"`
	APIDBPassword string `kong:"required,env='API_DB_PASSWORD',help='Lagoon API DB Password'"`
	APIDBUsername string `kong:"default='api',env='API_DB_USERNAME',help='Lagoon API DB Username'"`
	// keycloak client fields
	KeycloakClientID     string `kong:"default='lagoon-opensearch-sync',env='KEYCLOAK_CLIENT_ID',help='Keycloak OAuth2 Client ID'"`
	KeycloakClientSecret string `kong:"required,env='KEYCLOAK_CLIENT_SECRET',help='Keycloak OAuth2 Client Secret'"`
	KeycloakBaseURL      string `kong:"required,env='KEYCLOAK_BASE_URL',help='Keycloak Base URL'"`
	// opensearch client fields
	OpensearchUsername      string `kong:"default='admin',env='OPENSEARCH_ADMIN_USERNAME',help='Opensearch admin user'"`
	OpensearchPassword      string `kong:"required,env='OPENSEARCH_ADMIN_PASSWORD',help='Opensearch admin password'"`
	OpensearchBaseURL       string `kong:"required,env='OPENSEARCH_BASE_URL',help='Opensearch Base URL'"`
	OpensearchCACertificate string `kong:"required,env='OPENSEARCH_CA_CERTIFICATE',help='Opensearch CA Certificate'"`
	// dashboards client fields
	OpensearchDashboardsBaseURL string `kong:"required,env='OPENSEARCH_DASHBOARDS_BASE_URL',help='Opensearch Dashboards Base URL'"`
}

// Run the sync command.
func (cmd *SyncCmd) Run(log *zap.Logger) error {
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
	// run the sync
	return sync.Sync(ctx, log, l, k, o, d, cmd.DryRun, cmd.Objects)
}
