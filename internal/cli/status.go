package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/lebachhiep/claudex-cli/internal/auth"
	"github.com/lebachhiep/claudex-cli/internal/i18n"
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

			fmt.Printf("\n  %s\n", i18n.T("status.license", cyan(strings.ToUpper(authData.Plan))))
			fmt.Printf("  %s\n", i18n.T("status.validity"))

			// Project status
			cwd, _ := os.Getwd()
			absDir, _ := filepath.Abs(cwd)

			lock, err := rules.ReadLock(".")
			if err != nil {
				fmt.Printf("\n  %s\n", i18n.T("status.project", absDir))
				fmt.Printf("  %s\n\n", i18n.T("status.rules_none"))
				return nil
			}

			// Check for updates
			updateInfo, _ := rules.CheckUpdate(apiClient, authData, lock.Version)

			fmt.Printf("\n  %s\n", i18n.T("status.project", absDir))
			fmt.Printf("  %s\n", i18n.T("status.rules", lock.Version, lock.InstalledAt.Format("2006-01-02")))

			if updateInfo != nil && updateInfo.HasUpdate {
				fmt.Printf("  %s\n\n", i18n.T("status.latest_update", updateInfo.LatestVersion))
			} else {
				fmt.Printf("  %s\n\n", i18n.T("status.latest_ok", lock.Version, green("✓")))
			}

			return nil
		},
	}
}
