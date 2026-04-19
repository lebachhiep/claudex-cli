package auth

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/lebachhiep/claudex-cli/internal/api"
	"github.com/lebachhiep/claudex-cli/internal/config"
)

func writeFakeAuth(t *testing.T, authFile string) {
	t.Helper()
	if err := SaveAuth(authFile, &AuthData{
		Token:      "tok",
		MachineID:  "machine-1",
		Plan:       "paid",
		Server:     "https://api.test",
		LoggedInAt: time.Now().UTC(),
	}); err != nil {
		t.Fatalf("SaveAuth: %v", err)
	}
}

// TestLogoutSuccess: happy path — server returns 200, auth file wiped, localOnly=false.
func TestLogoutSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success":true,"data":{"message":"ok","devices_used":0,"devices_limit":3}}`))
	}))
	defer srv.Close()

	tempDir := t.TempDir()
	authFile := filepath.Join(tempDir, "auth.json")
	writeFakeAuth(t, authFile)

	cfg := &config.Config{AuthFile: authFile, ServerURL: srv.URL}
	client := api.NewClient(srv.URL)

	resp, localOnly, err := Logout(client, cfg)
	if err != nil {
		t.Fatalf("Logout: %v", err)
	}
	if localOnly {
		t.Error("expected localOnly=false on success")
	}
	if resp == nil || resp.DevicesLimit != 3 {
		t.Errorf("unexpected response: %+v", resp)
	}
	if IsLoggedIn(authFile) {
		t.Error("auth file should be deleted after successful logout")
	}
}

// TestLogoutLocalOnlyWhenSubscriptionGone: regression — server returns a "not found"-class
// APIError (e.g. subscription removed). Logout must still wipe local auth and signal localOnly.
func TestLogoutLocalOnlyWhenSubscriptionGone(t *testing.T) {
	cases := []struct {
		name    string
		status  int
		body    string
	}{
		{"not found", http.StatusNotFound, `{"error":"NOT_FOUND","message":"subscription not found"}`},
		{"expired", http.StatusUnauthorized, `{"error":"EXPIRED","message":"license expired"}`},
		{"not active", http.StatusForbidden, `{"error":"INACTIVE","message":"license not active"}`},
		{"not registered", http.StatusForbidden, `{"error":"GONE","message":"device not registered"}`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				w.Write([]byte(tc.body))
			}))
			defer srv.Close()

			tempDir := t.TempDir()
			authFile := filepath.Join(tempDir, "auth.json")
			writeFakeAuth(t, authFile)

			cfg := &config.Config{AuthFile: authFile, ServerURL: srv.URL}
			client := api.NewClient(srv.URL)

			resp, localOnly, err := Logout(client, cfg)
			if err != nil {
				t.Fatalf("Logout should not error on clearable APIError: %v", err)
			}
			if !localOnly {
				t.Error("expected localOnly=true when server session is gone")
			}
			if resp != nil {
				t.Errorf("expected nil response for localOnly path, got %+v", resp)
			}
			if IsLoggedIn(authFile) {
				t.Error("auth file should be deleted even when server-side session is gone")
			}
		})
	}
}

// TestLogoutPropagatesTransientErrors: non-clearable errors (network, 500) must bubble up
// AND leave local auth intact so the user can retry.
func TestLogoutPropagatesTransientErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"INTERNAL","message":"database unreachable"}`))
	}))
	defer srv.Close()

	tempDir := t.TempDir()
	authFile := filepath.Join(tempDir, "auth.json")
	writeFakeAuth(t, authFile)

	cfg := &config.Config{AuthFile: authFile, ServerURL: srv.URL}
	client := api.NewClient(srv.URL)

	_, localOnly, err := Logout(client, cfg)
	if err == nil {
		t.Fatal("expected error on 500")
	}
	if localOnly {
		t.Error("localOnly should be false when error is non-clearable")
	}
	if !IsLoggedIn(authFile) {
		t.Error("auth file should survive a transient server error so user can retry")
	}
}

// TestLogoutNotLoggedIn: no auth file → friendly error, nothing to do.
func TestLogoutNotLoggedIn(t *testing.T) {
	tempDir := t.TempDir()
	authFile := filepath.Join(tempDir, "missing.json")

	cfg := &config.Config{AuthFile: authFile, ServerURL: "http://unused"}
	client := api.NewClient("http://unused")

	_, _, err := Logout(client, cfg)
	if err == nil || err.Error() != "not logged in" {
		t.Errorf("expected 'not logged in' error, got %v", err)
	}
}
