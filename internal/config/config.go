// Package config manages CLI configuration: server URL, paths, defaults.
package config

import (
	"os"
	"path/filepath"
)

const (
	DefaultServerURL = "https://api.claudex.info"
	AuthFileName     = "auth.json"
	ConfigFileName   = "config.json"
	ProjectsFileName = "projects.json"
	CacheDirName     = "cache"
	LockFileName     = ".claudex.lock"
)

// Config holds all CLI configuration values.
type Config struct {
	ServerURL    string
	DataDir      string // ~/.claudex/
	AuthFile     string // ~/.claudex/auth.json
	ConfigFile   string // ~/.claudex/config.json
	ProjectsFile string // ~/.claudex/projects.json
	CacheDir     string // ~/.claudex/cache/
}

// DefaultConfig creates config with default values.
// Respects env overrides: CLAUDEX_SERVER, CLAUDEX_DATA_DIR.
func DefaultConfig() (*Config, error) {
	serverURL := envOrDefault("CLAUDEX_SERVER", DefaultServerURL)
	dataDir, err := resolveDataDir()
	if err != nil {
		return nil, err
	}

	return &Config{
		ServerURL:    serverURL,
		DataDir:      dataDir,
		AuthFile:     filepath.Join(dataDir, AuthFileName),
		ConfigFile:   filepath.Join(dataDir, ConfigFileName),
		ProjectsFile: filepath.Join(dataDir, ProjectsFileName),
		CacheDir:     filepath.Join(dataDir, CacheDirName),
	}, nil
}

// EnsureDataDir creates ~/.claudex/ and cache/ if they don't exist.
func (c *Config) EnsureDataDir() error {
	if err := os.MkdirAll(c.DataDir, 0700); err != nil {
		return err
	}
	return os.MkdirAll(c.CacheDir, 0700)
}

// resolveDataDir returns ~/.claudex/, respecting CLAUDEX_DATA_DIR override.
func resolveDataDir() (string, error) {
	if dir := os.Getenv("CLAUDEX_DATA_DIR"); dir != "" {
		return dir, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".claudex"), nil
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
