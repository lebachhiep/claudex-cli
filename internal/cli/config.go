package cli

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/lebachhiep/claudex-cli/internal/notification"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage CLI configuration (notification, coding level, ...)",
		RunE: func(cmd *cobra.Command, args []string) error {
			var choice string
			err := huh.NewSelect[string]().
				Title("What would you like to configure?").
				Options(
					huh.NewOption("Coding Level — Set AI explanation verbosity", "coding-level"),
					huh.NewOption("Notification — Configure Telegram/Discord/Slack", "notification"),
				).
				Value(&choice).
				Run()
			if err != nil {
				return err
			}

			switch choice {
			case "coding-level":
				if err := runCodingLevelConfig(); err != nil {
					return err
				}
			case "notification":
				if err := runNotificationConfig(); err != nil {
					return err
				}
			}

			return promptSyncAfterConfig()
		},
	}

	cmd.AddCommand(
		newNotificationSubCmd(),
		newCodingLevelCmd(),
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

	if err := cfg.EnsureDataDir(); err != nil {
		return fmt.Errorf("create data dir: %w", err)
	}

	globalCfg, err := notification.LoadGlobalConfig(cfg.ConfigFile)
	if err != nil {
		globalCfg = &notification.GlobalConfig{CodingLevel: -1}
	}

	if globalCfg.HasNotification() {
		var overwrite bool
		err := huh.NewConfirm().
			Title(fmt.Sprintf("Notification already configured (%s). Overwrite?", globalCfg.Notification.Provider)).
			Value(&overwrite).
			Run()
		if err != nil {
			return err
		}
		if !overwrite {
			fmt.Printf("\n  %s Keeping existing config.\n\n", green("✓"))
			return nil
		}
	}

	var provider string
	err = huh.NewSelect[string]().
		Title("Select notification provider").
		Options(
			huh.NewOption("Telegram", notification.ProviderTelegram),
			huh.NewOption("Discord", notification.ProviderDiscord),
			huh.NewOption("Slack", notification.ProviderSlack),
		).
		Value(&provider).
		Run()
	if err != nil {
		return err
	}

	notiCfg := notification.NotificationConfig{Provider: provider}

	switch provider {
	case notification.ProviderTelegram:
		var botToken, chatID string
		err = huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Bot Token").
					Description("Get from @BotFather on Telegram").
					Value(&botToken).
					Validate(notEmpty("Bot Token")),
				huh.NewInput().
					Title("Chat ID").
					Description("Get from @userinfobot on Telegram").
					Value(&chatID).
					Validate(notEmpty("Chat ID")),
			),
		).Run()
		if err != nil {
			return err
		}
		notiCfg.Telegram.BotToken = strings.TrimSpace(botToken)
		notiCfg.Telegram.ChatID = strings.TrimSpace(chatID)

	case notification.ProviderDiscord:
		var webhookURL string
		err = huh.NewInput().
			Title("Discord Webhook URL").
			Description("Server Settings → Integrations → Webhooks → Copy URL").
			Value(&webhookURL).
			Validate(notEmpty("Webhook URL")).
			Run()
		if err != nil {
			return err
		}
		notiCfg.Discord.WebhookURL = strings.TrimSpace(webhookURL)

	case notification.ProviderSlack:
		var webhookURL string
		err = huh.NewInput().
			Title("Slack Webhook URL").
			Description("api.slack.com → Your Apps → Incoming Webhooks → Copy URL").
			Value(&webhookURL).
			Validate(notEmpty("Webhook URL")).
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
	fmt.Printf("\n  %s Notification config saved (%s)\n", green("✓"), provider)

	return nil
}

// notEmpty returns a huh validator that rejects blank input.
func notEmpty(field string) func(string) error {
	return func(s string) error {
		if strings.TrimSpace(s) == "" {
			return fmt.Errorf("%s cannot be empty", field)
		}
		return nil
	}
}
