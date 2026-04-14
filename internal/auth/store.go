package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// AuthData represents the persisted login state in ~/.claudex/auth.json.
type AuthData struct {
	Token      string    `json:"token"`
	MachineID  string    `json:"machine_id"`
	Plan       string    `json:"plan"`
	Server     string    `json:"server"`
	LoggedInAt time.Time `json:"logged_in_at"`
}

// LoadAuth reads auth.json from the given path.
func LoadAuth(authFile string) (*AuthData, error) {
	data, err := os.ReadFile(authFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("not logged in")
		}
		return nil, fmt.Errorf("read auth file: %w", err)
	}

	var auth AuthData
	if err := json.Unmarshal(data, &auth); err != nil {
		return nil, fmt.Errorf("parse auth file: %w", err)
	}
	return &auth, nil
}

// SaveAuth writes auth.json with 0600 permissions (owner read/write only).
func SaveAuth(authFile string, auth *AuthData) error {
	data, err := json.MarshalIndent(auth, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal auth: %w", err)
	}
	if err := os.WriteFile(authFile, data, 0600); err != nil {
		return fmt.Errorf("write auth file: %w", err)
	}
	return nil
}

// DeleteAuth removes the auth.json file.
func DeleteAuth(authFile string) error {
	if err := os.Remove(authFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete auth file: %w", err)
	}
	return nil
}

// IsLoggedIn checks if auth.json exists (lifetime license — no expiry check).
func IsLoggedIn(authFile string) bool {
	_, err := LoadAuth(authFile)
	return err == nil
}
