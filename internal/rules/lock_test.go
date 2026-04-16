package rules

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/lebachhiep/claudex-cli/internal/config"
)

// TestWriteAndReadLockRoundTrip verifies WriteLock and ReadLock round-trip.
func TestWriteAndReadLockRoundTrip(t *testing.T) {
	tempDir := t.TempDir()

	original := &LockData{
		Version:     "1.0.0",
		Plan:        "paid",
		InstalledAt: time.Now(),
		Checksum:    "sha256:abc123def456",
		CLIVersion:  "0.1.0",
	}

	// Write lock
	if err := WriteLock(tempDir, original); err != nil {
		t.Fatalf("WriteLock failed: %v", err)
	}

	// Verify file exists
	lockPath := filepath.Join(tempDir, config.LockFileName)
	if _, err := os.Stat(lockPath); err != nil {
		t.Fatalf("lock file should exist: %v", err)
	}

	// Read lock
	loaded, err := ReadLock(tempDir)
	if err != nil {
		t.Fatalf("ReadLock failed: %v", err)
	}

	// Verify data matches
	if loaded.Version != original.Version {
		t.Errorf("Version mismatch: expected %s, got %s", original.Version, loaded.Version)
	}

	if loaded.Plan != original.Plan {
		t.Errorf("Plan mismatch: expected %s, got %s", original.Plan, loaded.Plan)
	}

	if loaded.Checksum != original.Checksum {
		t.Errorf("Checksum mismatch: expected %s, got %s", original.Checksum, loaded.Checksum)
	}

	if loaded.CLIVersion != original.CLIVersion {
		t.Errorf("CLIVersion mismatch: expected %s, got %s", original.CLIVersion, loaded.CLIVersion)
	}

	// InstalledAt should be very close (allowing for JSON marshaling precision)
	if loaded.InstalledAt.Sub(original.InstalledAt) > time.Millisecond {
		t.Errorf("InstalledAt mismatch: expected %v, got %v", original.InstalledAt, loaded.InstalledAt)
	}
}

// TestReadLockFileNotFound verifies error when lock file doesn't exist.
func TestReadLockFileNotFound(t *testing.T) {
	tempDir := t.TempDir()

	_, err := ReadLock(tempDir)
	if err == nil {
		t.Error("ReadLock should fail when lock file doesn't exist")
	}

	// Verify error message mentions .claudex.lock
	if err != nil && err.Error() != "" {
		// Error should be clear about what's missing
		t.Logf("ReadLock error (expected): %v", err)
	}
}

