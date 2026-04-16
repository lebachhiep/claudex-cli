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

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage CLI configuration (notification, coding level, ...)",
		RunE: func(cmd *cobra.Command, args []string) error {
			// First-run: ask language if not set yet
			globalCfg, _ := notification.LoadGlobalConfig(cfg.ConfigFile)
			if globalCfg != nil && globalCfg.Language == "" {
				if err := runLanguageConfig(); err != nil {
					return err
				}
			}

			var choice string
			err := huh.NewSelect[string]().
				Title(i18n.T("config.menu_title")).
				Options(
					huh.NewOption(i18n.T("config.language"), "language"),
					huh.NewOption(i18n.T("config.coding_level"), "coding-level"),
					huh.NewOption(i18n.T("config.notification"), "notification"),
					huh.NewOption(i18n.T("config.context7"), "context7"),
				).
				Value(&choice).
				Run()
			if err != nil {
				return err
			}

			switch choice {
			case "language":
				// Language is CLI-display only — no sync to projects needed
				return runLanguageConfig()
			case "coding-level":
				if err := runCodingLevelConfig(); err != nil {
					return err
				}
			case "notification":
				if err := runNotificationConfig(); err != nil {
					return err
				}
			case "context7":
				if err := runContext7Config(); err != nil {
					return err
				}
			}

			return promptSyncAfterConfig()
		},
	}

	cmd.AddCommand(
		newLanguageCmd(),
		newNotificationSubCmd(),
		newCodingLevelCmd(),
		newContext7Cmd(),
	)

	return cmd
}

func newNotificationSubCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "notification",
		Short: "Configure notification settings (Telegram, Discord, Slack)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runNotificationConfig()
		},
	}
}

func runNotificationConfig() error {
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

	if err := cfg.EnsureDataDir(); err != nil {
		return fmt.Errorf("create data dir: %w", err)
	}

	globalCfg, err := notification.LoadGlobalConfig(cfg.ConfigFile)
	if err != nil {
		globalCfg = &notification.GlobalConfig{CodingLevel: -1, EnableNotify: true}
	}

	// Show current status
	fmt.Printf("\n  %s %s\n", cyan("■"), i18n.T("notify.status_title"))
	if globalCfg.EnableNotify {
		fmt.Printf("    Enabled  : %s\n", green(i18n.T("notify.enabled_on")))
	} else {
		fmt.Printf("    Enabled  : %s\n", yellow(i18n.T("notify.enabled_off")))
	}
	if globalCfg.HasNotification() {
		fmt.Printf("    %s\n", i18n.T("notify.provider", globalCfg.Notification.Provider))
		printProviderDetails(globalCfg)
	} else {
		fmt.Printf("    %s\n", i18n.T("notify.provider", yellow(i18n.T("notify.not_configured"))))
	}
	fmt.Println()

	// If already configured, show action menu
	if globalCfg.HasNotification() {
		var action string
		opts := []huh.Option[string]{
			huh.NewOption(i18n.T("notify.toggle"), "toggle"),
			huh.NewOption(i18n.T("notify.reconfigure"), "reconfigure"),
			huh.NewOption(i18n.T("notify.skip_keep"), "skip"),
		}
		err := huh.NewSelect[string]().
			Title(i18n.T("notify.action_title")).
			Options(opts...).
			Value(&action).
			Run()
		if err != nil {
			return err
		}

		switch action {
		case "toggle":
			globalCfg.EnableNotify = !globalCfg.EnableNotify
			if err := globalCfg.Save(cfg.ConfigFile); err != nil {
				return fmt.Errorf("save config: %w", err)
			}
			if globalCfg.EnableNotify {
				fmt.Printf("\n  %s %s\n", green("✓"), i18n.T("notify.enabled_msg"))
			} else {
				fmt.Printf("\n  %s %s\n", green("✓"), i18n.T("notify.disabled_msg"))
			}
			return nil
		case "skip":
			return nil
		case "reconfigure":
			// Fall through to provider setup below
		}
	}

	// First-time setup: ask enable/disable before provider selection
	if !globalCfg.HasNotification() {
		green := color.New(color.FgGreen).SprintFunc()
		var enable bool
		if err := huh.NewConfirm().
			Title(i18n.T("notify.enable_prompt")).
			Value(&enable).
			Run(); err != nil {
			return err
		}
		if !enable {
			globalCfg.EnableNotify = false
			if err := globalCfg.Save(cfg.ConfigFile); err != nil {
				return fmt.Errorf("save config: %w", err)
			}
			fmt.Printf("\n  %s %s\n\n", green("✓"), i18n.T("notify.disabled_saved"))
			return nil
		}
		globalCfg.EnableNotify = true
	}

	// Provider setup flow
	return runProviderSetup(globalCfg)
}

