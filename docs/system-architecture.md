# ClaudeX CLI — System Architecture

## High-Level Overview

ClaudeX CLI is a stateless command-line client that communicates with a backend API (`claudex-api`) to manage license-bound rules distribution. The architecture emphasizes security (machine-specific auth), offline capability (local caching), and cross-platform support.

```
┌─────────────────────────────────────────────────────────────┐
│                    User's Machine                           │
│  ┌───────────────────────────────────────────────────────┐ │
│  │  ClaudeX CLI (Go Binary)                              │ │
│  │  ┌─────────────────────────────────────────────────┐ │ │
│  │  │  Commands: login, init, logout, status, etc.   │ │ │
│  │  └────────────────┬────────────────────────────────┘ │ │
│  │                   │                                   │ │
│  │  ┌────────────────▼──────────────────────────────┐  │ │
│  │  │  Core Services                                │  │ │
│  │  │  - Authentication                            │  │ │
│  │  │  - Rules Download & Extraction              │  │ │
│  │  │  - Configuration Management                 │  │ │
│  │  │  - Notification Setup                        │  │ │
│  │  └────────────────┬──────────────────────────────┘  │ │
│  │                   │                                   │ │
│  │  ┌────────────────▼──────────────────────────────┐  │ │
│  │  │  Local Storage (~/.claudex/)                 │  │ │
│  │  │  - auth.json (token, machine ID)            │  │ │
│  │  │  - config.json (notifications)              │  │ │
│  │  │  - projects.json (registry)                 │  │ │
│  │  │  - cache/ (bundle cache)                    │  │ │
│  │  │  - .claudex.lock (per-project)              │  │ │
│  │  └────────────────────────────────────────────┘  │ │
│  └───────────────────────────────────────────────────┘ │
│                     │                                    │
│                     │ HTTPS                              │
│                     ▼                                    │
└─────────────────────┬────────────────────────────────────┘
                      │
                      │ HTTP/HTTPS
                      ▼
         ┌────────────────────────────────┐
         │   claudex-api (Backend)        │
         │   - Login/Logout               │
         │   - Token Verification         │
         │   - Bundle Metadata            │
         │   - Signed URL Generation      │
         └────────────┬───────────────────┘
                      │
         ┌────────────▼───────────────┐
         │  Cloudflare R2 / S3         │
         │  (Signed URLs)              │
         └────────────────────────────┘
```

## Authentication & Device Binding

### Conceptual Flow

```
┌──────────┐                    ┌──────────┐                  ┌──────────┐
│  User    │                    │  Client  │                  │  Server  │
└────┬─────┘                    └────┬─────┘                  └────┬─────┘
     │                               │                             │
     │ claudex login --key=ABC       │                             │
     ├──────────────────────────────>│                             │
     │                               │ GenerateMachineID           │
     │                               │ (OS+Hostname+Arch)          │
     │                               │                             │
     │                               │ POST /login                 │
     │                               │ {key, machineID, info}      │
     │                               │────────────────────────────>│
     │                               │                             │ Verify key
     │                               │                             │ Bind device
     │                               │                             │ Generate token
     │                               │       {token, plan}         │
     │                               │<────────────────────────────│
     │                               │                             │
     │                               │ SaveAuth(token, machineID)  │
     │                               │ to ~/.claudex/auth.json     │
     │                               │                             │
     │ ✓ Authenticated               │                             │
     │<──────────────────────────────┤                             │
```

### Machine ID Generation

**Algorithm:** Deterministic hash of platform-specific identifiers

| Platform | Sources | Uniqueness |
|----------|---------|-----------|
| **Windows** | GUID (registry), OS, Arch | Very high (GUID is UUID) |
| **macOS** | Hardware UUID (system_profiler), OS, Arch | Very high (UUID-based) |
| **Linux** | /etc/machine-id, fallback /proc/cpuinfo, OS, Arch | High (system-assigned ID) |

**Purpose:** Prevent license sharing across devices; enable multi-device management with per-device binding.

**Storage:** Encrypted in `~/.claudex/auth.json` alongside token.

## Session Verification

Every authenticated command verifies the token with the API before proceeding:

```go
EnsureAuth(client, cfg) → 
  LoadAuth(authFile) →
    client.Verify(token, machineID) →
      API checks validity →
        Return AuthData or error
```

If verification fails, user must re-login.

## Rules Distribution Architecture

### Download Pipeline

