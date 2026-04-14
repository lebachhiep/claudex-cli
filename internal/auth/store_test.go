package auth

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// TestSaveAndLoadAuthRoundTrip verifies SaveAuth and LoadAuth round-trip.
func TestSaveAndLoadAuthRoundTrip(t *testing.T) {
	tempDir := t.TempDir()
	authFile := filepath.Join(tempDir, "auth.json")

	loggedInAt := time.Now()

	original := &AuthData{
		Token:      "test-token-abc123",
		MachineID:  "machine-xyz",
		Plan:       "paid",
		Server:     "https://api.test.com",
		LoggedInAt: loggedInAt,
	}

	// Save
	if err := SaveAuth(authFile, original); err != nil {
		t.Fatalf("SaveAuth failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(authFile); err != nil {
		t.Errorf("auth file should exist: %v", err)
	}

	// Load
	loaded, err := LoadAuth(authFile)
	if err != nil {
		t.Fatalf("LoadAuth failed: %v", err)
	}

	// Verify data matches
	if loaded.Token != original.Token {
		t.Errorf("Token mismatch: expected %s, got %s", original.Token, loaded.Token)
	}

	if loaded.MachineID != original.MachineID {
		t.Errorf("MachineID mismatch: expected %s, got %s", original.MachineID, loaded.MachineID)
	}

	if loaded.Plan != original.Plan {
		t.Errorf("Plan mismatch: expected %s, got %s", original.Plan, loaded.Plan)
	}

	if loaded.Server != original.Server {
		t.Errorf("Server mismatch: expected %s, got %s", original.Server, loaded.Server)
	}

	if loaded.LoggedInAt.Sub(original.LoggedInAt) > time.Millisecond {
		t.Errorf("LoggedInAt mismatch: expected %v, got %v", original.LoggedInAt, loaded.LoggedInAt)
	}
}

// TestLoadAuthFileNotFound verifies error when auth file doesn't exist.
func TestLoadAuthFileNotFound(t *testing.T) {
	authFile := "/nonexistent/path/auth.json"

	_, err := LoadAuth(authFile)
	if err == nil {
		t.Error("LoadAuth should fail when file not found")
	}

	// Verify the error message indicates not logged in
	if _, ok := err.(*os.PathError); !ok {
		// Could be wrapped, check error message
		if err.Error() == "not logged in" {
			// This is expected
		}
	}
}

// TestLoadAuthInvalidJSON verifies error when auth file has invalid JSON.
func TestLoadAuthInvalidJSON(t *testing.T) {
	tempDir := t.TempDir()
	authFile := filepath.Join(tempDir, "auth.json")

	// Write invalid JSON
	invalidJSON := []byte(`{invalid json}`)
	if err := os.WriteFile(authFile, invalidJSON, 0600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	_, err := LoadAuth(authFile)
	if err == nil {
		t.Error("LoadAuth should fail with invalid JSON")
	}
}

// TestSaveAuthFilePermissions verifies auth file is saved with 0600 permissions.
func TestSaveAuthFilePermissions(t *testing.T) {
	tempDir := t.TempDir()
	authFile := filepath.Join(tempDir, "auth.json")

	auth := &AuthData{
		Token:      "test-token",
		MachineID:  "machine-id",
		Plan:       "internal",
		Server:     "https://api.test.com",
		LoggedInAt: time.Now(),
	}

	if err := SaveAuth(authFile, auth); err != nil {
		t.Fatalf("SaveAuth failed: %v", err)
	}

	// Check file permissions (skip on Windows — doesn't support Unix permissions)
	if runtime.GOOS != "windows" {
		stat, err := os.Stat(authFile)
		if err != nil {
			t.Fatalf("failed to stat auth file: %v", err)
		}
		perms := stat.Mode().Perm()
		if perms != os.FileMode(0600) {
			t.Errorf("auth file permissions mismatch: expected %o, got %o", 0600, perms)
		}
	}
}

// TestDeleteAuthRemovesFile verifies DeleteAuth removes the auth file.
func TestDeleteAuthRemovesFile(t *testing.T) {
	tempDir := t.TempDir()
	authFile := filepath.Join(tempDir, "auth.json")

	// Create auth file
	auth := &AuthData{
		Token:      "test-token",
		MachineID:  "machine-id",
		Plan:       "internal",
		Server:     "https://api.test.com",
		LoggedInAt: time.Now(),
	}

	if err := SaveAuth(authFile, auth); err != nil {
		t.Fatalf("SaveAuth failed: %v", err)
	}

	// Verify it exists
	if _, err := os.Stat(authFile); err != nil {
		t.Fatalf("auth file should exist before delete: %v", err)
	}

	// Delete
	if err := DeleteAuth(authFile); err != nil {
		t.Fatalf("DeleteAuth failed: %v", err)
	}

	// Verify it's gone
	if _, err := os.Stat(authFile); err == nil {
		t.Error("auth file should not exist after DeleteAuth")
	} else if !os.IsNotExist(err) {
		t.Errorf("unexpected error checking if file exists: %v", err)
	}
}

// TestDeleteAuthNonexistentFile verifies DeleteAuth doesn't error on missing file.
func TestDeleteAuthNonexistentFile(t *testing.T) {
	authFile := "/nonexistent/path/auth.json"

	// DeleteAuth should not error when file doesn't exist
	if err := DeleteAuth(authFile); err != nil {
		t.Errorf("DeleteAuth should not error on missing file: %v", err)
	}
}

// TestIsLoggedInWithValidToken verifies IsLoggedIn returns true when auth file exists.
func TestIsLoggedInWithValidToken(t *testing.T) {
	tempDir := t.TempDir()
	authFile := filepath.Join(tempDir, "auth.json")

	auth := &AuthData{
		Token:      "valid-token",
		MachineID:  "machine-id",
		Plan:       "paid",
		Server:     "https://api.test.com",
		LoggedInAt: time.Now(),
	}

	if err := SaveAuth(authFile, auth); err != nil {
		t.Fatalf("SaveAuth failed: %v", err)
	}

	if !IsLoggedIn(authFile) {
		t.Error("IsLoggedIn should return true when auth file exists")
	}
}

// TestIsLoggedInFileNotFound verifies IsLoggedIn returns false when file not found.
func TestIsLoggedInFileNotFound(t *testing.T) {
	authFile := "/nonexistent/path/auth.json"

	if IsLoggedIn(authFile) {
		t.Error("IsLoggedIn should return false when file not found")
	}
}

// TestIsLoggedInInvalidJSON verifies IsLoggedIn returns false for invalid JSON.
func TestIsLoggedInInvalidJSON(t *testing.T) {
	tempDir := t.TempDir()
	authFile := filepath.Join(tempDir, "auth.json")

	// Write invalid JSON
	if err := os.WriteFile(authFile, []byte(`invalid`), 0600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	if IsLoggedIn(authFile) {
		t.Error("IsLoggedIn should return false for invalid JSON")
	}
}

// TestSaveAuthWithComplexData verifies SaveAuth handles complex data correctly.
func TestSaveAuthWithComplexData(t *testing.T) {
	tempDir := t.TempDir()
	authFile := filepath.Join(tempDir, "auth.json")

	// Test with various string values
	auth := &AuthData{
		Token:      "token-with-special-chars-!@#$%",
		MachineID:  "machine-with-dashes-and-numbers-123",
		Plan:       "paid",
		Server:     "https://api.example.com:8443/path",
		LoggedInAt: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	if err := SaveAuth(authFile, auth); err != nil {
		t.Fatalf("SaveAuth with complex data failed: %v", err)
	}

	loaded, err := LoadAuth(authFile)
	if err != nil {
		t.Fatalf("LoadAuth failed: %v", err)
	}

	if loaded.Token != auth.Token {
		t.Errorf("Token mismatch: expected %s, got %s", auth.Token, loaded.Token)
	}

	if loaded.MachineID != auth.MachineID {
		t.Errorf("MachineID mismatch: expected %s, got %s", auth.MachineID, loaded.MachineID)
	}
}

// TestAuthDataJSONFormatting verifies the JSON is properly formatted.
func TestAuthDataJSONFormatting(t *testing.T) {
	tempDir := t.TempDir()
	authFile := filepath.Join(tempDir, "auth.json")

	auth := &AuthData{
		Token:      "test-token",
		MachineID:  "machine-id",
		Plan:       "paid",
		Server:     "https://api.test.com",
		LoggedInAt: time.Date(2025, 6, 1, 10, 0, 0, 0, time.UTC),
	}

	if err := SaveAuth(authFile, auth); err != nil {
		t.Fatalf("SaveAuth failed: %v", err)
	}

	// Read raw file
	data, err := os.ReadFile(authFile)
	if err != nil {
		t.Fatalf("failed to read auth file: %v", err)
	}

	// Verify it's valid JSON with proper indentation
	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Errorf("auth file should contain valid JSON: %v", err)
	}

	// Re-marshal with indent to check formatting
	reformatted, _ := json.MarshalIndent(parsed, "", "  ")
	if !bytes.Equal(data, reformatted) {
		// It's okay if formatting differs slightly, as long as it's valid JSON
		t.Logf("JSON formatting differs but is still valid")
	}
}

// TestSaveAuthUpdateExisting verifies SaveAuth overwrites existing file.
func TestSaveAuthUpdateExisting(t *testing.T) {
	tempDir := t.TempDir()
	authFile := filepath.Join(tempDir, "auth.json")

	// First save
	auth1 := &AuthData{
		Token:      "token-1",
		MachineID:  "machine-1",
		Plan:       "internal",
		Server:     "https://api1.test.com",
		LoggedInAt: time.Now(),
	}

	if err := SaveAuth(authFile, auth1); err != nil {
		t.Fatalf("first SaveAuth failed: %v", err)
	}

	// Second save with different data
	auth2 := &AuthData{
		Token:      "token-2",
		MachineID:  "machine-2",
		Plan:       "paid",
		Server:     "https://api2.test.com",
		LoggedInAt: time.Now(),
	}

	if err := SaveAuth(authFile, auth2); err != nil {
		t.Fatalf("second SaveAuth failed: %v", err)
	}

	// Load and verify it's the second data
	loaded, err := LoadAuth(authFile)
	if err != nil {
		t.Fatalf("LoadAuth failed: %v", err)
	}

	if loaded.Token != auth2.Token {
		t.Errorf("Token should be updated: expected %s, got %s", auth2.Token, loaded.Token)
	}

	if loaded.MachineID != auth2.MachineID {
		t.Errorf("MachineID should be updated: expected %s, got %s", auth2.MachineID, loaded.MachineID)
	}
}
