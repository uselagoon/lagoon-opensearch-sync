package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/go-sql-driver/mysql"
	"github.com/uselagoon/lagoon-opensearch-sync/internal/lagoondb"
)

// DumpProjectsCmd represents the `dump-projects` command.
type DumpProjectsCmd struct {
	APIDBAddress  string `kong:"required,env='API_DB_RO_ADDRESS,API_DB_ADDRESS',help='Lagoon API DB Address (host[:port])'"`
	APIDBDatabase string `kong:"default='infrastructure',env='API_DB_DATABASE',help='Lagoon API DB Database Name'"`
	APIDBPassword string `kong:"required,env='API_DB_RO_PASSWORD,API_DB_PASSWORD',help='Lagoon API DB Password'"`
	APIDBUsername string `kong:"default='api',env='API_DB_RO_USERNAME,API_DB_USERNAME',help='Lagoon API DB Username'"`
}

// Run the dump-projects command.
func (cmd *DumpProjectsCmd) Run() error {
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
	projects, err := l.Projects(ctx)
	if err != nil {
		return fmt.Errorf("couldn't get lagoondb projects: %v", err)
	}
	j, err := json.Marshal(projects)
	if err != nil {
		return fmt.Errorf("couldn't marshal projects: %v", err)
	}
	_, err = fmt.Println(string(j))
	return err
}
