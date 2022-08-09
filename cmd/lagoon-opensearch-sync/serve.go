package main

import (
	"fmt"

	"github.com/uselagoon/lagoon-opensearch-sync/internal/server"
)

// ServeCmd represents the `serve` command.
type ServeCmd struct{}

// Run the serve command.
func (*ServeCmd) Run() error {
	fmt.Println(server.Serve())
	return nil
}
