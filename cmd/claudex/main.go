package main

import (
	"fmt"
	"os"

	"github.com/claudex/claudex-cli/internal/cli"
)

// Injected via ldflags at build time.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	rootCmd := cli.NewRootCmd(version, commit, date)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}
