package main

import (
	"github.com/alecthomas/kong"
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
	Version      VersionCmd      `kong:"cmd,help='Print version information'"`
	DumpProjects DumpProjectsCmd `kong:"cmd,help='Print Lagoon Projects JSON to standard out'"`
}

func main() {
	// parse CLI config
	cli := CLI{}
	kctx := kong.Parse(&cli,
		kong.UsageOnError(),
	)
	// execute CLI
	kctx.FatalIfErrorf(kctx.Run())
}
