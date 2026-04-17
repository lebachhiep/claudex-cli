// Package cli defines all cobra commands for the claudex CLI.
package cli

import (
	"os"
	"runtime"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/lebachhiep/claudex-cli/internal/api"
	"github.com/lebachhiep/claudex-cli/internal/config"
	"github.com/lebachhiep/claudex-cli/internal/i18n"
	"github.com/lebachhiep/claudex-cli/internal/notification"
)

// Shared state initialized in PersistentPreRun, available to all subcommands.
var (
	cfg       *config.Config
	apiClient *api.Client
)

// NewRootCmd creates the root cobra command with all subcommands.
func NewRootCmd(version, commit, date string) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "claudex",
		Short: "ClaudeX — Claude Code skills distribution CLI",
		Long:  "Download, install, and manage Claude Code skills/agents/rules from ClaudeX.",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if shouldDisableColor() {
				color.NoColor = true
			}

			var err error
			cfg, err = config.DefaultConfig()
			if err != nil {
				return err
			}

			apiClient = api.NewClient(cfg.ServerURL)

			// Load language preference for i18n
			if globalCfg, err := notification.LoadGlobalConfig(cfg.ConfigFile); err == nil && globalCfg.Language != "" {
				i18n.Init(globalCfg.Language)
			}

			return nil
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Register all subcommands
	rootCmd.AddCommand(
		newLoginCmd(),
		newInitCmd(version),
		newUpdateCmd(version),
		newVersionsCmd(version),
		newStatusCmd(),
		newLogoutCmd(),
		newConfigCmd(),
		newProjectsCmd(),
		newVersionCmd(version, commit, date),
	)

	return rootCmd
}

// shouldDisableColor detects legacy PowerShell on Windows (the blue console)
// which doesn't support ANSI escape codes properly.
func shouldDisableColor() bool {
	if runtime.GOOS != "windows" {
		return false
	}

	// Modern terminals support ANSI — keep colors
	if os.Getenv("WT_SESSION") != "" { // Windows Terminal
		return false
	}
	if os.Getenv("TERM_PROGRAM") != "" { // VS Code, iTerm, etc.
		return false
	}
	if os.Getenv("ConEmuANSI") == "ON" { // ConEmu/Cmder
		return false
	}

	// Detect legacy PowerShell (5.x, the blue window)
	// PSModulePath contains "WindowsPowerShell" but NOT "PowerShell\7"
	psModulePath := strings.ToLower(os.Getenv("PSModulePath"))
	if strings.Contains(psModulePath, "windowspowershell") && !strings.Contains(psModulePath, "powershell\\7") {
		return true
	}

	return false
}
