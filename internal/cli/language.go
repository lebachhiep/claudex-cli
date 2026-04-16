package cli

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/lebachhiep/claudex-cli/internal/i18n"
	"github.com/lebachhiep/claudex-cli/internal/notification"
)

func newLanguageCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "language",
		Short: "Select CLI language (English / Tiếng Việt)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLanguageConfig()
		},
	}
}

func runLanguageConfig() error {
	green := color.New(color.FgGreen).SprintFunc()

	if err := cfg.EnsureDataDir(); err != nil {
		return fmt.Errorf("create data dir: %w", err)
	}

	globalCfg, err := notification.LoadGlobalConfig(cfg.ConfigFile)
	if err != nil {
		globalCfg = &notification.GlobalConfig{CodingLevel: -1, EnableNotify: true}
	}

	// Show current language
	if globalCfg.Language != "" {
		langName := i18n.T("lang.english")
		if globalCfg.Language == "vi" {
			langName = i18n.T("lang.vietnamese")
		}
		fmt.Printf("\n  %s\n\n", fmt.Sprintf(i18n.T("lang.current"), langName))
	}

	var selected string
	err = huh.NewSelect[string]().
		Title(i18n.T("lang.title")).
		Options(
			huh.NewOption("English", "en"),
			huh.NewOption("Tiếng Việt", "vi"),
		).
		Value(&selected).
		Run()
	if err != nil {
		return err
	}

	globalCfg.Language = selected
	if err := globalCfg.Save(cfg.ConfigFile); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	// Apply immediately
	i18n.Init(selected)

	langName := i18n.T("lang.english")
	if selected == "vi" {
		langName = i18n.T("lang.vietnamese")
	}
	fmt.Printf("\n  %s %s\n\n", green("✓"), fmt.Sprintf(i18n.T("lang.saved"), langName))

	return nil
}
