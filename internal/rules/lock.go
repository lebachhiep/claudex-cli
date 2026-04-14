package rules

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/claudex/claudex-cli/internal/config"
)

// LockData represents the .claudex.lock file in a project directory.
type LockData struct {
	Version     string    `json:"version"`
	Plan        string    `json:"plan"`
	InstalledAt time.Time `json:"installed_at"`
	Checksum    string    `json:"checksum"`
	CLIVersion  string    `json:"cli_version"`
}

// ReadLock reads .claudex.lock from the project directory.
func ReadLock(projectDir string) (*LockData, error) {
	lockPath := filepath.Join(projectDir, config.LockFileName)
	data, err := os.ReadFile(lockPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no .claudex.lock found. Run `claudex init` first")
		}
		return nil, fmt.Errorf("read lock file: %w", err)
	}

	var lock LockData
	if err := json.Unmarshal(data, &lock); err != nil {
		return nil, fmt.Errorf("parse lock file: %w", err)
	}
	return &lock, nil
}

// WriteLock writes .claudex.lock to the project directory.
func WriteLock(projectDir string, lock *LockData) error {
	lockPath := filepath.Join(projectDir, config.LockFileName)
	data, err := json.MarshalIndent(lock, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal lock: %w", err)
	}
	if err := os.WriteFile(lockPath, data, 0644); err != nil {
		return fmt.Errorf("write lock file: %w", err)
	}
	return nil
}
