# ClaudeX CLI — Project Overview & Product Development Requirements

## Executive Summary

ClaudeX CLI is a command-line interface for securely managing Claude Code skills distribution, authentication, and license binding. It enables users to download rules bundles, install project-specific configurations, and manage multi-device access through machine-specific authentication.

**Version:** 0.1.0 (beta)
**Language:** Go 1.23
**License:** Proprietary

## Project Vision

Provide a lightweight, cross-platform (Windows/macOS/Linux) CLI tool that:
- Securely authenticates users via license keys
- Binds authentication to individual machines (hardware fingerprinting)
- Manages project-specific rules bundles and `.claudex.lock` files
- Supports notifications via Telegram, Discord, or Slack
- Handles cryptographic key operations transparently to the user

## Core Objectives

1. **Secure Device Binding:** Bind licenses to machines using hardware identification, preventing license sharing
2. **Rules Distribution:** Download and validate pre-built rules bundles via signed R2 URLs
3. **Multi-Platform Support:** Binary releases for Windows (amd64), macOS (amd64/arm64), Linux (amd64)
4. **Transparent Authentication:** Handle token-based auth with automatic session verification
5. **Notification Integration:** Alert users of important events via configured providers

## Functional Requirements

### Authentication (FR-001)
- Users login with a license key via `claudex login --key=<LICENSE_KEY>`
- CLI generates a unique machine ID based on hardware characteristics
- API verifies license and returns a session token
- Token is stored in `~/.claudex/auth.json` with permissions 0600
- Session tokens are verified on each command requiring auth

### Rules Management (FR-002)
- `claudex init` downloads the latest rules bundle for a project
- Bundles are validated via SHA-256 checksum against API response
- Valid bundles are cached in `~/.claudex/cache/` for offline use
- Extract bundle contents and write `.claudex.lock` to project directory
- Lock file tracks version, plan, install timestamp, and CLI version

### Project Tracking (FR-003)
- Maintain a projects registry at `~/.claudex/projects.json`
- Track installed projects with their current versions
- Support multiple projects across the filesystem

### Configuration Management (FR-004)
- Configuration stored in `~/.claudex/config.json` (permissions 0600)
- Support environment variable overrides:
  - `CLAUDEX_SERVER`: Override API server URL
  - `CLAUDEX_DATA_DIR`: Override `~/.claudex` directory
- Interactive setup for notification providers (Telegram/Discord/Slack)

### Notifications (FR-005)
- Support Telegram, Discord, Slack as notification providers
- Interactive CLI prompts for provider configuration
- Sync credentials to `.env` files for downstream tools

## Non-Functional Requirements

### Security (NFR-001)
- All auth tokens and credentials encrypted at rest using AES-256-GCM
- HKDF-based key derivation from token and machine ID
- File permissions: auth/config 0600 (user-only), others 0644
- HTTPS-only for API communication and signed URL downloads
- Refuse non-HTTPS redirects during downloads

### Performance (NFR-002)
- API timeout: 30 seconds
- Download timeout: 5 minutes
- Max bundle size: 200 MB
- Cache check before download to avoid redundant transfers

### Reliability (NFR-003)
- Graceful error handling with user-friendly error messages
- Automatic retry logic for transient network failures (if applicable)
- Validation of downloaded bundles before extraction

### Usability (NFR-004)
- Interactive TUI using charmbracelet/huh for user prompts
- Colored output with disable flag (`--no-color`)
- Clear, actionable error messages with remediation suggestions

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│  CLI Commands (cmd/claudex/)                                    │
│  - login, logout, init, update, status, versions, notification  │
└──────────────────┬──────────────────────────────────────────────┘
                   │
        ┌──────────┴──────────┬──────────────┬───────────────┐
        │                     │              │               │
┌───────▼─────────┐  ┌────────▼──────┐  ┌──▼────────┐  ┌───▼─────────┐
│ auth/           │  │ rules/        │  │ api/      │  │ config/     │
│ - auth.go       │  │ - download.go │  │ client.go │  │ config.go   │
│ - store.go      │  │ - extract.go  │  │ types.go  │  └─────────────┘
│ - machine.go    │  │ - validate.go │  └───────────┘
└─────────────────┘  │ - lock.go     │
                     │ - cache.go    │
                     └───────────────┘
        │
        └──────────────────┬──────────────────┐
                           │                  │
                    ┌──────▼──────┐   ┌──────▼──────┐
                    │ crypto/     │   │ notification/
                    │ - aes.go    │   │ - config.go
                    │ - hkdf.go   │   │ - env_writer.go
                    └─────────────┘   └─────────────┘
```

## Key Design Decisions

1. **Hardware-Based Machine ID:** Deterministic fingerprint (OS + hostname + arch) prevents license duplication
2. **Token-Based Auth:** Stateless API design; each request includes token + machine ID for verification
3. **Signed URLs from R2:** Cloudflare R2 provides secure, time-limited download links without exposed secrets
4. **Local Caching:** Reduce bandwidth and latency for repeated downloads of same bundle
5. **Encrypted Config Storage:** Sensitive credentials protected via AES-256-GCM with derived keys

## Success Metrics

- CLI binary size < 20 MB across all platforms
- Auth/login latency < 2 seconds
- Rules download time < 1 minute for 100 MB bundle (assuming typical internet speed)
- 100% test coverage for crypto and auth packages
- Zero security vulnerabilities in dependencies

## Timeline & Phases

| Phase | Duration | Status |
|-------|----------|--------|
| Phase 1: Core Auth & Config | Weeks 1-2 | Complete |
| Phase 2: Rules Download & Extraction | Weeks 2-3 | Complete |
| Phase 3: Notifications & Advanced Features | Week 4 | Complete |
| Phase 4: Testing & Hardening | Week 5 | Complete |
| Phase 5: Cross-Platform Builds & Release | Week 6 | In Progress |

## Dependencies

- **Cobra:** CLI framework and command routing
- **charmbracelet/huh:** Interactive terminal UI for prompts
- **golang.org/x/crypto:** AES-GCM, HKDF cryptographic primitives
- **fatih/color:** Terminal color output

## Known Limitations & Future Work

- No offline mode for rules without prior cache (requires API connectivity)
- Notifications are optional; missing config not fatal
- No built-in auto-update for the CLI itself (use system package manager or manual installation)
- API server URL currently hardcoded per build; runtime override via `--server` flag

## Assumptions

- Users have write access to `~/.claudex/`
- Network connectivity to API server is available for auth
- Machine hardware identifiers are sufficiently unique (no identical machines expected in typical use)
- Users trust the installed rules bundles (signature verification is out of scope for v0.1)
