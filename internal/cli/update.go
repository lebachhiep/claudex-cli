package cli

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/lebachhiep/claudex-cli/internal/auth"
	"github.com/lebachhiep/claudex-cli/internal/i18n"
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
				fmt.Printf("  %s %s\n", yellow("!"), i18n.T("update.removed_stale", cleaned))
			}

			projectList := store.List()
			if len(projectList) == 0 {
				return fmt.Errorf(i18n.T("update.no_projects"))
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
				return fmt.Errorf(i18n.T("update.no_lock"))
			}

			// Check for updates
			fmt.Printf("\n  %s\n", i18n.T("update.checking"))
			updateInfo, err := rules.CheckUpdate(apiClient, authData, currentVersion)
			if err != nil {
				return fmt.Errorf(i18n.T("update.check_err", err))
			}

			if !updateInfo.HasUpdate {
				fmt.Printf("  %s %s\n\n", green("✓"), i18n.T("update.already_latest", updateInfo.LatestVersion))
				return nil
			}

			fmt.Printf("  %s %s\n", green("✓"), i18n.T("update.new_version", updateInfo.CurrentVersion, updateInfo.LatestVersion))
			if updateInfo.Changelog != "" {
				fmt.Printf("  %s\n", i18n.T("update.changelog", updateInfo.Changelog))
			}

			// Download bundle (will be cached at ~/.claudex/cache/)
			fmt.Printf("  %s\n", i18n.T("update.downloading", updateInfo.LatestVersion))
			result, err := rules.Download(apiClient, authData, cfg, currentVersion, updateInfo.LatestVersion)
			if err != nil {
				return fmt.Errorf("download: %w", err)
			}
			fmt.Printf("  %s %s\n", green("✓"), i18n.T("update.downloaded", result.SizeBytes/1024))

			// Ask to update all projects
			fmt.Printf("\n  %s\n", i18n.T("update.projects_count", len(projectList)))
			for i, p := range projectList {
				lock, _ := rules.ReadLock(p.Path)
				ver := "unknown"
				if lock != nil {
					ver = lock.Version
				}
				fmt.Printf("  %s\n", i18n.T("update.project_item", i+1, p.Path, ver))
			}

			var choice string
			err = huh.NewSelect[string]().
				Title(i18n.T("update.confirm", len(projectList), updateInfo.LatestVersion)).
				Options(
					huh.NewOption(i18n.T("update.confirm_all"), "all"),
					huh.NewOption(i18n.T("common.skip"), "skip"),
				).
				Value(&choice).
				Run()
			if err != nil {
				return err
			}

			if choice == "skip" {
				fmt.Printf("\n  %s %s\n\n", yellow("!"), i18n.T("update.skipped"))
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
				if globalCfg != nil && globalCfg.CodingLevel != -1 {
					_ = notification.SyncCodingLevelEnvToPath(globalCfg.CodingLevel, p.Path)
				}
				if globalCfg != nil && globalCfg.HasNotification() {
					_ = notification.SyncToPath(globalCfg.Notification, globalCfg.EnableNotify, p.Path)
				}
				if globalCfg != nil && globalCfg.HasContext7() {
					_ = notification.SyncContext7ToPath(globalCfg.Context7, p.Path)
				}

				fmt.Printf("  %s %s\n",
					green("✓"), i18n.T("update.project_ok", p.Path, stats.SkillCount, stats.AgentCount, stats.RuleCount))
				updated++
			}

			if err := store.Save(); err != nil {
				fmt.Printf("  %s %s\n", yellow("!"), i18n.T("update.save_err", err))
			}

			if updated > 0 {
				fmt.Printf("\n  %s %s\n", green("✓"), i18n.T("update.done", updated, result.Version))
			}
			for _, e := range errors {
				fmt.Printf("  %s %s\n", color.RedString("✗"), e)
			}
			fmt.Println()

			return nil
		},
	}
}
