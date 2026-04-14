// Package api provides HTTP client for communicating with claudex-api.
package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client wraps HTTP communication with claudex-api.
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewClient creates an API client with 30s timeout.
func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// --- Request/Response types ---

// MachineInfo describes the machine hardware for device binding.
type MachineInfo struct {
	OS       string `json:"os"`
	Hostname string `json:"hostname"`
	Arch     string `json:"arch"`
}

type LoginRequest struct {
	LicenseKey  string      `json:"license_key"`
	MachineID   string      `json:"machine_id"`
	MachineInfo MachineInfo `json:"machine_info"`
}

type LoginResponse struct {
	Token        string `json:"token"`
	Plan         string `json:"plan"`
	DevicesUsed  int    `json:"devices_used"`
	DevicesLimit int    `json:"devices_limit"`
	Version      string `json:"version"`
}

type LogoutRequest struct {
	Token     string `json:"token"`
	MachineID string `json:"machine_id"`
}

type LogoutResponse struct {
	Message      string `json:"message"`
	DevicesUsed  int    `json:"devices_used"`
	DevicesLimit int    `json:"devices_limit"`
}

type VerifyRequest struct {
	Token     string `json:"token"`
	MachineID string `json:"machine_id"`
}

type VerifyResponse struct {
	Valid bool   `json:"valid"`
	Plan  string `json:"plan"`
}

type DownloadRequest struct {
	Token          string `json:"token"`
	MachineID      string `json:"machine_id"`
	CurrentVersion string `json:"current_version"`
	Version        string `json:"version,omitempty"`
}

type DownloadResponse struct {
	Version   string `json:"version"`
	SignedURL string `json:"signed_url"`
	Checksum  string `json:"checksum"`
	SizeBytes int64  `json:"size_bytes"`
	Changelog string `json:"changelog"`
	UpToDate  bool   `json:"up_to_date"`
}

type VersionInfo struct {
	Version   string `json:"version"`
	Changelog string `json:"changelog"`
	SizeBytes int64  `json:"size_bytes"`
	CreatedAt string `json:"created_at"`
}

type CheckUpdateRequest struct {
	Token          string `json:"token"`
	MachineID      string `json:"machine_id"`
	CurrentVersion string `json:"current_version"`
}

type CheckUpdateResponse struct {
	CurrentVersion string `json:"current_version"`
	LatestVersion  string `json:"latest_version"`
	HasUpdate      bool   `json:"has_update"`
	Changelog      string `json:"changelog"`
}

// APIError represents a structured error from the server.
type APIError struct {
	StatusCode int
	ErrorCode  string `json:"error"`
	Message    string `json:"message"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode, e.Message)
}

// --- API methods ---

func (c *Client) Login(req *LoginRequest) (*LoginResponse, error) {
	var resp LoginResponse
	if err := c.post("/cli/auth/login", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) Logout(req *LogoutRequest) (*LogoutResponse, error) {
	var resp LogoutResponse
	if err := c.postAuth("/cli/auth/logout", req.Token, req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) Verify(req *VerifyRequest) (*VerifyResponse, error) {
	var resp VerifyResponse
	if err := c.postAuth("/cli/auth/verify", req.Token, req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) DownloadRules(req *DownloadRequest) (*DownloadResponse, error) {
	var resp DownloadResponse
	if err := c.postAuth("/cli/rules/download", req.Token, req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) CheckUpdate(req *CheckUpdateRequest) (*CheckUpdateResponse, error) {
	var resp CheckUpdateResponse
	if err := c.postAuth("/cli/rules/check-update", req.Token, req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) GetVersions(token string) ([]VersionInfo, error) {
	var resp []VersionInfo
	if err := c.getAuth("/cli/rules/versions", token, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// --- HTTP helpers ---

// getAuth sends a GET with Authorization: Bearer header for guarded endpoints.
func (c *Client) getAuth(path, token string, result any) error {
	url := c.BaseURL + path
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	return c.doRequest(req, result)
}

// postAuth sends a POST with Authorization: Bearer header for guarded endpoints.
func (c *Client) postAuth(path, token string, body any, result any) error {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	url := c.BaseURL + path
	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	return c.doRequest(req, result)
}

func (c *Client) post(path string, body any, result any) error {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	url := c.BaseURL + path
	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	return c.doRequest(req, result)
}

func (c *Client) doRequest(req *http.Request, result any) error {
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("cannot reach API server: %w", err)
	}
	defer resp.Body.Close()

	const maxResponseSize = 50 * 1024 * 1024 // 50MB
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize))
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		var apiErr APIError
		if json.Unmarshal(respBody, &apiErr) == nil {
			apiErr.StatusCode = resp.StatusCode
			return &apiErr
		}
		return &APIError{
			StatusCode: resp.StatusCode,
			ErrorCode:  "unknown",
			Message:    string(respBody),
		}
	}

	if result != nil {
		// API wraps responses in { success: true, data: {...} }
		var wrapper struct {
			Success bool            `json:"success"`
			Data    json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal(respBody, &wrapper); err == nil && wrapper.Data != nil {
			respBody = wrapper.Data
		}
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}