```
┌──────────────────┐
│  `claudex init`  │
└────────┬─────────┘
         │
    ┌────▼─────────────────────┐
    │ Check local .claudex.lock │
    │ (exists? what version?)   │
    └────┬─────────────────────┘
         │
    ┌────▼──────────────────────┐
    │ POST /download {version}  │
    │ to API (with auth)        │
    └────┬──────────────────────┘
         │
    ┌────▼───────────────────────────┐
    │ API returns:                    │
    │ - version, signedURL, checksum  │
    │ - upToDate flag, changelog      │
    └────┬───────────────────────────┘
         │
    ┌────▼──────────────────┐
    │ Check cache for       │
    │ version+checksum      │
    │ (valid? serve local)  │
    └────┬─────────────────┘
         │
    ┌────▼──────────────────────┐
    │ GET from signed URL        │
    │ (5min timeout, 200MB max)  │
    │ Validate checksum          │
    └────┬───────────────────────┘
         │
    ┌────▼──────────────────┐
    │ Store in cache/       │
    │ (best-effort)         │
    └────┬─────────────────┘
         │
    ┌────▼──────────────────┐
    │ Extract ZIP to dir    │
    │ Write .claudex.lock   │
    │ Register in projects  │
    └────┬─────────────────┘
         │
    ┌────▼──────────────────┐
    │ ✓ Rules installed     │
    └──────────────────────┘
```

### Cache Strategy

**Location:** `~/.claudex/cache/{version}/`

**Validity Check:** Version + checksum must match API metadata. If either changes, cache is invalid and re-download occurs.

**Purpose:** Reduce bandwidth and latency for repeated deployments of same rules version.

**Cleanup:** Manual (users can delete `~/.claudex/cache/` to clear all).

## Lock File Design

**Location:** `.claudex.lock` (in project root)

**Purpose:** Track which version is installed, prevent accidental downgrades, enable audit trails.

**Schema:**
```json
{
  "version": "1.0.5",
  "plan": "pro",
  "installed_at": "2026-04-13T10:30:00Z",
  "checksum": "sha256:abc123...",
  "cli_version": "0.1.0"
}
```

**Usage:**
- On `claudex init`, compare current lock version against available version
- If up-to-date, skip download unless `--force`
- After successful installation, write new lock with metadata
- Used by CI/CD to verify rules consistency across environments

## Configuration Storage Architecture

### Data Directory Structure

```
~/.claudex/
├── auth.json        [0600] Encrypted token + machine ID
├── config.json      [0600] Notification configuration
├── projects.json    [0644] Project registry (non-sensitive)
└── cache/           [0700]
    ├── v1.0.0/          # Bundle cache by version
    ├── v1.0.1/
    └── ...
```

### Permissions Rationale

| File | Mode | Reason |
|------|------|--------|
| `auth.json` | 0600 | Contains secrets (token) |
| `config.json` | 0600 | Contains notification credentials |
| `projects.json` | 0644 | Non-sensitive metadata; readable by tools |
| `cache/` | 0700 | Bundle directory; restricted access |

### Environment Variable Overrides

```bash
CLAUDEX_SERVER="https://custom-api.example.com" claudex login
CLAUDEX_DATA_DIR="/var/claudex" claudex init
```

## Cryptographic Architecture

### Key Derivation

**Scenario:** Client has token and machine ID; needs encryption key for sensitive data.

**HKDF Process:**
```
Token (input keying material)
  + Machine ID (salt)
  + "claudex-wrap" (info/context)
  = 32-byte key (via HKDF-SHA256)
```

**Purpose:** Derive unique encryption keys per machine without exposing raw tokens.

### Encryption (AES-256-GCM)

**Wire Format:**
```
[12-byte nonce] + [ciphertext] + [16-byte auth tag]
```

**Properties:**
- Authenticated encryption (detects tampering)
- Randomized nonce (fresh per message)
- 256-bit key size (post-quantum resistant)

**Use Cases:**
- Wrapping server-provided keys
- Encrypting credentials at rest (future use)

## Multi-Platform Binary Management

### Build Matrix

| Platform | Architecture | Binary | Size (est.) |
|----------|--------------|--------|------------|
| Windows | amd64 | claudex-windows-amd64.exe | ~10 MB |
| macOS | amd64 | claudex-darwin-amd64 | ~9 MB |
| macOS | arm64 | claudex-darwin-arm64 | ~9 MB |
| Linux | amd64 | claudex-linux-amd64 | ~9 MB |

### Obfuscation (Production)

**Tool:** `garble` (Go code obfuscation)

**Purpose:** Obscure license logic, API endpoints, and error messages from reverse engineering.

**Options:**
- `-literals` — Encode string literals
- `-tiny` — Minimal binary size overhead
- `-seed=random` — Randomized per build (prevents reproducible builds for security)

**Dev Build:** No obfuscation (faster iteration, easier debugging)

## Notification System Architecture

### Flow: Interactive Setup

