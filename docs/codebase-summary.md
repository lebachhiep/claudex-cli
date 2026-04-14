# ClaudeX CLI — Codebase Summary

## Directory Structure

```
claudex-cli/
├── cmd/
│   └── claudex/
│       └── main.go                 # CLI entry point; initializes root command
├── internal/
│   ├── api/
│   │   └── client.go              # HTTP client for claudex-api; request/response types
│   ├── auth/
│   │   ├── auth.go                # Login, logout, session verification logic
│   │   ├── machine.go             # Machine ID generation interface
│   │   ├── machine_{platform}.go  # Platform-specific hardware fingerprinting
│   │   └── store.go               # Auth file persistence (read/write/delete)
│   ├── cli/
│   │   ├── root.go                # Root command setup and subcommand registration
│   │   ├── login.go               # `claudex login` command
│   │   ├── logout.go              # `claudex logout` command
│   │   ├── init.go                # `claudex init` command (download & install rules)
│   │   ├── update.go              # `claudex update` command (deprecated wrapper)
│   │   ├── status.go              # `claudex status` command
│   │   ├── versions.go            # `claudex versions` command
│   │   ├── version.go             # `claudex version` command
│   │   └── notification.go        # `claudex notification` command (interactive setup)
│   ├── config/
│   │   ├── config.go              # Configuration management; data directory paths
│   │   └── config_test.go         # Unit tests for config package
│   ├── crypto/
│   │   ├── aes.go                 # AES-256-GCM encryption/decryption
│   │   └── hkdf.go                # HKDF key derivation
│   ├── notification/
│   │   ├── config.go              # Notification provider configuration (Telegram/Discord/Slack)
│   │   └── env_writer.go          # .env file synchronization for credentials
│   ├── projects/
│   │   └── store.go               # Projects registry (JSON persistence)
│   └── rules/
│       ├── download.go            # Rules bundle download via signed URL
│       ├── extract.go             # ZIP extraction and file placement
│       ├── validate.go            # SHA-256 checksum validation
│       ├── lock.go                # .claudex.lock file read/write
│       ├── cache.go               # Local bundle caching
│       └── extract_test.go        # Unit tests for extraction logic
├── Makefile                       # Build targets (dev, prod, cross-platform)
├── .goreleaser.yml               # GoReleaser config for multi-platform releases
├── go.mod, go.sum                # Go module dependencies
├── VERSION                       # Current version (0.1.0)
└── docs/                         # Documentation
```

## File Statistics

- **Total Lines of Code:** ~3,984 (excluding tests)
- **Go Files:** 34 (27 implementation + 7 test files)
- **Largest File:** `internal/cli/init.go` (~100 LOC)
- **Smallest Files:** Platform-specific machine ID generators (~30 LOC each)

## Key Packages & Responsibilities

### `cmd/claudex` (CLI Entry)
- **Entry Point:** `main.go`
- Injects build metadata (version, commit, date) via ldflags
- Delegates to `internal/cli.NewRootCmd()`

### `internal/api` (HTTP Communication)
- **Client:** HTTP wrapper with 30-second timeout
- **Types:** Request/response structures for all API endpoints
- Endpoints: Login, Logout, Verify, Download, CheckUpdate
- Error propagation from API server

### `internal/auth` (Authentication & Machine Binding)
- **Login Flow:**
  1. Generate machine ID (hardware fingerprint)
  2. Send LicenseKey + MachineID to API
  3. Receive Token + Plan from API
  4. Persist token in `~/.claudex/auth.json`
- **Logout Flow:** Notify API, delete local auth file
- **Verification:** Token validation on each authenticated command
- **Machine ID:** Deterministic hash of OS + hostname + arch (platform-specific)

### `internal/cli` (Command Definitions)
- **Root Command:** Aggregates all subcommands
- **Available Commands:**
  - `login --key=<LICENSE_KEY>` — Authenticate and bind machine
  - `logout` — Unbind machine and clear auth
  - `init [version]` — Download and install latest or specified version
  - `update` — Deprecated; delegates to `init`
  - `status` — Check authentication and project status
  - `versions` — List available rule versions
  - `version` — Show CLI version
  - `notification` — Interactive setup for Telegram/Discord/Slack
- **Shared State:** `cfg` (config) and `apiClient` initialized in `PersistentPreRun`

### `internal/config` (Configuration Management)
- **Data Directory:** `~/.claudex/` (customizable via `CLAUDEX_DATA_DIR` env var)
- **Default Server:** `https://api-dev.claudex.info` (customizable via `CLAUDEX_SERVER` env var)
- **Files Managed:**
  - `auth.json` — Serialized AuthData (token, machine ID, plan)
  - `config.json` — Notification configuration
  - `projects.json` — Installed projects registry
  - `cache/` — Cached rule bundles