// printProviderDetails shows the current provider credentials (masked).
func printProviderDetails(globalCfg *notification.GlobalConfig) {
	switch globalCfg.Notification.Provider {
	case notification.ProviderTelegram:
		fmt.Printf("    Bot Token: %s\n", maskValue(globalCfg.Notification.Telegram.BotToken))
		fmt.Printf("    Chat ID  : %s\n", globalCfg.Notification.Telegram.ChatID)
	case notification.ProviderDiscord:
		fmt.Printf("    Webhook  : %s\n", maskValue(globalCfg.Notification.Discord.WebhookURL))
	case notification.ProviderSlack:
		fmt.Printf("    Webhook  : %s\n", maskValue(globalCfg.Notification.Slack.WebhookURL))
	}
}

// maskValue shows first 10 and last 4 chars of a secret value.
func maskValue(s string) string {
	if len(s) <= 14 {
		return strings.Repeat("*", len(s))
	}
	return s[:10] + "..." + s[len(s)-4:]
}

// runProviderSetup handles the interactive provider selection and credential input.
func runProviderSetup(globalCfg *notification.GlobalConfig) error {
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	var provider string
	err := huh.NewSelect[string]().
		Title(i18n.T("notify.select_provider")).
		Options(
			huh.NewOption("Telegram", notification.ProviderTelegram),
			huh.NewOption("Discord", notification.ProviderDiscord),
			huh.NewOption("Slack", notification.ProviderSlack),
			huh.NewOption(i18n.T("notify.back"), "back"),
		).
		Value(&provider).
		Run()
	if err != nil {
		return err
	}
	if provider == "back" {
		fmt.Printf("\n  %s %s\n\n", yellow("!"), i18n.T("notify.cancelled"))
		return nil
	}

	notiCfg := notification.NotificationConfig{Provider: provider}

	switch provider {
	case notification.ProviderTelegram:
		botToken, chatID, err := runTelegramSetup()
		if err != nil {
			return err
		}
		if botToken == "" || chatID == "" {
			fmt.Printf("\n  %s %s\n\n", yellow("!"), i18n.T("notify.cancelled"))
			return nil
		}
		notiCfg.Telegram.BotToken = botToken
		notiCfg.Telegram.ChatID = chatID

	case notification.ProviderDiscord:
		var webhookURL string
		err = huh.NewInput().
			Title(i18n.T("notify.discord_url")).
			Description(i18n.T("notify.discord_desc")).
			Value(&webhookURL).
			Validate(notEmpty(i18n.T("notify.discord_url"))).
			Run()
		if err != nil {
			return err
		}
		notiCfg.Discord.WebhookURL = strings.TrimSpace(webhookURL)

	case notification.ProviderSlack:
		var webhookURL string
		err = huh.NewInput().
			Title(i18n.T("notify.slack_url")).
			Description(i18n.T("notify.slack_desc")).
			Value(&webhookURL).
			Validate(notEmpty(i18n.T("notify.slack_url"))).
			Run()
		if err != nil {
			return err
		}
		notiCfg.Slack.WebhookURL = strings.TrimSpace(webhookURL)
	}

	globalCfg.Notification = notiCfg
	if err := globalCfg.Save(cfg.ConfigFile); err != nil {
		return fmt.Errorf("save config: %w", err)
	}
	fmt.Printf("\n  %s %s\n", green("✓"), i18n.T("notify.saved", provider))

	return nil
}

// notEmpty returns a huh validator that rejects blank input.
func notEmpty(field string) func(string) error {
	return func(s string) error {
		if strings.TrimSpace(s) == "" {
			return fmt.Errorf(i18n.T("notify.field_required"), field)
		}
		return nil
	}
}