```
claudex notification
  ↓
[TUI] Choose provider (Telegram/Discord/Slack)
  ↓
[TUI] Enter credentials (bot token, webhook URL, etc.)
  ↓
Validate format
  ↓
Save to ~/.claudex/config.json (0600 perms)
  ↓
Offer to sync to .env
  ↓
✓ Configured
```

### Credential Storage

**Config File:** `~/.claudex/config.json`
```json
{
  "notification": {
    "provider": "telegram",
    "telegram": {
      "bot_token": "123:ABC...",
      "chat_id": "987654321"
    },
    "discord": { },
    "slack": { }
  }
}
```

**Env Sync:** Optional; writes credentials to project `.env` for downstream tools:
```bash
CLAUDEX_NOTIFY_PROVIDER=telegram
CLAUDEX_TELEGRAM_BOT_TOKEN=123:ABC...
CLAUDEX_TELEGRAM_CHAT_ID=987654321
```

## Error Handling Architecture

### Classification

| Category | Example | Handler |
|----------|---------|---------|
| **Auth** | Invalid key, device limit exceeded | Map to user message, suggest remediation |
| **Network** | Connection timeout, DNS failure | "Check internet connection" |
| **Validation** | Checksum mismatch, corrupted file | "Bundle integrity check failed" |
| **File IO** | Permission denied, disk full | Direct OS error with context |
| **Config** | Missing data dir, invalid JSON | "Run `claudex login` first" or "Corrupt config" |

### Propagation

```
Low-level error (crypto, IO)
  ↓
Wrapped with context (fmt.Errorf)
  ↓
CLI handler (mapAuthError, etc.)
  ↓
User-facing message (colored, actionable)
```

## Security Considerations

### Threat Model

| Threat | Mitigation |
|--------|-----------|
| **License sharing across devices** | Hardware-based machine ID binding; API validates |
| **Token theft** | Stored in user-owned files (0600 perms); encrypted at rest (future) |
| **Bundle tampering** | SHA-256 checksum validation against API response |
| **Man-in-the-middle (MITM)** | HTTPS-only for API and signed URL downloads; HTTPS redirect enforcement |
| **Reverse engineering** | Binary obfuscation via garble; string encoding |
| **Offline attacks** | No offline signature verification (out of scope for v0.1) |

### Data Handling

- **Sensitive Data:** Tokens, credentials stored in files with 0600 permissions
- **Logs:** No logging of credentials or tokens (CLI errors are descriptive but sanitized)
- **Memory:** Go runtime garbage collector handles cleanup (no manual zeroing)

## Scalability Considerations

**CLI Design:**
- Stateless: each invocation is independent
- No persistent background service or daemon
- No database dependency (local JSON files only)

**API Bottlenecks:**
- Token verification (every auth command) — cached at client would require token refresh
- Download rate limiting (not enforced by CLI; API can implement per-license quotas)

**Cache Growth:**
- Per-version bundles in cache/ — users should periodically clean old versions
- No automatic cache eviction (TTL-based cleanup not implemented in v0.1)

## Integration Points

### External Services

1. **claudex-api (Backend)**
   - Protocols: REST HTTP/HTTPS
   - Endpoints: /login, /logout, /verify, /download, /check-update
   - Auth: Token + Machine ID in request body

2. **Cloudflare R2 / S3 (Bundle Storage)**
   - Protocol: HTTPS signed URLs (time-limited)
   - No authentication credentials in CLI (API handles signing)
   - Max download size: 200 MB

3. **Notification Providers** (Optional)
   - Telegram: Bot API (polling or webhook)
   - Discord: Webhook URLs
   - Slack: Webhook URLs
   - Not called by CLI directly (credentials stored for downstream tools)

### File System Contracts

**Project Directory:**
- `.claudex.lock` — Lock file (JSON, created/updated by CLI)
- `.env` — Notification credentials (created if `--sync-env` used)

## Deployment & Distribution

### Release Process

1. **Versioning:** Semantic versioning (MAJOR.MINOR.PATCH), stored in `VERSION` file
2. **Build:** `make build-all` generates multi-platform binaries
3. **Distribution:** Binaries published to release artifacts; users download via web/package manager
4. **Installation:** Manual (copy binary to PATH) or via package managers (future)

### Update Strategy

- Manual: Users download and replace binary
- Auto-update: Not implemented in v0.1 (CLI does not update itself)
- Notification: API returns available versions; `claudex versions` lists them

---

## Future Architecture Enhancements

- **Daemon Mode:** Background service for automatic rule updates
- **Plugin System:** Third-party rules providers
- **Signature Verification:** Cryptographic signing of bundles
- **Offline Mode:** Fully offline operation with cached rules
- **Token Refresh:** Implement refresh tokens to reduce login frequency
