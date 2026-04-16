package cli

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/huh"
	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/lebachhiep/claudex-cli/internal/i18n"
	"github.com/lebachhiep/claudex-cli/internal/notification"
)

func newCodingLevelCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "coding-level",
		Short: "Set coding experience level for tailored AI explanations",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCodingLevelConfig()
		},
	}
}

func runCodingLevelConfig() error {
	green := color.New(color.FgGreen).SprintFunc()

	if err := cfg.EnsureDataDir(); err != nil {
		return fmt.Errorf("create data dir: %w", err)
	}

	globalCfg, err := notification.LoadGlobalConfig(cfg.ConfigFile)
	if err != nil {
		globalCfg = &notification.GlobalConfig{CodingLevel: -1}
	}

	// Show current level
	fmt.Printf("\n  %s\n\n", i18n.T("coding.current", notification.CodingLevelName(globalCfg.CodingLevel)))

	// Interactive picker — build options from i18n keys
	var selected string
	var opts []huh.Option[string]
	for _, lvl := range []int{-1, 0, 1, 2, 3} {
		label := fmt.Sprintf("%2d: %s", lvl, i18n.T(fmt.Sprintf("coding.level_%d", lvl)))
		opts = append(opts, huh.NewOption(label, strconv.Itoa(lvl)))
	}
	err = huh.NewSelect[string]().
		Title(i18n.T("coding.title")).
		Options(opts...).
		Value(&selected).
		Run()
	if err != nil {
		return err
	}

	level, err := strconv.Atoi(selected)
	if err != nil {
		return fmt.Errorf("invalid level: %w", err)
	}
	globalCfg.CodingLevel = level

	if err := globalCfg.Save(cfg.ConfigFile); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	fmt.Printf("\n  %s %s\n", green("✓"), i18n.T("coding.set", notification.CodingLevelName(level)))
	fmt.Printf("  %s %s\n\n", green("✓"), i18n.T("coding.apply"))
	return nil
}
