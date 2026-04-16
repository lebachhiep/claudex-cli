package notification

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lebachhiep/claudex-cli/internal/projects"
)

const (
	markerStart = "# --- claudex-notification-start ---"
	markerEnd   = "# --- claudex-notification-end ---"

	context7MarkerStart = "# --- claudex-context7-start ---"
	context7MarkerEnd   = "# --- claudex-context7-end ---"

	codingLevelMarkerStart = "# --- claudex-coding-level-start ---"
	codingLevelMarkerEnd   = "# --- claudex-coding-level-end ---"

	languageMarkerStart = "# --- claudex-language-start ---"
	languageMarkerEnd   = "# --- claudex-language-end ---"
)

// sanitizeEnvValue strips newlines to prevent .env injection.
func sanitizeEnvValue(s string) string {
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, "\r", "")
	return s
}

// GenerateEnvBlock returns the notification .env section content.
// Includes ENABLE_NOTIFY flag at the top of the block.
func GenerateEnvBlock(cfg NotificationConfig, enableNotify bool) string {
	var lines []string
	lines = append(lines, markerStart)
	lines = append(lines, fmt.Sprintf("ENABLE_NOTIFY=%t", enableNotify))

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

// GenerateContext7Block returns the context7 .env section content.
func GenerateContext7Block(cfg Context7Config) string {
	var lines []string
	lines = append(lines, context7MarkerStart)
	lines = append(lines, "CONTEXT7_API_KEY="+sanitizeEnvValue(cfg.APIKey))
	lines = append(lines, context7MarkerEnd)
	return strings.Join(lines, "\n")
}

// GenerateCodingLevelBlock returns the coding level .env section content.
func GenerateCodingLevelBlock(level int) string {
	var lines []string
	lines = append(lines, codingLevelMarkerStart)
	lines = append(lines, fmt.Sprintf("CODING_LEVEL=%d", level))
	lines = append(lines, codingLevelMarkerEnd)
	return strings.Join(lines, "\n")
}

// GenerateLanguageBlock returns the language .env section content.
func GenerateLanguageBlock(lang string) string {
	var lines []string
	lines = append(lines, languageMarkerStart)
	lines = append(lines, "LANGUAGE="+sanitizeEnvValue(lang))
	lines = append(lines, languageMarkerEnd)
	return strings.Join(lines, "\n")
}

// SyncLanguageToPath writes the language .env block to a single project path.
func SyncLanguageToPath(lang, projectPath string) error {
	claudeDir := filepath.Join(projectPath, ".claude")
	if _, err := os.Stat(claudeDir); os.IsNotExist(err) {
		return fmt.Errorf("no .claude/ dir in %s", projectPath)
	}

	envPath := filepath.Join(claudeDir, ".env")
	return writeEnvBlock(envPath, GenerateLanguageBlock(lang))
}

// SyncCodingLevelEnvToPath writes the coding level .env block to a single project path.
func SyncCodingLevelEnvToPath(level int, projectPath string) error {
	claudeDir := filepath.Join(projectPath, ".claude")
	if _, err := os.Stat(claudeDir); os.IsNotExist(err) {
		return fmt.Errorf("no .claude/ dir in %s", projectPath)
	}

	envPath := filepath.Join(claudeDir, ".env")
	return writeEnvBlock(envPath, GenerateCodingLevelBlock(level))
}

// SyncContext7ToPath writes the context7 .env block to a single project path.
func SyncContext7ToPath(cfg Context7Config, projectPath string) error {
	claudeDir := filepath.Join(projectPath, ".claude")
	if _, err := os.Stat(claudeDir); os.IsNotExist(err) {
		return fmt.Errorf("no .claude/ dir in %s", projectPath)
	}

	envPath := filepath.Join(claudeDir, ".env")
	return writeEnvBlock(envPath, GenerateContext7Block(cfg))
}

// SyncResult holds the outcome of syncing .env to projects.
type SyncResult struct {
	Synced  int
	Skipped int
	Errors  []string
}

// SyncToPath writes the notification .env block to a single project path.
// Returns nil on success, error on failure. Skips if .claude/ dir doesn't exist.
func SyncToPath(cfg NotificationConfig, enableNotify bool, projectPath string) error {
	claudeDir := filepath.Join(projectPath, ".claude")
	if _, err := os.Stat(claudeDir); os.IsNotExist(err) {
		return fmt.Errorf("no .claude/ dir in %s", projectPath)
	}

	envPath := filepath.Join(claudeDir, ".env")
	return writeEnvBlock(envPath, GenerateEnvBlock(cfg, enableNotify))
}

// SyncToProjects writes the notification .env block to all tracked projects.
func SyncToProjects(cfg NotificationConfig, enableNotify bool, store *projects.Store) SyncResult {
	result := SyncResult{}
	block := GenerateEnvBlock(cfg, enableNotify)

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

// writeEnvBlock inserts or replaces a managed section in a .env file.
// The block must start and end with marker comments. The function detects
// the start marker (first line of block) to find and replace existing sections.
func writeEnvBlock(envPath, block string) error {
	var existing string
	if data, err := os.ReadFile(envPath); err == nil {
		existing = string(data)
	}

	// Extract start and end markers from the block itself
	lines := strings.SplitN(block, "\n", 2)
	blockStartMarker := lines[0]
	// Find end marker: last non-empty line
	blockLines := strings.Split(block, "\n")
	blockEndMarker := blockLines[len(blockLines)-1]

	// Replace existing managed section or append
	startIdx := strings.Index(existing, blockStartMarker)
	if startIdx != -1 {
		endIdx := strings.Index(existing, blockEndMarker)
		if endIdx != -1 {
			endIdx += len(blockEndMarker)
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
