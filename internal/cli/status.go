package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/lebachhiep/claudex-cli/internal/auth"
	"github.com/lebachhiep/claudex-cli/internal/rules"
)

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show license, machine, and project rules status",
		RunE: func(cmd *cobra.Command, args []string) error {
			authData, err := auth.EnsureAuth(apiClient, cfg)
			if err != nil {
				return err
			}

			green := color.New(color.FgGreen).SprintFunc()
			cyan := color.New(color.FgCyan).SprintFunc()

			fmt.Printf("\n  License:    %s (active)\n", cyan(strings.ToUpper(authData.Plan)))
			fmt.Printf("  Validity:   Lifetime\n")

			// Project status
			cwd, _ := os.Getwd()
			absDir, _ := filepath.Abs(cwd)

			lock, err := rules.ReadLock(".")
			if err != nil {
				fmt.Printf("\n  Project:    %s\n", absDir)
				fmt.Printf("  Rules:      not installed (run `claudex init`)\n\n")
				return nil
			}

			// Check for updates
			updateInfo, _ := rules.CheckUpdate(apiClient, authData, lock.Version)

			fmt.Printf("\n  Project:    %s\n", absDir)
			fmt.Printf("  Rules:      %s (installed %s)\n", lock.Version, lock.InstalledAt.Format("2006-01-02"))

			if updateInfo != nil && updateInfo.HasUpdate {
				fmt.Printf("  Latest:     %s (run `claudex update`)\n\n", updateInfo.LatestVersion)
			} else {
				fmt.Printf("  Latest:     %s %s up to date\n\n", lock.Version, green("✓"))
			}

			return nil
		},
	}
}
