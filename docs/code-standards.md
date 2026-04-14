# ClaudeX CLI — Code Structure & Standards

## Project Organization

The ClaudeX CLI follows a clean, layered architecture separating concerns into domain-specific packages:

```
cmd/
  └── claudex/          # Executable entry point (minimal, delegates to internal)

internal/
  ├── cli/              # Command definitions (Cobra)
  ├── api/              # HTTP client and request/response types
  ├── auth/             # Authentication and machine binding logic
  ├── config/           # Configuration management
  ├── crypto/           # Encryption and key derivation
  ├── notification/     # Multi-provider notification setup
  ├── projects/         # Project registry persistence
  └── rules/            # Rules bundle download, validation, extraction
```

**Principle:** Each package has a single responsibility; packages depend downward (higher packages depend on lower, no circular dependencies).

## Naming Conventions

| Type | Convention | Example |
|------|-----------|---------|
| **Packages** | lowercase, short | `auth`, `crypto`, `rules` |
| **Types/Structs** | PascalCase | `AuthData`, `LockData`, `Client` |
| **Functions** | PascalCase (exported), camelCase (unexported) | `Login()`, `newInitCmd()` |
| **Constants** | UPPER_SNAKE_CASE | `NonceSize`, `KeySize`, `DefaultServerURL` |
| **Variables** | camelCase | `licenseKey`, `cfg`, `apiClient` |
| **Files** | snake_case.go | `client.go`, `auth.go`, `machine_windows.go` |

## Code Structure Guidelines

### 1. Package Organization

**Each package file should:**
- Start with a package comment explaining purpose
- Group related types and functions logically
- Keep files under 150 lines; split if exceeding this

**Example (auth/auth.go):**
```go
// Package auth handles login, logout, and session verification.
package auth

// Login authenticates with the API, binds this machine, and saves auth state.
func Login(client *api.Client, licenseKey string, cfg *config.Config) (*AuthData, *api.LoginResponse, error) {
    // implementation
}

// Logout unbinds this machine from the license and removes local auth state.
func Logout(client *api.Client, cfg *config.Config) (*api.LogoutResponse, error) {
    // implementation
}

// EnsureAuth loads auth state and verifies the token with the API.
func EnsureAuth(client *api.Client, cfg *config.Config) (*AuthData, error) {
    // implementation
}
```

### 2. Type Definitions

**Struct Types:**
- Include JSON tags with snake_case field names for API/file serialization
- Add comments to exported fields
- Use `json:"-"` for unexported or non-serialized fields

**Example:**
```go
// AuthData represents stored authentication state.
type AuthData struct {
    Token      string    `json:"token"`      // Session token from API
    MachineID  string    `json:"machine_id"` // Hardware fingerprint
    Plan       string    `json:"plan"`       // License plan tier
    Server     string    `json:"server"`     // API server URL used
    LoggedInAt time.Time `json:"logged_in_at"`
}
```

### 3. Function Design

**Public Functions (exported):**
- Exported if used by other packages or CLI commands
- Return error as last return value
- Include explanatory comment above function

**Example:**
```go
// Download fetches the rules bundle via signed URL from R2.
// Checks local cache first; downloads only if cache miss.
func Download(client *api.Client, authData *auth.AuthData, cfg *config.Config, currentVersion string, targetVersion string) (*DownloadResult, error) {
    // implementation
}
```

**Private Functions (unexported):**
- Used internally within a package
- Name starts with lowercase letter
- Example: `newInitCmd()`, `mapAuthError()`, `downloadFromURL()`

### 4. Error Handling

**Pattern:** Always return error as last value; wrap errors with context.

```go
// Good: wrap with context
if err := config.EnsureDataDir(); err != nil {
    return nil, fmt.Errorf("create data dir: %w", err)
}

// Good: format user-facing errors
return fmt.Errorf("license has been deactivated")

// Avoid: logging and error suppression
if err != nil {
    log.Printf("error: %v", err) // Don't use log; let caller handle
    return nil
}
```

**Error Messages:**
- Lowercase first letter for context-wrapped errors
- Concise; avoid redundant prefixes
- User-facing errors (CLI output) should be actionable

### 5. Dependency Injection

**Pattern:** Pass dependencies as function parameters, not globals.

```go
// Good: dependencies explicit
func Login(client *api.Client, licenseKey string, cfg *config.Config) error {
    // implementation
}

// Avoid: global state
var globalConfig *config.Config // Don't do this

func Login(licenseKey string) error {
    // implicit dependency on globalConfig
}
```

**Exception:** CLI shared state (`cfg`, `apiClient`) initialized in Cobra's `PersistentPreRunE` and available to all subcommands.

### 6. Constants & Magic Numbers

**All hardcoded values should be named constants:**

```go
// Good
const (
    DefaultServerURL = "https://api-dev.claudex.info"
    KeySize          = 32  // AES-256
    NonceSize        = 12  // GCM standard
    MaxBundleSize    = 200 * 1024 * 1024 // 200MB
)

// Avoid
url := "https://api-dev.claudex.info"
size := 32
```

## Testing Standards

### Test File Naming
- Test files: `{package}_test.go`
- Example: `auth/store_test.go`, `config/config_test.go`

