package cli

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/claudex/claudex-cli/internal/auth"
	"github.com/claudex/claudex-cli/internal/rules"
)

func newVersionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "versions [current]",
		Short: "List available rules versions",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 && args[0] == "current" {
				return showCurrentVersion()
			}
			return listVersions()
		},
	}

	return cmd
}

func listVersions() error {
	authData, err := auth.EnsureAuth(apiClient, cfg)
	if err != nil {
		return err
	}

	versions, err := apiClient.GetVersions(authData.Token)
	if err != nil {
		return fmt.Errorf("fetch versions: %w", err)
	}

	if len(versions) == 0 {
		fmt.Println("\n  No versions available.\n")
		return nil
	}

	// Read current lock for marking installed version
	lock, _ := rules.ReadLock(".")
	currentVersion := ""
	if lock != nil {
		currentVersion = lock.Version
	}

	green := color.New(color.FgGreen).SprintFunc()

	fmt.Printf("\n  Available versions:\n\n")
	fmt.Printf("    %-14s %-10s %-12s %s\n", "VERSION", "SIZE", "DATE", "CHANGELOG")
	fmt.Printf("    %-14s %-10s %-12s %s\n", "-------", "----", "----", "---------")

	// Print oldest first, newest last
	for i := len(versions) - 1; i >= 0; i-- {
		v := versions[i]
		marker := "  "
		// Pre-pad version before coloring to avoid ANSI codes breaking alignment
		ver := fmt.Sprintf("%-14s", v.Version)
		if v.Version == currentVersion {
			marker = green("→") + " "
			ver = green(fmt.Sprintf("%-14s", v.Version))
		}

		size := fmt.Sprintf("%d KB", v.SizeBytes/1024)

		date := v.CreatedAt
		if len(date) >= 10 {
			date = date[:10]
		}

		changelog := v.Changelog
		if len(changelog) > 40 {
			changelog = changelog[:37] + "..."
		}

		fmt.Printf("  %s%s %-10s %-12s %s\n", marker, ver, size, date, changelog)
	}
	fmt.Println()
	return nil
}

func showCurrentVersion() error {
	lock, err := rules.ReadLock(".")
	if err != nil {
		return fmt.Errorf("no rules installed in this project. Run `claudex init` first")
	}

	green := color.New(color.FgGreen).SprintFunc()
	fmt.Printf("\n  %s Current version: %s\n", green("✓"), lock.Version)
	fmt.Printf("    Plan:      %s\n", lock.Plan)
	fmt.Printf("    Installed: %s\n", lock.InstalledAt.Format("2006-01-02 15:04"))
	fmt.Printf("    CLI:       %s\n\n", lock.CLIVersion)
	return nil
}
