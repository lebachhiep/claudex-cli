package notification

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/claudex/claudex-cli/internal/projects"
)

const (
	markerStart = "# --- claudex-notification-start ---"
	markerEnd   = "# --- claudex-notification-end ---"
)

// sanitizeEnvValue strips newlines to prevent .env injection.
func sanitizeEnvValue(s string) string {
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, "\r", "")
	return s
}

// GenerateEnvBlock returns the notification .env section content.
func GenerateEnvBlock(cfg NotificationConfig) string {
	var lines []string
	lines = append(lines, markerStart)

	switch cfg.Provider {
	case ProviderTelegram:
		lines = append(lines, "TELEGRAM_BOT_TOKEN="+sanitizeEnvValue(cfg.Telegram.BotToken))
		lines = append(lines, "TELEGRAM_CHAT_ID="+sanitizeEnvValue(cfg.Telegram.ChatID))
		lines = append(lines, "# DISCORD_WEBHOOK_URL=")
		lines = append(lines, "# SLACK_WEBHOOK_URL=")
	case ProviderDiscord:
		lines = append(lines, "# TELEGRAM_BOT_TOKEN=")
		lines = append(lines, "# TELEGRAM_CHAT_ID=")
		lines = append(lines, "DISCORD_WEBHOOK_URL="+sanitizeEnvValue(cfg.Discord.WebhookURL))
		lines = append(lines, "# SLACK_WEBHOOK_URL=")
	case ProviderSlack:
		lines = append(lines, "# TELEGRAM_BOT_TOKEN=")
		lines = append(lines, "# TELEGRAM_CHAT_ID=")
		lines = append(lines, "# DISCORD_WEBHOOK_URL=")
		lines = append(lines, "SLACK_WEBHOOK_URL="+sanitizeEnvValue(cfg.Slack.WebhookURL))
	}

	lines = append(lines, markerEnd)
	return strings.Join(lines, "\n")
}

// SyncResult holds the outcome of syncing .env to projects.
type SyncResult struct {
	Synced  int
	Skipped int
	Errors  []string
}

// SyncToPath writes the notification .env block to a single project path.
// Returns nil on success, error on failure. Skips if .claude/ dir doesn't exist.
func SyncToPath(cfg NotificationConfig, projectPath string) error {
	claudeDir := filepath.Join(projectPath, ".claude")
	if _, err := os.Stat(claudeDir); os.IsNotExist(err) {
		return fmt.Errorf("no .claude/ dir in %s", projectPath)
	}

	envPath := filepath.Join(claudeDir, ".env")
	return writeEnvBlock(envPath, GenerateEnvBlock(cfg))
}

// SyncToProjects writes the notification .env block to all tracked projects.
func SyncToProjects(cfg NotificationConfig, store *projects.Store) SyncResult {
	result := SyncResult{}
	block := GenerateEnvBlock(cfg)

	for _, p := range store.List() {
		claudeDir := filepath.Join(p.Path, ".claude")
		if _, err := os.Stat(claudeDir); os.IsNotExist(err) {
			result.Skipped++
			continue
		}

		envPath := filepath.Join(claudeDir, ".env")
		if err := writeEnvBlock(envPath, block); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %s", p.Path, err))
			continue
		}
		result.Synced++
	}

	return result
}

// writeEnvBlock inserts or replaces the managed notification section in a .env file.
func writeEnvBlock(envPath, block string) error {
	var existing string
	if data, err := os.ReadFile(envPath); err == nil {
		existing = string(data)
	}

	// Replace existing managed section or append
	startIdx := strings.Index(existing, markerStart)
	if startIdx != -1 {
		endIdx := strings.Index(existing, markerEnd)
		if endIdx != -1 {
			endIdx += len(markerEnd)
			existing = existing[:startIdx] + block + existing[endIdx:]
		} else {
			// End marker missing — replace from start marker to end of file
			existing = existing[:startIdx] + block + "\n"
		}
	} else {
		// Append with newline separator
		if existing != "" && !strings.HasSuffix(existing, "\n") {
			existing += "\n"
		}
		if existing != "" {
			existing += "\n"
		}
		existing += block + "\n"
	}

	return os.WriteFile(envPath, []byte(existing), 0600)
}
