package cli

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/lebachhiep/claudex-cli/internal/api"
	"github.com/lebachhiep/claudex-cli/internal/auth"
	"github.com/lebachhiep/claudex-cli/internal/i18n"
	"github.com/lebachhiep/claudex-cli/internal/notification"
	"github.com/lebachhiep/claudex-cli/internal/projects"
	"github.com/lebachhiep/claudex-cli/internal/rules"
)

func newVersionsCmd(cliVersion string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "versions [current]",
		Short: "List available rules versions and switch projects",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 && args[0] == "current" {
				return showCurrentVersion()
			}
			return listVersions(cliVersion)
		},
	}

	return cmd
}

func listVersions(cliVersion string) error {
	authData, err := auth.EnsureAuth(apiClient, cfg)
	if err != nil {
		return err
	}

	versions, err := apiClient.GetVersions(authData.Token)
	if err != nil {
		return fmt.Errorf("fetch versions: %w", err)
	}

	if len(versions) == 0 {
		fmt.Printf("\n  %s\n\n", i18n.T("versions.none"))
		return nil
	}

	// Read current lock for marking installed version
	lock, _ := rules.ReadLock(".")
	currentVersion := ""
	if lock != nil {
		currentVersion = lock.Version
	}

	printVersionsTable(versions, currentVersion)

	// Build select options (newest first for UX)
	var opts []huh.Option[string]
	for _, v := range versions {
		label := v.Version
		if v.Version == currentVersion {
			label += " " + i18n.T("versions.opt_current")
		}
		opts = append(opts, huh.NewOption(label, v.Version))
	}
	opts = append(opts, huh.NewOption(i18n.T("common.skip"), ""))

	var selected string
	err = huh.NewSelect[string]().
		Title(i18n.T("versions.select_title")).
		Options(opts...).
		Value(&selected).
		Run()
	if err != nil {
		return err
	}

	if selected == "" {
		return nil
	}

	return applyVersionToProjects(authData, selected, cliVersion)
}