### Test Function Naming
```go
func TestAuthData_Encrypt(t *testing.T) { }        // Method receiver
func TestLoadConfig_FileNotFound(t *testing.T) { } // Function + scenario
func TestDownloadWithCache(t *testing.T) { }       // Feature test
```

### Test Structure
```go
func TestLogin(t *testing.T) {
    // Arrange: Set up test data
    client := &api.Client{}
    licenseKey := "test-key-123"

    // Act: Execute function
    authData, resp, err := auth.Login(client, licenseKey, cfg)

    // Assert: Verify results
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if authData.Token == "" {
        t.Error("token should not be empty")
    }
}
```

## CLI Command Pattern (Cobra)

**Structure:** Each command is a separate file in `internal/cli/`.

```go
// internal/cli/login.go
func newLoginCmd() *cobra.Command {
    var licenseKey string

    cmd := &cobra.Command{
        Use:   "login",
        Short: "Authenticate with a license key and bind this machine",
        RunE: func(cmd *cobra.Command, args []string) error {
            // Command logic here; use shared `cfg` and `apiClient`
            return nil
        },
    }

    cmd.Flags().StringVar(&licenseKey, "key", "", "License key (required)")
    _ = cmd.MarkFlagRequired("key")

    return cmd
}
```

**Conventions:**
- Function name: `newXxxCmd()` (private constructor)
- Errors returned from `RunE` are formatted by Cobra's error handler
- Use color output via `fatih/color` for emphasis
- Support `--no-color` flag to disable colored output

## File Persistence Patterns

### JSON Serialization
```go
// Load: Read and unmarshal JSON
data, err := os.ReadFile(path)
if err != nil {
    if os.IsNotExist(err) {
        return &DefaultValue{}, nil // Return empty/default if not found
    }
    return nil, fmt.Errorf("read file: %w", err)
}

var value MyType
if err := json.Unmarshal(data, &value); err != nil {
    return nil, fmt.Errorf("parse json: %w", err)
}
return &value, nil

// Save: Marshal and write with specific permissions
data, err := json.MarshalIndent(value, "", "  ")
if err != nil {
    return fmt.Errorf("marshal json: %w", err)
}
if err := os.WriteFile(path, data, 0600); err != nil {
    return fmt.Errorf("write file: %w", err)
}
return nil
```

## Cryptographic Operations

### AES-256-GCM
```go
// Encrypt: plaintext → nonce || ciphertext || tag
ciphertext, err := crypto.EncryptGCM(key, plaintext)
if err != nil {
    return fmt.Errorf("encrypt: %w", err)
}

// Decrypt: nonce || ciphertext || tag → plaintext
plaintext, err := crypto.DecryptGCM(key, ciphertext)
if err != nil {
    return fmt.Errorf("decrypt: invalid key or corrupted data")
}
```

### HKDF Key Derivation
```go
// Derive 32-byte key from token, machine ID, and context string
key, err := crypto.DeriveKey(
    []byte(token),      // IKM (Input Keying Material)
    []byte(machineID),  // Salt
    []byte("context"),  // Info (distinguishes use case)
    32,                 // Output key length
)
if err != nil {
    return fmt.Errorf("derive key: %w", err)
}
```

## HTTP Client Patterns

### API Client
```go
// Create client with timeout
client := api.NewClient(baseURL)

// Make request
resp, err := client.Login(&api.LoginRequest{
    LicenseKey:  licenseKey,
    MachineID:   machineID,
    MachineInfo: machineInfo,
})
if err != nil {
    return fmt.Errorf("login failed: %w", err)
}
```

### Download with Safeguards
```go
httpClient := &http.Client{
    Timeout: 5 * time.Minute,
    CheckRedirect: func(req *http.Request, via []*http.Request) error {
        // Enforce HTTPS
        if req.URL.Scheme != "https" {
            return fmt.Errorf("refusing non-HTTPS redirect")
        }
        // Limit redirects
        if len(via) > 3 {
            return fmt.Errorf("too many redirects")
        }
        return nil
    },
}
```

## Configuration & Defaults

**Environment Variables (optional overrides):**
- `CLAUDEX_SERVER` — Override API server URL
- `CLAUDEX_DATA_DIR` — Override `~/.claudex` directory

**Default Values:**
- Server: `https://api-dev.claudex.info`
- Data directory: `~/.claudex/`
- File permissions: auth/config 0600, cache 0700, projects 0644

## Code Quality Standards

| Aspect | Standard |
|--------|----------|
| **Linting** | `golangci-lint run ./...` (pre-commit check) |
| **Format** | `gofmt` (automatic via IDE) |
| **Imports** | Organized in groups: std lib, third-party, internal |
| **Comments** | Document exported items; clarify complex logic |
| **Line Length** | Prefer < 100 chars (hard limit 120) |
| **Functions** | Keep under 50 lines; break out helpers if longer |
| **Tests** | 100% coverage for crypto and auth packages |

## Build & Compilation

**Development:** `make build` (no obfuscation, full symbols for debugging)

**Production:** `make build-prod` (garble obfuscation, string hiding, random seed)

**Compilation Flags:**
```
-ldflags "-s -w -X main.version=0.1.0 -X main.commit=abc1234 -X main.date=2026-04-13T..."
```
- `-s -w` — Strip symbols and DWARF for smaller binary
- `-X main.version` — Inject version string at build time