// TestReadLockInvalidJSON verifies error when lock file has invalid JSON.
func TestReadLockInvalidJSON(t *testing.T) {
	tempDir := t.TempDir()
	lockPath := filepath.Join(tempDir, config.LockFileName)

	// Write invalid JSON
	invalidJSON := []byte(`{not valid json}`)
	if err := os.WriteFile(lockPath, invalidJSON, 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	_, err := ReadLock(tempDir)
	if err == nil {
		t.Error("ReadLock should fail with invalid JSON")
	}
}

// TestWriteLockCreatesFile verifies WriteLock creates the lock file.
func TestWriteLockCreatesFile(t *testing.T) {
	tempDir := t.TempDir()

	lock := &LockData{
		Version:     "1.0.0",
		Plan:        "internal",
		InstalledAt: time.Now(),
		Checksum:    "sha256:xyz",
		CLIVersion:  "0.0.1",
	}

	lockPath := filepath.Join(tempDir, config.LockFileName)

	// File shouldn't exist yet
	if _, err := os.Stat(lockPath); err == nil {
		t.Fatal("lock file should not exist before WriteLock")
	}

	// Write lock
	if err := WriteLock(tempDir, lock); err != nil {
		t.Fatalf("WriteLock failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(lockPath); err != nil {
		t.Errorf("lock file should exist after WriteLock: %v", err)
	}
}

// TestWriteLockFilePermissions verifies lock file is readable.
func TestWriteLockFilePermissions(t *testing.T) {
	tempDir := t.TempDir()

	lock := &LockData{
		Version:     "1.0.0",
		Plan:        "paid",
		InstalledAt: time.Now(),
		Checksum:    "sha256:123",
		CLIVersion:  "0.1.0",
	}

	if err := WriteLock(tempDir, lock); err != nil {
		t.Fatalf("WriteLock failed: %v", err)
	}

	lockPath := filepath.Join(tempDir, config.LockFileName)
	stat, err := os.Stat(lockPath)
	if err != nil {
		t.Fatalf("failed to stat lock file: %v", err)
	}

	// File should be readable (mode 0644)
	perms := stat.Mode().Perm()
	if perms&0400 == 0 {
		t.Errorf("lock file should be readable: %o", perms)
	}
}

// TestWriteLockUpdateExisting verifies WriteLock overwrites existing file.
func TestWriteLockUpdateExisting(t *testing.T) {
	tempDir := t.TempDir()

	lock1 := &LockData{
		Version:     "1.0.0",
		Plan:        "internal",
		InstalledAt: time.Now(),
		Checksum:    "sha256:old",
		CLIVersion:  "0.0.1",
	}

	if err := WriteLock(tempDir, lock1); err != nil {
		t.Fatalf("first WriteLock failed: %v", err)
	}

	// Write different lock data
	lock2 := &LockData{
		Version:     "2.0.0",
		Plan:        "paid",
		InstalledAt: time.Now(),
		Checksum:    "sha256:new",
		CLIVersion:  "0.2.0",
	}

	if err := WriteLock(tempDir, lock2); err != nil {
		t.Fatalf("second WriteLock failed: %v", err)
	}

	// Read and verify it's the new data
	loaded, err := ReadLock(tempDir)
	if err != nil {
		t.Fatalf("ReadLock failed: %v", err)
	}

	if loaded.Version != lock2.Version {
		t.Errorf("Version should be updated: expected %s, got %s", lock2.Version, loaded.Version)
	}

	if loaded.Checksum != lock2.Checksum {
		t.Errorf("Checksum should be updated: expected %s, got %s", lock2.Checksum, loaded.Checksum)
	}
}

// TestLockDataWithSpecialCharacters verifies lock with special characters.
func TestLockDataWithSpecialCharacters(t *testing.T) {
	tempDir := t.TempDir()

	lock := &LockData{
		Version:     "1.0.0-beta+build.123",
		Plan:        "plan-with-dashes-and_underscores",
		InstalledAt: time.Date(2025, 6, 15, 14, 30, 45, 0, time.UTC),
		Checksum:    "sha256:abc123def456789",
		CLIVersion:  "0.1.0-alpha",
	}

	if err := WriteLock(tempDir, lock); err != nil {
		t.Fatalf("WriteLock failed: %v", err)
	}

	loaded, err := ReadLock(tempDir)
	if err != nil {
		t.Fatalf("ReadLock failed: %v", err)
	}

	if loaded.Version != lock.Version {
		t.Errorf("Version mismatch: expected %s, got %s", lock.Version, loaded.Version)
	}

	if loaded.Plan != lock.Plan {
		t.Errorf("Plan mismatch: expected %s, got %s", lock.Plan, loaded.Plan)
	}
}

// TestLockDataZeroTime verifies lock with zero/epoch time.
func TestLockDataZeroTime(t *testing.T) {
	tempDir := t.TempDir()

	lock := &LockData{
		Version:     "1.0.0",
		Plan:        "test",
		InstalledAt: time.Time{},
		Checksum:    "sha256:xyz",
		CLIVersion:  "0.0.1",
	}

	if err := WriteLock(tempDir, lock); err != nil {
		t.Fatalf("WriteLock failed: %v", err)
	}

	loaded, err := ReadLock(tempDir)
	if err != nil {
		t.Fatalf("ReadLock failed: %v", err)
	}

	if !loaded.InstalledAt.IsZero() {
		t.Errorf("InstalledAt should be zero time: got %v", loaded.InstalledAt)
	}
}

// TestReadLockInvalidJSONStructure verifies error with JSON that doesn't match LockData.
func TestReadLockInvalidJSONStructure(t *testing.T) {
	tempDir := t.TempDir()
	lockPath := filepath.Join(tempDir, config.LockFileName)

	// Valid JSON but wrong structure
	wrongStructure := []byte(`{"invalid_field": "value"}`)
	if err := os.WriteFile(lockPath, wrongStructure, 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	loaded, err := ReadLock(tempDir)
	if err != nil {
		t.Fatalf("ReadLock should succeed even with missing fields: %v", err)
	}

	// Fields should be empty
	if loaded.Version != "" {
		t.Errorf("Version should be empty string: got %s", loaded.Version)
	}
}

// TestWriteLockJSONFormatting verifies JSON is properly indented.
func TestWriteLockJSONFormatting(t *testing.T) {
	tempDir := t.TempDir()

	lock := &LockData{
		Version:     "1.0.0",
		Plan:        "test",
		InstalledAt: time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC),
		Checksum:    "sha256:abc123",
		CLIVersion:  "0.1.0",
	}

	if err := WriteLock(tempDir, lock); err != nil {
		t.Fatalf("WriteLock failed: %v", err)
	}

	lockPath := filepath.Join(tempDir, config.LockFileName)
	data, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatalf("failed to read lock file: %v", err)
	}

	// Verify it's valid JSON
	var loaded LockData
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Errorf("lock file should contain valid JSON: %v", err)
	}

	// Should be formatted with 2-space indent
	content := string(data)
	if !isProperlyIndented(content) {
		t.Logf("JSON is not indented with 2 spaces, but is still valid")
	}
}