func printVersionsTable(versions []api.VersionInfo, currentVersion string) {
	green := color.New(color.FgGreen).SprintFunc()

	fmt.Printf("\n  %s\n\n", i18n.T("versions.header"))
	fmt.Printf("    %-14s %-10s %-12s %s\n", i18n.T("versions.col_ver"), i18n.T("versions.col_size"), i18n.T("versions.col_date"), i18n.T("versions.col_log"))
	fmt.Printf("    %-14s %-10s %-12s %s\n", "-------", "----", "----", "---------")

	for i := len(versions) - 1; i >= 0; i-- {
		v := versions[i]
		marker := "  "
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
}

// applyVersionToProjects shows mismatched projects and lets user install the
// selected version into all/some/none of them.
func applyVersionToProjects(authData *auth.AuthData, targetVersion, cliVersion string) error {
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

	if err := cfg.EnsureDataDir(); err != nil {
		return fmt.Errorf("create data dir: %w", err)
	}

	store, err := projects.Load(cfg.ProjectsFile)
	if err != nil {
		return fmt.Errorf("load projects: %w", err)
	}
	if cleaned := store.CleanStale(); cleaned > 0 {
		_ = store.Save()
	}

	all := store.List()
	if len(all) == 0 {
		fmt.Printf("\n  %s %s\n\n", yellow("!"), i18n.T("update.no_projects"))
		return nil
	}

	type entry struct {
		proj    projects.Project
		version string
	}
	var mismatched []entry
	for _, p := range all {
		ver := ""
		if lock, lockErr := rules.ReadLock(p.Path); lockErr == nil {
			ver = lock.Version
		}
		if ver != targetVersion {
			mismatched = append(mismatched, entry{p, ver})
		}
	}

	if len(mismatched) == 0 {
		fmt.Printf("\n  %s %s\n\n", green("✓"), i18n.T("versions.all_on_target", targetVersion))
		return nil
	}

	fmt.Printf("\n  %s %s\n\n", cyan("■"), i18n.T("versions.mismatched_title", targetVersion))
	fmt.Printf("  %-4s %-50s %s\n", "#", i18n.T("projects.col_path"), i18n.T("versions.col_current"))
	fmt.Printf("  %s\n", strings.Repeat("─", 80))
	for i, e := range mismatched {
		ver := e.version
		if ver == "" {
			ver = i18n.T("versions.no_lock_short")
		}
		fmt.Printf("  %-4d %-50s %s\n", i+1, e.proj.Path, ver)
	}
	fmt.Println()

	var mode string
	err = huh.NewSelect[string]().
		Title(i18n.T("versions.confirm_title", len(mismatched), targetVersion)).
		Options(
			huh.NewOption(i18n.T("versions.confirm_all"), "all"),
			huh.NewOption(i18n.T("versions.confirm_pick"), "pick"),
			huh.NewOption(i18n.T("common.skip"), "skip"),
		).
		Value(&mode).
		Run()
	if err != nil {
		return err
	}

	if mode == "skip" {
		fmt.Printf("\n  %s %s\n\n", yellow("!"), i18n.T("update.skipped"))
		return nil
	}

	var targets []entry
	if mode == "all" {
		targets = mismatched
	} else {
		var pickOpts []huh.Option[int]
		for i, e := range mismatched {
			ver := e.version
			if ver == "" {
				ver = i18n.T("versions.no_lock_short")
			}
			pickOpts = append(pickOpts, huh.NewOption(fmt.Sprintf("%s (%s)", e.proj.Path, ver), i))
		}
		var picked []int
		err = huh.NewMultiSelect[int]().
			Title(i18n.T("versions.pick_title")).
			Options(pickOpts...).
			Value(&picked).
			Run()
		if err != nil {
			return err
		}
		if len(picked) == 0 {
			fmt.Printf("\n  %s %s\n\n", yellow("!"), i18n.T("update.skipped"))
			return nil
		}
		for _, idx := range picked {
			targets = append(targets, mismatched[idx])
		}
	}

	// Pick a sentinel currentVersion for the download request — any project's
	// current works since none equals targetVersion.
	currentForDownload := targets[0].version
	if currentForDownload == "" {
		currentForDownload = "0.0.0"
	}

	fmt.Printf("\n  %s\n", i18n.T("update.downloading", targetVersion))
	result, err := rules.Download(apiClient, authData, cfg, currentForDownload, targetVersion)
	if err != nil {
		return fmt.Errorf("download: %w", err)
	}
	if result.UpToDate || len(result.Bundle) == 0 {
		// Server said up-to-date relative to currentForDownload — refetch using
		// a definitely-older sentinel.
		result, err = rules.Download(apiClient, authData, cfg, "0.0.0", targetVersion)
		if err != nil {
			return fmt.Errorf("download: %w", err)
		}
	}
	fmt.Printf("  %s %s\n\n", green("✓"), i18n.T("update.downloaded", result.SizeBytes/1024))

	globalCfg, _ := notification.LoadGlobalConfig(cfg.ConfigFile)
	installed := 0
	var errors []string

	for _, e := range targets {
		plan := ""
		if lock, _ := rules.ReadLock(e.proj.Path); lock != nil {
			plan = lock.Plan
		}

		stats, installErr := rules.InstallWithMode(result, plan, e.proj.Path, false, cliVersion, rules.ModeUpdate)
		if installErr != nil {
			errors = append(errors, fmt.Sprintf("%s: %s", e.proj.Path, installErr))
			continue
		}

		_ = store.UpdateVersion(e.proj.Path, result.Version)

		_ = rules.SyncCodingLevel(cfg.ConfigFile, e.proj.Path)
		if globalCfg != nil && globalCfg.CodingLevel != -1 {
			_ = notification.SyncCodingLevelEnvToPath(globalCfg.CodingLevel, e.proj.Path)
		}
		if globalCfg != nil && globalCfg.HasNotification() {
			_ = notification.SyncToPath(globalCfg.Notification, globalCfg.EnableNotify, e.proj.Path)
		}
		if globalCfg != nil && globalCfg.HasContext7() {
			_ = notification.SyncContext7ToPath(globalCfg.Context7, e.proj.Path)
		}

		fmt.Printf("  %s %s\n", green("✓"), i18n.T("update.project_ok", e.proj.Path, stats.SkillCount, stats.AgentCount, stats.RuleCount))
		installed++
	}

	if err := store.Save(); err != nil {
		fmt.Printf("  %s %s\n", yellow("!"), i18n.T("update.save_err", err))
	}

	if installed > 0 {
		fmt.Printf("\n  %s %s\n", green("✓"), i18n.T("update.done", installed, result.Version))
	}
	for _, e := range errors {
		fmt.Printf("  %s %s\n", color.RedString("✗"), e)
	}
	fmt.Println()
	return nil
}

func showCurrentVersion() error {
	lock, err := rules.ReadLock(".")
	if err != nil {
		return fmt.Errorf(i18n.T("versions.no_rules"))
	}

	green := color.New(color.FgGreen).SprintFunc()
	fmt.Printf("\n  %s %s\n", green("✓"), i18n.T("versions.current", lock.Version))
	fmt.Printf("    %s\n", i18n.T("versions.plan", lock.Plan))
	fmt.Printf("    %s\n", i18n.T("versions.installed", lock.InstalledAt.Format("2006-01-02 15:04")))
	fmt.Printf("    %s\n\n", i18n.T("versions.cli", lock.CLIVersion))
	return nil
}
