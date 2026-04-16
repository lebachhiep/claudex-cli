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

// Context7Config holds Context7 API credentials.
type Context7Config struct {
	APIKey string `json:"api_key"`
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
	Context7     Context7Config     `json:"context7"`
	EnableNotify bool               `json:"enable_notify"`
	Language     string             `json:"language"`
}

// CodingLevelName returns the display name for a coding level.
func CodingLevelName(level int) string {
	names := map[int]string{
		-1: "Disabled",
		0:  "Intern — explain all terms, avoid jargon",
		1:  "Junior — explain patterns first time, suggest best practices",
		2:  "Mid — only explain complex logic or unclear trade-offs",
		3:  "Senior+ — terse, focus scalability/security/business impact",
	}
	if name, ok := names[level]; ok {
		return name
	}
	return "Unknown"
}

// LoadGlobalConfig reads config.json. Returns empty config if not found.
// EnableNotify defaults to true for backward compatibility.
func LoadGlobalConfig(path string) (*GlobalConfig, error) {
	c := &GlobalConfig{CodingLevel: -1, EnableNotify: true}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return c, nil
		}
		return nil, fmt.Errorf("read config: %w", err)
	}

	// Default EnableNotify to true before unmarshal — if JSON has the field it overrides.
	c.EnableNotify = true
	if err := json.Unmarshal(data, c); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return c, nil
}

// HasContext7 returns true if Context7 API key is configured.
func (c *GlobalConfig) HasContext7() bool {
	return c.Context7.APIKey != ""
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
