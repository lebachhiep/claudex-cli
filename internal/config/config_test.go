package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// TestDefaultConfigWithoutEnvVars verifies default config with default values.
func TestDefaultConfigWithoutEnvVars(t *testing.T) {
	// Save original env vars
	origServer := os.Getenv("CLAUDEX_SERVER")
	origDataDir := os.Getenv("CLAUDEX_DATA_DIR")
	defer func() {
		os.Setenv("CLAUDEX_SERVER", origServer)
		os.Setenv("CLAUDEX_DATA_DIR", origDataDir)
	}()

	// Clear env vars to use defaults
	os.Unsetenv("CLAUDEX_SERVER")
	os.Unsetenv("CLAUDEX_DATA_DIR")

	cfg, err := DefaultConfig()
	if err != nil {
		t.Fatalf("DefaultConfig failed: %v", err)
	}

	if cfg.ServerURL != DefaultServerURL {
		t.Errorf("ServerURL mismatch: expected %s, got %s", DefaultServerURL, cfg.ServerURL)
	}

	if cfg.DataDir == "" {
		t.Error("DataDir should not be empty")
	}

	// Verify paths are resolved correctly
	if !filepath.IsAbs(cfg.DataDir) {
		t.Errorf("DataDir should be absolute path: %s", cfg.DataDir)
	}

	expectedAuthFile := filepath.Join(cfg.DataDir, AuthFileName)
	if cfg.AuthFile != expectedAuthFile {
		t.Errorf("AuthFile mismatch: expected %s, got %s", expectedAuthFile, cfg.AuthFile)
	}

	expectedCacheDir := filepath.Join(cfg.DataDir, CacheDirName)
	if cfg.CacheDir != expectedCacheDir {
		t.Errorf("CacheDir mismatch: expected %s, got %s", expectedCacheDir, cfg.CacheDir)
	}
}

// TestDefaultConfigWithServerURLEnv verifies CLAUDEX_SERVER env override.
func TestDefaultConfigWithServerURLEnv(t *testing.T) {
	origServer := os.Getenv("CLAUDEX_SERVER")
	defer os.Setenv("CLAUDEX_SERVER", origServer)

	customServer := "https://custom.server.com"
	os.Setenv("CLAUDEX_SERVER", customServer)

	cfg, err := DefaultConfig()
	if err != nil {
		t.Fatalf("DefaultConfig failed: %v", err)
	}

	if cfg.ServerURL != customServer {
		t.Errorf("ServerURL mismatch: expected %s, got %s", customServer, cfg.ServerURL)
	}
}

// TestDefaultConfigWithDataDirEnv verifies CLAUDEX_DATA_DIR env override.
func TestDefaultConfigWithDataDirEnv(t *testing.T) {
	origDataDir := os.Getenv("CLAUDEX_DATA_DIR")
	defer os.Setenv("CLAUDEX_DATA_DIR", origDataDir)

	customDataDir := t.TempDir()
	os.Setenv("CLAUDEX_DATA_DIR", customDataDir)

	cfg, err := DefaultConfig()
	if err != nil {
		t.Fatalf("DefaultConfig failed: %v", err)
	}

	if cfg.DataDir != customDataDir {
		t.Errorf("DataDir mismatch: expected %s, got %s", customDataDir, cfg.DataDir)
	}

	expectedAuthFile := filepath.Join(customDataDir, AuthFileName)
	if cfg.AuthFile != expectedAuthFile {
		t.Errorf("AuthFile mismatch: expected %s, got %s", expectedAuthFile, cfg.AuthFile)
	}
}

// TestEnsureDataDirCreatesDirectories verifies EnsureDataDir creates necessary directories.
func TestEnsureDataDirCreatesDirectories(t *testing.T) {
	tempDir := t.TempDir()
	dataDir := filepath.Join(tempDir, ".claudex")
	cacheDir := filepath.Join(dataDir, CacheDirName)

	cfg := &Config{
		ServerURL: DefaultServerURL,
		DataDir:   dataDir,
		CacheDir:  cacheDir,
		AuthFile:  filepath.Join(dataDir, AuthFileName),
	}

	// Directories should not exist yet
	if _, err := os.Stat(dataDir); err == nil || !os.IsNotExist(err) {
		t.Fatal("DataDir should not exist before EnsureDataDir")
	}

	// Create directories
	if err := cfg.EnsureDataDir(); err != nil {
		t.Fatalf("EnsureDataDir failed: %v", err)
	}

	// Verify directories exist
	if _, err := os.Stat(dataDir); err != nil {
		t.Errorf("DataDir should exist after EnsureDataDir: %v", err)
	}

	if _, err := os.Stat(cacheDir); err != nil {
		t.Errorf("CacheDir should exist after EnsureDataDir: %v", err)
	}

	// Verify permissions (skip on Windows — doesn't support Unix permissions)
	if runtime.GOOS != "windows" {
		stat, _ := os.Stat(dataDir)
		if stat.Mode().Perm()&0077 != 0 {
			t.Errorf("DataDir permissions too open: %o", stat.Mode().Perm())
		}
	}
}

