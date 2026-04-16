package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/lebachhiep/claudex-cli/internal/notification"
	"github.com/lebachhiep/claudex-cli/internal/projects"
	"github.com/lebachhiep/claudex-cli/internal/rules"
)

func newProjectsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "projects",
		Short: "List all tracked projects and sync config",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProjectsList()
		},
	}
}

func runProjectsList() error {
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

	// Clean stale projects first
	cleaned := store.CleanStale()
	if cleaned > 0 {
		_ = store.Save()
	}

	projectList := store.List()
	if len(projectList) == 0 {
		fmt.Printf("\n  %s No tracked projects. Run 'claudex init' in a project first.\n\n", yellow("!"))
		return nil
	}

	// Detect current project
	cwd, _ := filepath.Abs(".")
	cwd = filepath.Clean(cwd)
	currentIdx := -1
	for i, p := range projectList {
		if filepath.Clean(p.Path) == cwd {
			currentIdx = i
			break
		}
	}

	// Print table
	fmt.Printf("\n  %s Tracked Projects\n\n", cyan("■"))
	fmt.Printf("  %-4s %-45s %-10s %-12s %s\n", "#", "Path", "Version", "Installed", "Updated")
	fmt.Printf("  %s\n", strings.Repeat("─", 95))

	for i, p := range projectList {
		label := p.Path
		if i == currentIdx {
			label += " " + green("(current)")
		}
		fmt.Printf("  %-4d %-45s %-10s %-12s %s\n",
			i+1,
			label,
			p.Version,
			p.InstalledAt.Format("2006-01-02"),
			p.UpdatedAt.Format("2006-01-02"),
		)
	}

	if cleaned > 0 {
		fmt.Printf("\n  %s Removed %d stale project(s)\n", yellow("!"), cleaned)
	}
	fmt.Printf("\n  %s %d project(s) tracked\n", green("✓"), len(projectList))

	// Show global config summary
	globalCfg, err := notification.LoadGlobalConfig(cfg.ConfigFile)
	if err != nil {
		globalCfg = &notification.GlobalConfig{CodingLevel: -1}
	}

	fmt.Printf("\n  %s Global Config\n", cyan("■"))
	fmt.Printf("    Coding Level : %s\n", notification.CodingLevelName(globalCfg.CodingLevel))
	if globalCfg.HasNotification() {
		fmt.Printf("    Notification : %s (configured)\n", globalCfg.Notification.Provider)
	} else {
		fmt.Printf("    Notification : %s\n", yellow("not configured"))
	}
	fmt.Println()

	// Ask to sync config
	return askSyncProjects(globalCfg, store, currentIdx)
}

// promptSyncAfterConfig loads projects and prompts sync after config changes.
func promptSyncAfterConfig() error {
	store, err := projects.Load(cfg.ProjectsFile)
	if err != nil || len(store.List()) == 0 {
		return nil // no projects to sync
	}

	cleaned := store.CleanStale()
	if cleaned > 0 {
		_ = store.Save()
	}
	if len(store.List()) == 0 {
		return nil
	}

	globalCfg, err := notification.LoadGlobalConfig(cfg.ConfigFile)
	if err != nil {
		return nil
	}

	cwd, _ := filepath.Abs(".")
	cwd = filepath.Clean(cwd)
	currentIdx := -1
	for i, p := range store.List() {
		if filepath.Clean(p.Path) == cwd {
			currentIdx = i
			break
		}
	}

	fmt.Printf("\n  %d project(s) tracked.\n", len(store.List()))
	return askSyncProjects(globalCfg, store, currentIdx)
}

// askSyncProjects prompts user to sync global config to tracked projects.
func askSyncProjects(globalCfg *notification.GlobalConfig, store *projects.Store, currentIdx int) error {
	green := color.New(color.FgGreen).SprintFunc()

	var opts []huh.Option[string]
	if currentIdx >= 0 {
		opts = append(opts, huh.NewOption("Update current project only", "current"))
	}
	opts = append(opts,
		huh.NewOption("Update all projects", "all"),
		huh.NewOption("Skip", "skip"),
	)

	var choice string
	err := huh.NewSelect[string]().
		Title("Sync global config to projects?").
		Options(opts...).
		Value(&choice).
		Run()
	if err != nil {
		return err
	}

	if choice == "skip" {
		return nil
	}

	projectList := store.List()
	var targets []projects.Project

	switch choice {
	case "current":
		targets = []projects.Project{projectList[currentIdx]}
	case "all":
		targets = projectList
	}

	synced := 0
	var errors []string
	for _, p := range targets {
		hasError := false
		// Sync notification .env
		if globalCfg.HasNotification() {
			if err := notification.SyncToPath(globalCfg.Notification, p.Path); err != nil {
				errors = append(errors, fmt.Sprintf("%s: notification: %s", p.Path, err))
				hasError = true
			}
		}
		// Sync coding level
		if globalCfg.CodingLevel != -1 {
			if err := rules.SyncCodingLevel(cfg.ConfigFile, p.Path); err != nil {
				errors = append(errors, fmt.Sprintf("%s: coding-level: %s", p.Path, err))
				hasError = true
			}
		}
		if !hasError {
			synced++
		}
	}

	if synced > 0 {
		fmt.Printf("\n  %s Synced config to %d project(s)\n", green("✓"), synced)
	}
	for _, e := range errors {
		fmt.Printf("  %s %s\n", color.RedString("✗"), e)
	}
	fmt.Println()
	return nil
}
