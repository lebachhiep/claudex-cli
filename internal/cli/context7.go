package cli

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/lebachhiep/claudex-cli/internal/i18n"
	"github.com/lebachhiep/claudex-cli/internal/notification"
)

func newContext7Cmd() *cobra.Command {
	return &cobra.Command{
		Use:   "context7",
		Short: "Configure Context7 API key for docs-seeker",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runContext7Config()
		},
	}
}

func runContext7Config() error {
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	if err := cfg.EnsureDataDir(); err != nil {
		return fmt.Errorf("create data dir: %w", err)
	}

	globalCfg, err := notification.LoadGlobalConfig(cfg.ConfigFile)
	if err != nil {
		globalCfg = &notification.GlobalConfig{CodingLevel: -1, EnableNotify: true}
	}

	// Show current status
	if globalCfg.HasContext7() {
		masked := maskAPIKey(globalCfg.Context7.APIKey)
		fmt.Printf("\n  %s\n\n", i18n.T("context7.current", green(masked)))
	} else {
		fmt.Printf("\n  %s\n\n", i18n.T("context7.current", yellow(i18n.T("context7.not_set"))))
	}

	var apiKey string
	err = huh.NewInput().
		Title(i18n.T("context7.title")).
		Description(i18n.T("context7.desc")).
		Value(&apiKey).
		Run()
	if err != nil {
		return err
	}

	globalCfg.Context7.APIKey = strings.TrimSpace(apiKey)

	if err := globalCfg.Save(cfg.ConfigFile); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	if globalCfg.HasContext7() {
		fmt.Printf("\n  %s %s\n", green("✓"), i18n.T("context7.saved"))
	} else {
		fmt.Printf("\n  %s %s\n", green("✓"), i18n.T("context7.cleared"))
	}

	return promptSyncAfterConfig()
}

// maskAPIKey shows first 8 and last 4 chars, masking the rest.
func maskAPIKey(key string) string {
	if len(key) <= 12 {
		return strings.Repeat("*", len(key))
	}
	return key[:8] + strings.Repeat("*", len(key)-12) + key[len(key)-4:]
}