// TestEnsureDataDirIdempotent verifies EnsureDataDir is idempotent.
func TestEnsureDataDirIdempotent(t *testing.T) {
	tempDir := t.TempDir()
	dataDir := filepath.Join(tempDir, ".claudex")
	cacheDir := filepath.Join(dataDir, CacheDirName)

	cfg := &Config{
		ServerURL: DefaultServerURL,
		DataDir:   dataDir,
		CacheDir:  cacheDir,
		AuthFile:  filepath.Join(dataDir, AuthFileName),
	}

	// First call
	if err := cfg.EnsureDataDir(); err != nil {
		t.Fatalf("first EnsureDataDir failed: %v", err)
	}

	// Second call should succeed without error
	if err := cfg.EnsureDataDir(); err != nil {
		t.Fatalf("second EnsureDataDir failed: %v", err)
	}

	// Verify directories still exist
	if _, err := os.Stat(dataDir); err != nil {
		t.Errorf("DataDir should still exist: %v", err)
	}

	if _, err := os.Stat(cacheDir); err != nil {
		t.Errorf("CacheDir should still exist: %v", err)
	}
}

// TestEnvOrDefaultWithSetEnv verifies envOrDefault returns env value when set.
func TestEnvOrDefaultWithSetEnv(t *testing.T) {
	origVal := os.Getenv("TEST_VAR_EXISTS")
	defer os.Setenv("TEST_VAR_EXISTS", origVal)

	testValue := "custom-value"
	os.Setenv("TEST_VAR_EXISTS", testValue)

	result := envOrDefault("TEST_VAR_EXISTS", "default")
	if result != testValue {
		t.Errorf("envOrDefault should return env value: expected %s, got %s", testValue, result)
	}
}

// TestEnvOrDefaultWithUnsetEnv verifies envOrDefault returns fallback when unset.
func TestEnvOrDefaultWithUnsetEnv(t *testing.T) {
	origVal := os.Getenv("TEST_VAR_UNSET")
	defer os.Setenv("TEST_VAR_UNSET", origVal)

	os.Unsetenv("TEST_VAR_UNSET")

	fallback := "default-value"
	result := envOrDefault("TEST_VAR_UNSET", fallback)
	if result != fallback {
		t.Errorf("envOrDefault should return fallback: expected %s, got %s", fallback, result)
	}
}

// TestEnvOrDefaultWithEmptyEnv verifies envOrDefault returns fallback when empty.
func TestEnvOrDefaultWithEmptyEnv(t *testing.T) {
	origVal := os.Getenv("TEST_VAR_EMPTY")
	defer os.Setenv("TEST_VAR_EMPTY", origVal)

	os.Setenv("TEST_VAR_EMPTY", "")

	fallback := "default-value"
	result := envOrDefault("TEST_VAR_EMPTY", fallback)
	if result != fallback {
		t.Errorf("envOrDefault should return fallback for empty env: expected %s, got %s", fallback, result)
	}
}

// TestEnvOrDefaultWithEmptyFallback verifies envOrDefault works with empty fallback.
func TestEnvOrDefaultWithEmptyFallback(t *testing.T) {
	origVal := os.Getenv("TEST_VAR_EMPTY_FB")
	defer os.Setenv("TEST_VAR_EMPTY_FB", origVal)

	os.Unsetenv("TEST_VAR_EMPTY_FB")

	result := envOrDefault("TEST_VAR_EMPTY_FB", "")
	if result != "" {
		t.Errorf("envOrDefault should return empty fallback: got %s", result)
	}
}

// TestConfigPathConstruction verifies all path fields are properly constructed.
func TestConfigPathConstruction(t *testing.T) {
	tempDir := t.TempDir()

	cfg := &Config{
		ServerURL: "https://test.dev",
		DataDir:   tempDir,
		AuthFile:  filepath.Join(tempDir, AuthFileName),
		CacheDir:  filepath.Join(tempDir, CacheDirName),
	}

	// Verify AuthFile contains correct filename
	if !filepath.HasPrefix(cfg.AuthFile, cfg.DataDir) {
		t.Errorf("AuthFile should be within DataDir: %s not in %s", cfg.AuthFile, cfg.DataDir)
	}

	if filepath.Base(cfg.AuthFile) != AuthFileName {
		t.Errorf("AuthFile should end with %s: got %s", AuthFileName, filepath.Base(cfg.AuthFile))
	}

	// Verify CacheDir is within DataDir
	if !filepath.HasPrefix(cfg.CacheDir, cfg.DataDir) {
		t.Errorf("CacheDir should be within DataDir: %s not in %s", cfg.CacheDir, cfg.DataDir)
	}

	if filepath.Base(cfg.CacheDir) != CacheDirName {
		t.Errorf("CacheDir should end with %s: got %s", CacheDirName, filepath.Base(cfg.CacheDir))
	}
}
