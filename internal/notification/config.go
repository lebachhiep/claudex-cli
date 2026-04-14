// Package notification manages notification provider config and .env sync.
package notification

import (
	"encoding/json"
	"fmt"
	"os"
)

// Provider name constants.
const (
	ProviderTelegram = "telegram"
	ProviderDiscord  = "discord"
	ProviderSlack    = "slack"
)

// TelegramConfig holds Telegram bot credentials.
type TelegramConfig struct {
	BotToken string `json:"bot_token"`
	ChatID   string `json:"chat_id"`
}

// DiscordConfig holds Discord webhook URL.
type DiscordConfig struct {
	WebhookURL string `json:"webhook_url"`
}

// SlackConfig holds Slack webhook URL.
type SlackConfig struct {
	WebhookURL string `json:"webhook_url"`
}

// NotificationConfig holds all provider configs and active provider.
type NotificationConfig struct {
	Provider string         `json:"provider"`
	Telegram TelegramConfig `json:"telegram"`
	Discord  DiscordConfig  `json:"discord"`
	Slack    SlackConfig    `json:"slack"`
}

// GlobalConfig is the top-level persistent config file (~/.claudex/config.json).
type GlobalConfig struct {
	Notification NotificationConfig `json:"notification"`
	CodingLevel  int                `json:"coding_level"`
}

// CodingLevelName returns the display name for a coding level.
func CodingLevelName(level int) string {
	names := map[int]string{
		-1: "Disabled",
		0:  "ELI5 — explain like I'm 5",
		1:  "Junior — explain WHY, common mistakes",
		2:  "Mid-Level — design patterns, system thinking",
		3:  "Senior — trade-offs, architecture",
		4:  "Tech Lead — risk, business impact",
		5:  "God Mode — terse, code only",
	}
	if name, ok := names[level]; ok {
		return name
	}
	return "Unknown"
}

// LoadGlobalConfig reads config.json. Returns empty config if not found.
func LoadGlobalConfig(path string) (*GlobalConfig, error) {
	c := &GlobalConfig{CodingLevel: -1}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return c, nil
		}
		return nil, fmt.Errorf("read config: %w", err)
	}

	if err := json.Unmarshal(data, c); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return c, nil
}

// Save writes config to disk with 0600 permissions.
func (c *GlobalConfig) Save(path string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	return os.WriteFile(path, data, 0600)
}

// HasNotification returns true if any provider is configured.
func (c *GlobalConfig) HasNotification() bool {
	switch c.Notification.Provider {
	case ProviderTelegram:
		return c.Notification.Telegram.BotToken != "" && c.Notification.Telegram.ChatID != ""
	case ProviderDiscord:
		return c.Notification.Discord.WebhookURL != ""
	case ProviderSlack:
		return c.Notification.Slack.WebhookURL != ""
	}
	return false
}
