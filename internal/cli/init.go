package cli

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/lebachhiep/claudex-cli/internal/auth"
	"github.com/lebachhiep/claudex-cli/internal/i18n"
	"github.com/lebachhiep/claudex-cli/internal/notification"
	"github.com/lebachhiep/claudex-cli/internal/projects"
	"github.com/lebachhiep/claudex-cli/internal/rules"
)

func newInitCmd(cliVersion string) *cobra.Command {
	var force bool
	var dir string

	cmd := &cobra.Command{
		Use:   "init [version]",
		Short: "Install or update rules in the current project",
		Long:  "Download and install rules. If already installed, updates to latest or specified version.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			authData, err := auth.EnsureAuth(apiClient, cfg)
			if err != nil {
				return err
			}

			// Determine target version from args
			targetVersion := ""
			if len(args) > 0 {
				targetVersion = strings.TrimPrefix(args[0], "v")
			}

			// Detect mode: init (fresh) or update (existing)
			lock, lockErr := rules.ReadLock(dir)
			currentVersion := ""
			mode := rules.ModeInit
			if lockErr == nil {
				currentVersion = lock.Version
				mode = rules.ModeUpdate
			}

			if force {
				currentVersion = "" // Force re-download
			}

			// Download
			result, err := rules.Download(apiClient, authData, cfg, currentVersion, targetVersion)
			if err != nil {
				return fmt.Errorf("download failed: %w", err)
			}

			green := color.New(color.FgGreen).SprintFunc()

			if result.UpToDate && !force {
				fmt.Printf("\n  %s %s\n\n", green("✓"), i18n.T("init.already_latest", lock.Version))
				return nil
			}

			// Install
			stats, err := rules.InstallWithMode(result, authData.Plan, dir, force || mode == rules.ModeInit, cliVersion, mode)
			if err != nil {
				return err
			}

			// Track project
			store, _ := projects.Load(cfg.ProjectsFile)
			if store == nil {
				store = projects.NewStore(cfg.ProjectsFile)
			}
			if mode == rules.ModeInit {
				_ = store.Register(dir, result.Version)
			} else {
				_ = store.UpdateVersion(dir, result.Version)
			}
			_ = store.Save()

			// Sync global config to project
			yellow := color.New(color.FgYellow).SprintFunc()
			if err := rules.SyncCodingLevel(cfg.ConfigFile, dir); err != nil {
				fmt.Printf("  %s %s\n", yellow("!"), i18n.T("init.coding_sync", err))
			}
			globalCfg, _ := notification.LoadGlobalConfig(cfg.ConfigFile)
			if globalCfg != nil && globalCfg.CodingLevel != -1 {
				if err := notification.SyncCodingLevelEnvToPath(globalCfg.CodingLevel, dir); err != nil {
					fmt.Printf("  %s %s\n", yellow("!"), i18n.T("init.coding_env_sync", err))
				}
			}
			if globalCfg != nil && globalCfg.HasNotification() {
				if err := notification.SyncToPath(globalCfg.Notification, globalCfg.EnableNotify, dir); err != nil {
					fmt.Printf("  %s %s\n", yellow("!"), i18n.T("init.notify_sync", err))
				}
			}
			if globalCfg != nil && globalCfg.HasContext7() {
				if err := notification.SyncContext7ToPath(globalCfg.Context7, dir); err != nil {
					fmt.Printf("  %s %s\n", yellow("!"), i18n.T("init.context7_sync", err))
				}
			}

			// Output
			if mode == rules.ModeUpdate && lockErr == nil {
				fmt.Printf("\n  %s %s\n", green("✓"), i18n.T("init.updated", lock.Version, result.Version))
			} else {
				fmt.Printf("\n  %s %s\n", green("✓"), i18n.T("init.installed", result.Version, result.SizeBytes/1024))
			}
			fmt.Printf("  %s\n", i18n.T("init.stats", stats.SkillCount, stats.AgentCount, stats.RuleCount))
			fmt.Printf("  %s %s\n\n", green("✓"), i18n.T("init.ready"))
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Force re-download and overwrite")
	cmd.Flags().StringVar(&dir, "dir", ".", "Target project directory")

	return cmd
}
