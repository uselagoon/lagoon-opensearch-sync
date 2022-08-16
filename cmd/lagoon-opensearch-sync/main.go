package main

import (
	"github.com/alecthomas/kong"
	"go.uber.org/zap"
)

var (
	commit      string
	date        string
	goVersion   string
	projectName string
	version     string
)

// CLI represents the command-line interface.
type CLI struct {
	Debug            bool                `kong:"env='DEBUG',help='Enable debug logging'"`
	Version          VersionCmd          `kong:"cmd,help='Print version information'"`
	DumpProjects     DumpProjectsCmd     `kong:"cmd,help='Print Lagoon Projects JSON to standard out'"`
	DumpGroups       DumpGroupsCmd       `kong:"cmd,help='Print Keycloak Groups JSON to standard out'"`
	DumpRoles        DumpRolesCmd        `kong:"cmd,help='Print Opensearch Roles JSON to standard out'"`
	DumpRolesmapping DumpRolesmappingCmd `kong:"cmd,help='Print Opensearch Rolesmapping JSON to standard out'"`
}

func main() {
	// parse CLI config
	cli := CLI{}
	kctx := kong.Parse(&cli,
		kong.UsageOnError(),
	)
	// init logger
	var log *zap.Logger
	var err error
	if cli.Debug {
		log, err = zap.NewDevelopment(zap.AddStacktrace(zap.ErrorLevel))
	} else {
		log, err = zap.NewProduction()
	}
	if err != nil {
		panic(err)
	}
	defer log.Sync() //nolint:errcheck
	// execute CLI
	kctx.FatalIfErrorf(kctx.Run(log))
}
