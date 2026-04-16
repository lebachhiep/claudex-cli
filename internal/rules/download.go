package rules

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/lebachhiep/claudex-cli/internal/api"
	"github.com/lebachhiep/claudex-cli/internal/auth"
	"github.com/lebachhiep/claudex-cli/internal/config"
)

// DownloadResult holds the bundle data ready for extraction.
type DownloadResult struct {
	Version   string
	Bundle    []byte // raw ZIP bytes (plaintext)
	Checksum  string
	SizeBytes int64
	UpToDate  bool
}

// Download fetches the rules bundle via signed URL from R2.
// Checks local cache first; downloads only if cache miss.
func Download(client *api.Client, authData *auth.AuthData, cfg *config.Config, currentVersion string, targetVersion string) (*DownloadResult, error) {
	resp, err := client.DownloadRules(&api.DownloadRequest{
		Token:          authData.Token,
		MachineID:      authData.MachineID,
		CurrentVersion: currentVersion,
		Version:        targetVersion,
	})
	if err != nil {
		return nil, err
	}

	if resp.UpToDate {
		return &DownloadResult{
			Version:  resp.Version,
			UpToDate: true,
		}, nil
	}

	// Check local cache
	cache := NewCache(cfg.CacheDir)
	if cached, _ := cache.GetIfValid(resp.Version, resp.Checksum); cached != nil {
		return &DownloadResult{
			Version:   resp.Version,
			Bundle:    cached,
			Checksum:  resp.Checksum,
			SizeBytes: resp.SizeBytes,
		}, nil
	}

	// Download from signed URL
	data, err := downloadFromURL(resp.SignedURL)
	if err != nil {
		return nil, fmt.Errorf("download from R2: %w", err)
	}

	// Verify checksum
	if err := VerifyChecksum(data, resp.Checksum); err != nil {
		return nil, fmt.Errorf("bundle integrity check failed: %w", err)
	}

	// Cache for future use (best-effort)
	_ = cache.Put(resp.Version, data)

	return &DownloadResult{
		Version:   resp.Version,
		Bundle:    data,
		Checksum:  resp.Checksum,
		SizeBytes: resp.SizeBytes,
	}, nil
}

// CheckUpdate queries the API for available updates without downloading.
func CheckUpdate(client *api.Client, authData *auth.AuthData, currentVersion string) (*api.CheckUpdateResponse, error) {
	return client.CheckUpdate(&api.CheckUpdateRequest{
		Token:          authData.Token,
		MachineID:      authData.MachineID,
		CurrentVersion: currentVersion,
	})
}

// downloadFromURL fetches data from a URL with a 5-minute timeout.
func downloadFromURL(url string) ([]byte, error) {
	httpClient := &http.Client{
		Timeout: 5 * time.Minute,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if req.URL.Scheme != "https" {
				return fmt.Errorf("refusing non-HTTPS redirect")
			}
			if len(via) > 3 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("HTTP GET failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	const maxSize = 200 * 1024 * 1024 // 200MB
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxSize))
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	return data, nil
}
