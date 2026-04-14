package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newVersionCmd(version, commit, date string) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show CLI version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("claudex v%s (build: %s, date: %s)\n", version, commit, date)
		},
	}
}