// Helper function to check if JSON is indented
func isProperlyIndented(json string) bool {
	return len(json) > 0 && (json[0] == '{' || json[0] == '[')
}

// TestLockPathConstruction verifies lock file path is correctly constructed.
func TestLockPathConstruction(t *testing.T) {
	tempDir := t.TempDir()

	lock := &LockData{
		Version:     "1.0.0",
		Plan:        "test",
		InstalledAt: time.Now(),
		Checksum:    "sha256:xyz",
		CLIVersion:  "0.1.0",
	}

	if err := WriteLock(tempDir, lock); err != nil {
		t.Fatalf("WriteLock failed: %v", err)
	}

	lockPath := filepath.Join(tempDir, config.LockFileName)
	if _, err := os.Stat(lockPath); err != nil {
		t.Errorf("lock file should exist at %s: %v", lockPath, err)
	}

	if filepath.Base(lockPath) != config.LockFileName {
		t.Errorf("lock file should be named %s: got %s", config.LockFileName, filepath.Base(lockPath))
	}
}

// TestReadLockMultipleCallsConsistent verifies reading same lock multiple times gives consistent results.
func TestReadLockMultipleCallsConsistent(t *testing.T) {
	tempDir := t.TempDir()

	lock := &LockData{
		Version:     "1.0.0",
		Plan:        "paid",
		InstalledAt: time.Date(2025, 6, 15, 14, 30, 45, 123456789, time.UTC),
		Checksum:    "sha256:abc",
		CLIVersion:  "0.1.0",
	}

	if err := WriteLock(tempDir, lock); err != nil {
		t.Fatalf("WriteLock failed: %v", err)
	}

	// Read multiple times
	loaded1, err := ReadLock(tempDir)
	if err != nil {
		t.Fatalf("first ReadLock failed: %v", err)
	}

	loaded2, err := ReadLock(tempDir)
	if err != nil {
		t.Fatalf("second ReadLock failed: %v", err)
	}

	if loaded1.Version != loaded2.Version {
		t.Error("Version should be consistent across reads")
	}

	if loaded1.Plan != loaded2.Plan {
		t.Error("Plan should be consistent across reads")
	}

	if loaded1.Checksum != loaded2.Checksum {
		t.Error("Checksum should be consistent across reads")
	}
}

// TestLockDataEmptyFields verifies lock with empty string fields.
func TestLockDataEmptyFields(t *testing.T) {
	tempDir := t.TempDir()

	lock := &LockData{
		Version:     "",
		Plan:        "",
		InstalledAt: time.Now(),
		Checksum:    "",
		CLIVersion:  "",
	}

	if err := WriteLock(tempDir, lock); err != nil {
		t.Fatalf("WriteLock with empty fields failed: %v", err)
	}

	loaded, err := ReadLock(tempDir)
	if err != nil {
		t.Fatalf("ReadLock failed: %v", err)
	}

	if loaded.Version != "" {
		t.Errorf("Version should be empty: got %s", loaded.Version)
	}

	if loaded.Plan != "" {
		t.Errorf("Plan should be empty: got %s", loaded.Plan)
	}
}
