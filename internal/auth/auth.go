package auth

import (
	"fmt"
	"strings"
	"time"

	"github.com/claudex/claudex-cli/internal/api"
	"github.com/claudex/claudex-cli/internal/config"
)

// Login authenticates with the API, binds this machine, and saves auth state.
func Login(client *api.Client, licenseKey string, cfg *config.Config) (*AuthData, *api.LoginResponse, error) {
	machineID, machineInfo, err := GenerateMachineID()
	if err != nil {
		return nil, nil, fmt.Errorf("generate machine ID: %w", err)
	}

	resp, err := client.Login(&api.LoginRequest{
		LicenseKey: licenseKey,
		MachineID:  machineID,
		MachineInfo: api.MachineInfo{
			OS:       machineInfo.OS,
			Hostname: machineInfo.Hostname,
			Arch:     machineInfo.Arch,
		},
	})
	if err != nil {
		return nil, nil, err
	}

	authData := &AuthData{
		Token:      resp.Token,
		MachineID:  machineID,
		Plan:       resp.Plan,
		Server:     cfg.ServerURL,
		LoggedInAt: time.Now().UTC(),
	}

	if err := cfg.EnsureDataDir(); err != nil {
		return nil, nil, fmt.Errorf("create data dir: %w", err)
	}

	if err := SaveAuth(cfg.AuthFile, authData); err != nil {
		return nil, nil, err
	}

	return authData, resp, nil
}

// Logout unbinds this machine from the license and removes local auth state.
func Logout(client *api.Client, cfg *config.Config) (*api.LogoutResponse, error) {
	authData, err := LoadAuth(cfg.AuthFile)
	if err != nil {
		return nil, fmt.Errorf("not logged in")
	}

	resp, err := client.Logout(&api.LogoutRequest{
		Token:     authData.Token,
		MachineID: authData.MachineID,
	})
	if err != nil {
		return nil, err
	}

	if err := DeleteAuth(cfg.AuthFile); err != nil {
		return nil, err
	}

	return resp, nil
}

// EnsureAuth loads auth state and verifies the token with the API.
// Returns auth data for downstream use, or error if not authenticated.
func EnsureAuth(client *api.Client, cfg *config.Config) (*AuthData, error) {
	authData, err := LoadAuth(cfg.AuthFile)
	if err != nil {
		return nil, fmt.Errorf("not logged in. Run `claudex login --key=YOUR_KEY` first")
	}

	resp, err := client.Verify(&api.VerifyRequest{
		Token:     authData.Token,
		MachineID: authData.MachineID,
	})
	if err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			if shouldClearAuth(apiErr) {
				_ = DeleteAuth(cfg.AuthFile)
			}
			return nil, fmt.Errorf("%s", mapVerifyError(apiErr))
		}
		return nil, fmt.Errorf("verification failed: %w", err)
	}

	if !resp.Valid {
		_ = DeleteAuth(cfg.AuthFile)
		return nil, fmt.Errorf("session invalid. Run `claudex login --key=YOUR_KEY` to re-authenticate")
	}

	return authData, nil
}

// mapVerifyError converts API errors from verify endpoint to user-friendly messages.
func mapVerifyError(apiErr *api.APIError) string {
	msg := apiErr.Message
	switch {
	case strings.Contains(msg, "expired"):
		return "Session expired. Run `claudex login --key=YOUR_KEY` to re-authenticate"
	case strings.Contains(msg, "not active"):
		return "License deactivated. Contact support or check your subscription status"
	case strings.Contains(msg, "not found"):
		return "License not found. Your subscription may have been removed"
	case strings.Contains(msg, "not registered"):
		return "Device removed. Run `claudex login --key=YOUR_KEY` to re-register this machine"
	case strings.Contains(msg, "No token"):
		return "Not logged in. Run `claudex login --key=YOUR_KEY` first"
	default:
		return fmt.Sprintf("Authentication failed: %s", msg)
	}
}

// shouldClearAuth returns true if the error indicates auth.json is permanently stale.
func shouldClearAuth(apiErr *api.APIError) bool {
	msg := apiErr.Message
	return strings.Contains(msg, "expired") ||
		strings.Contains(msg, "not registered") ||
		strings.Contains(msg, "not found") ||
		strings.Contains(msg, "not active")
}
