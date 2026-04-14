package rules

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/claudex/claudex-cli/internal/notification"
)

// codingLevelRe matches the codingLevel key-value pair in JSON.
var codingLevelRe = regexp.MustCompile(`("codingLevel"\s*:\s*)(-?\d+)`)

// SyncCodingLevel reads codingLevel from global config and patches the project's .claude-config.json.
// Skips silently if global codingLevel is -1 (disabled) or project config doesn't exist.
// Uses regex replacement to preserve JSON key order and formatting.
func SyncCodingLevel(globalCfgPath, projectDir string) error {
	globalCfg, err := notification.LoadGlobalConfig(globalCfgPath)
	if err != nil || globalCfg.CodingLevel == -1 {
		return nil // disabled or not set — skip
	}

	configPath := filepath.Join(projectDir, ".claude", ".claude-config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // no project config — skip
		}
		return fmt.Errorf("read project config: %w", err)
	}

	content := string(data)

	// Check if codingLevel already matches — skip unnecessary write
	match := codingLevelRe.FindStringSubmatch(content)
	if match == nil {
		return nil // codingLevel key not found — skip
	}

	existing, _ := strconv.Atoi(match[2])
	if existing == globalCfg.CodingLevel {
		return nil // already correct — no write needed
	}

	// Replace in-place, preserving formatting
	content = codingLevelRe.ReplaceAllString(content, fmt.Sprintf("${1}%d", globalCfg.CodingLevel))

	return os.WriteFile(configPath, []byte(content), 0644)
}