- **Permissions:** Auth/config files: 0600 (user-only); cache: 0700

### `internal/crypto` (Encryption & Key Derivation)
- **AES-256-GCM:** Encrypt/decrypt sensitive data (nonce || ciphertext || tag)
- **HKDF:** Derive encryption keys from token + machine ID + context
- **Key Wrapping:** Unwrap server-provided wrapped keys using derived wrap key
- Constants: 12-byte nonce, 32-byte key, 16-byte auth tag

### `internal/notification` (Multi-Provider Alerts)
- **Providers:** Telegram (bot token + chat ID), Discord (webhook URL), Slack (webhook URL)
- **Storage:** Global config at `~/.claudex/config.json`
- **Env Sync:** Write credentials to project `.env` for downstream tools
- **Interactive Setup:** TUI prompts via charmbracelet/huh

### `internal/projects` (Project Registry)
- **Storage:** `~/.claudex/projects.json`
- **Tracks:** Installed projects with versions
- **Operations:** Register, update, query projects

### `internal/rules` (Bundle Management)
- **Download:** Fetch via signed R2 URL; cache hit check
- **Validate:** SHA-256 checksum verification against API response
- **Extract:** Unzip bundle, place files in project directory
- **Lock File:** `.claudex.lock` tracks version, plan, install time, checksum, CLI version
- **Cache:** Local caching with validity checks before serving
- **Download Limits:** 5-minute timeout, max 200 MB per bundle

## Build & Release Process

### Development Build
```bash
make build
# Output: dist/claudex (unobfuscated, full symbols)
```

### Production Build
```bash
make build-prod
# Uses garble: removes strings, applies code obfuscation, random seed
```

### Cross-Platform Release
```bash
make build-all
# Outputs:
# - dist/claudex-windows-amd64.exe
# - dist/claudex-darwin-amd64
# - dist/claudex-darwin-arm64
# - dist/claudex-linux-amd64
```

### Binary Metadata
Injected at build time via ldflags:
- `main.version` — From VERSION file (0.1.0)
- `main.commit` — Short git SHA
- `main.date` — UTC timestamp

## Testing

- **Test Coverage:** 6 test files (~1,300 LOC)
- **Test Packages:** config, auth, rules/extract
- **Command:** `make test` or `go test ./...`

## Dependencies Overview

| Package | Purpose | Version |
|---------|---------|---------|
| spf13/cobra | CLI framework | 1.9.1 |
| charmbracelet/huh | Interactive TUI | 1.0.0 |
| golang.org/x/crypto | Cryptography | 0.36.0 |
| fatih/color | Terminal colors | 1.18.0 |
| charmbracelet/bubbles | TUI components | 0.21.1+ |
| charmbracelet/bubbletea | TUI framework | 1.3.6 |

## Execution Flow: `claudex init`

1. **PersistentPreRun:** Load config, initialize API client
2. **Init Command:**
   - Verify authentication via `auth.EnsureAuth()`
   - Detect mode (init vs. update) by reading existing `.claudex.lock`
   - Download rules bundle via `rules.Download()`
   - Check cache first; only fetch if cache miss
   - Validate SHA-256 checksum
   - Extract bundle files to project directory
   - Write `.claudex.lock` metadata
   - Register/update project in `~/.claudex/projects.json`
   - Display success message with stats

## Error Handling Patterns

- **Auth Errors:** Map API error codes to user-friendly messages (invalid_key, device_limit_exceeded, etc.)
- **Network Errors:** Propagate with context ("cannot reach API server")
- **File Operations:** Explicit error wrapping with `fmt.Errorf`
- **Crypto Operations:** Validation of key sizes; clear decryption failure messages

## Security Considerations

1. **File Permissions:** Auth/config files 0600 to prevent unauthorized access
2. **HTTPS-Only:** Refuse non-HTTPS redirects during downloads
3. **Timeout Protections:** Network operations have explicit timeouts (30s API, 5min downloads)
4. **Memory:** No explicit credential zeroing (Go runtime gc handles cleanup)
5. **Token Storage:** Tokens stored locally but never logged or exposed in error messages

## Platform-Specific Details

- **Windows:** Machine ID includes Windows product GUID (registry lookup)
- **macOS:** Machine ID includes hardware UUID (system_profiler lookup)
- **Linux:** Machine ID includes /etc/machine-id or /proc/cpuinfo fallback
