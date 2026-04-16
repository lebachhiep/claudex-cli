package cli

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/lebachhiep/claudex-cli/internal/auth"
	"github.com/lebachhiep/claudex-cli/internal/notification"
	"github.com/lebachhiep/claudex-cli/internal/projects"
	"github.com/lebachhiep/claudex-cli/internal/rules"
)

func newUpdateCmd(cliVersion string) *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Check for new rules version and update all tracked projects",
		RunE: func(cmd *cobra.Command, args []string) error {
			green := color.New(color.FgGreen).SprintFunc()
			yellow := color.New(color.FgYellow).SprintFunc()

			authData, err := auth.EnsureAuth(apiClient, cfg)
			if err != nil {
				return err
			}

			// Load tracked projects
			store, err := projects.Load(cfg.ProjectsFile)
			if err != nil {
				return fmt.Errorf("load projects: %w", err)
			}
			cleaned := store.CleanStale()
			if cleaned > 0 {
				_ = store.Save()
				fmt.Printf("  %s Removed %d stale project(s)\n", yellow("!"), cleaned)
			}

			projectList := store.List()
			if len(projectList) == 0 {
				return fmt.Errorf("no tracked projects. Run 'claudex init' in a project first")
			}

			// Find oldest version across all projects to ensure all get updated
			var currentVersion string
			for _, p := range projectList {
				if lock, lockErr := rules.ReadLock(p.Path); lockErr == nil {
					if currentVersion == "" || lock.Version < currentVersion {
						currentVersion = lock.Version
					}
				}
			}
			if currentVersion == "" {
				return fmt.Errorf("no project has a lock file — cannot determine current version. Run 'claudex init' first")
			}

			// Check for updates
			fmt.Printf("\n  Checking for updates...\n")
			updateInfo, err := rules.CheckUpdate(apiClient, authData, currentVersion)
			if err != nil {
				return fmt.Errorf("check update: %w", err)
			}

			if !updateInfo.HasUpdate {
				fmt.Printf("  %s Already on latest version (%s)\n\n", green("✓"), updateInfo.LatestVersion)
				return nil
			}

			fmt.Printf("  %s New version available: %s → %s\n", green("✓"), updateInfo.CurrentVersion, updateInfo.LatestVersion)
			if updateInfo.Changelog != "" {
				fmt.Printf("  Changelog: %s\n", updateInfo.Changelog)
			}

			// Download bundle (will be cached at ~/.claudex/cache/)
			fmt.Printf("  Downloading %s...\n", updateInfo.LatestVersion)
			result, err := rules.Download(apiClient, authData, cfg, currentVersion, updateInfo.LatestVersion)
			if err != nil {
				return fmt.Errorf("download: %w", err)
			}
			fmt.Printf("  %s Downloaded (%d KB)\n", green("✓"), result.SizeBytes/1024)

			// Ask to update all projects
			fmt.Printf("\n  %d project(s) tracked:\n", len(projectList))
			for i, p := range projectList {
				lock, _ := rules.ReadLock(p.Path)
				ver := "unknown"
				if lock != nil {
					ver = lock.Version
				}
				fmt.Printf("    %d. %s (v%s)\n", i+1, p.Path, ver)
			}

			var choice string
			err = huh.NewSelect[string]().
				Title(fmt.Sprintf("Update all %d project(s) to %s?", len(projectList), updateInfo.LatestVersion)).
				Options(
					huh.NewOption("Update all projects", "all"),
					huh.NewOption("Skip", "skip"),
				).
				Value(&choice).
				Run()
			if err != nil {
				return err
			}

			if choice == "skip" {
				fmt.Printf("\n  %s Skipped. Bundle cached — run 'claudex init' in a project to install.\n\n", yellow("!"))
				return nil
			}

			// Install to all projects
			globalCfg, _ := notification.LoadGlobalConfig(cfg.ConfigFile)
			updated := 0
			var errors []string

			for _, p := range projectList {
				lock, _ := rules.ReadLock(p.Path)
				plan := ""
				if lock != nil {
					plan = lock.Plan
				}

				stats, installErr := rules.InstallWithMode(result, plan, p.Path, false, cliVersion, rules.ModeUpdate)
				if installErr != nil {
					errors = append(errors, fmt.Sprintf("%s: %s", p.Path, installErr))
					continue
				}

				// Update project tracking
				_ = store.UpdateVersion(p.Path, result.Version)

				// Sync config
				_ = rules.SyncCodingLevel(cfg.ConfigFile, p.Path)
				if globalCfg != nil && globalCfg.HasNotification() {
					_ = notification.SyncToPath(globalCfg.Notification, p.Path)
				}

				fmt.Printf("  %s %s — %d skills, %d agents, %d rules\n",
					green("✓"), p.Path, stats.SkillCount, stats.AgentCount, stats.RuleCount)
				updated++
			}

			if err := store.Save(); err != nil {
				fmt.Printf("  %s Failed to save project tracking: %s\n", yellow("!"), err)
			}

			if updated > 0 {
				fmt.Printf("\n  %s Updated %d project(s) to %s\n", green("✓"), updated, result.Version)
			}
			for _, e := range errors {
				fmt.Printf("  %s %s\n", color.RedString("✗"), e)
			}
			fmt.Println()

			return nil
		},
	}
}
