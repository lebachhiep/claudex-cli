# ClaudeX CLI — Project Roadmap

## Release Status

**Current Version:** 0.1.0 (Beta)
**Release Date:** April 2026
**Maintenance Status:** Active development

## Version 0.1.0 (Current — Beta)

### Completed Features

#### Core Authentication (Milestone 1)
- [x] Machine-specific device binding (hardware fingerprinting)
- [x] License key validation via API
- [x] Session token generation and storage (encrypted)
- [x] Token verification on authenticated commands
- [x] Multi-device support (device limit tracking)
- [x] `claudex login --key=<LICENSE_KEY>` command
- [x] `claudex logout` command

#### Rules Distribution (Milestone 2)
- [x] Download rules bundles via signed R2 URLs
- [x] SHA-256 checksum validation
- [x] Local bundle caching for offline use
- [x] ZIP extraction and file placement
- [x] `.claudex.lock` file generation and tracking
- [x] Version comparison and update detection
- [x] `claudex init [version]` command (install/update)
- [x] Project registry (`~/.claudex/projects.json`)

#### Configuration & Storage (Milestone 3)
- [x] Configuration directory (`~/.claudex/`)
- [x] Secure file storage (0600 permissions)
- [x] Environment variable overrides (`CLAUDEX_SERVER`, `CLAUDEX_DATA_DIR`)
- [x] `claudex status` command (check auth and project status)
- [x] `claudex versions` command (list available versions)

#### Notifications (Milestone 4)
- [x] Telegram provider support (bot token + chat ID)
- [x] Discord provider support (webhook URL)
- [x] Slack provider support (webhook URL)
- [x] Interactive TUI setup (`claudex notification`)
- [x] Credential storage in `~/.claudex/config.json`
- [x] `.env` file synchronization

#### Cryptography (Milestone 5)
- [x] AES-256-GCM encryption/decryption
- [x] HKDF key derivation (token + machine ID)
- [x] Secure key wrapping and unwrapping

#### Build & Distribution (Milestone 6)
- [x] Development build (`make build`)
- [x] Production build with obfuscation (`make build-prod`)
- [x] Cross-platform compilation (Windows/macOS/Linux × amd64/arm64)
- [x] Binary metadata injection (version, commit, date)
- [x] GoReleaser configuration for multi-platform releases

#### Testing (Milestone 7)
- [x] Unit tests for auth, config, rules packages
- [x] Test coverage for crypto operations
- [x] Test utilities for file and API mocking

#### Documentation (Milestone 8)
- [x] Project overview & PDR (Product Development Requirements)
- [x] Codebase summary and directory structure
- [x] Code standards and style guidelines
- [x] System architecture documentation
- [x] Project roadmap
- [x] README with setup and usage instructions

### Known Limitations (v0.1.0)

| Limitation | Impact | Future Solution |
|-----------|--------|------------------|
| No offline mode without prior cache | Users must be online for first install | Cache all bundles on first login |
| Manual binary updates | Users must download new binaries | Auto-update daemon (v0.2) |
| No signature verification for bundles | Bundle origin not cryptographically verified | GPG/Ed25519 signing (v0.2) |
| String literals visible in obfuscated binary | Reverse engineering possible | Enhanced garble configuration (v0.1.1) |
| Single license per user | Multi-tenant not supported | Team/org licensing (future major version) |
| No credential rotation | Tokens persist indefinitely | Token refresh mechanism (v0.2) |

---

## Version 0.1.1 (Planned — Bug Fixes & Polish)

**Target:** Q2 2026 (May–June)

### Security Hardening
- [ ] Implement secure credential zeroing (override memory before dealloc)
- [ ] Add HTTP Public Key Pinning (HPKP) for API certificates
- [ ] Enhance garble configuration: hide symbol names, obfuscate control flow
- [ ] Add binary signature verification (code signing certificates)

### Bug Fixes
- [ ] Fix potential race conditions in file writes (atomic writes with temp files)
- [ ] Improve error messages for edge cases (corrupted cache, partial downloads)
- [ ] Handle network timeouts more gracefully (retry logic with exponential backoff)
- [ ] Fix platform-specific machine ID edge cases (Windows GUID fallback)

### Performance Improvements
- [ ] Cache API responses (metadata, version lists) with TTL
- [ ] Parallel download support for large bundles
- [ ] Reduce binary size via CGO_ENABLED=0 and UPX compression

### User Experience
- [ ] Add progress bars for downloads (`github.com/schollz/progressbar`)
- [ ] Improve TUI accessibility (keyboard navigation improvements)
- [ ] Add `claudex config` command for managing settings
- [ ] Support config file templates for common setups

### Testing
- [ ] Add integration tests with mock API server
- [ ] Platform-specific testing (VM-based CI for macOS, Windows)
- [ ] Stress testing: large bundles (100+ MB), many projects, cache growth
- [ ] End-to-end tests with real API (private staging environment)

---

## Version 0.2.0 (Planned — Advanced Features)

**Target:** Q3–Q4 2026 (July–December)

### Token Refresh & Session Management
- [ ] Implement refresh tokens (long-lived, used to issue short-lived access tokens)
- [ ] Automatic token rotation on expiry
- [ ] Session timeout handling (graceful re-login prompts)
- [ ] Device trust and PIN-based secondary auth

### Auto-Update Mechanism
- [ ] Background daemon (`claudex daemon`) for periodic update checks
- [ ] Staged rollouts (gradual user adoption)
- [ ] Automatic rollback on failed updates
- [ ] Update notifications via configured providers

### Bundle Signing & Verification
- [ ] GPG/Ed25519 signature verification for bundles
- [ ] CA certificate pinning for API
- [ ] Certificate transparency logging
- [ ] Revocation checking (CRL or OCSP)

### Advanced Caching
- [ ] Intelligent cache eviction (LRU with TTL)
- [ ] Cache compression (zstd or brotli for stored bundles)
- [ ] Partial download resume (HTTP Range requests)
- [ ] P2P bundle sharing (optional, opt-in)

### Notification Enhancements
- [ ] Custom notification templates
- [ ] Notification history and log
- [ ] Scheduled digests (daily/weekly summaries)
- [ ] Webhook integration for custom handlers

### Configuration Management
- [ ] `claudex config set <key> <value>` for runtime config changes
- [ ] Configuration validation and schema
- [ ] Config migration tools for version upgrades
- [ ] Profile support (dev/staging/prod configs)

### Project Management
- [ ] `claudex projects list/remove` for project management
- [ ] Project-specific rules overrides
- [ ] Dependency management between rules bundles
- [ ] Lock file version constraints (semver support)

---

## Version 1.0.0 (Planned — General Availability)

**Target:** Q1–Q2 2027 (January–June)

### Stability & Hardening
- [ ] 99.9% uptime SLA on API
- [ ] Comprehensive security audit (third-party)
- [ ] Formal release notes and changelog
- [ ] LTS (Long-Term Support) commitment

### Daemon & Background Operations
- [ ] `claudex daemon` service for Windows/macOS/Linux
- [ ] Systemd/launchd integration for auto-start
- [ ] IPC (inter-process communication) for CLI ↔ daemon
- [ ] Real-time rule updates without manual `claudex init`

### Multi-User & Organization Support
- [ ] Team/organization licensing
- [ ] Shared projects across team members
- [ ] Audit logging of all operations
- [ ] SAML/OIDC single sign-on (SSO)

### Advanced Rules Management
- [ ] Rules dependency graph and conflict detection
- [ ] Rollback to previous rule versions
- [ ] A/B testing support (multiple rule variants)
- [ ] Custom rule packaging and distribution

### CLI Enhancements
- [ ] Shell completion (bash, zsh, fish, PowerShell)
- [ ] Man page generation
- [ ] JSON output format for scripting
- [ ] Verbose/debug logging modes

### Documentation & Community
- [ ] Comprehensive user guide and tutorials
- [ ] API documentation for third-party integrations
- [ ] Community forum and support channels
- [ ] Video tutorials and webinars

---

## Roadmap Phases Summary

```
v0.1.0 (Apr 2026)  ┬─ Beta Launch: Core auth, rules distro, config, notifications
                   │
v0.1.1 (May 2026)  ├─ Polish: Security hardening, bug fixes, performance
                   │
v0.2.0 (Jul 2026)  ├─ Advanced: Auto-update, token refresh, signing, caching enhancements
                   │
v1.0.0 (Jan 2027)  └─ GA: Stability, daemon, org support, full feature set
```

## Priority Matrix

### High Priority (v0.1–v0.2)
- Security hardening and vulnerability fixes
- Auto-update mechanism (reduce manual deployment burden)
- Token refresh (extend session lifetimes)
- Bundle signing (cryptographic integrity verification)

### Medium Priority (v0.2–v1.0)
- CLI enhancements (shell completion, JSON output)
- Project management improvements
- Notification enhancements
- Configuration flexibility

### Low Priority (v1.0+)
- Daemon mode (background operations)
- Multi-tenant/org support
- P2P bundle sharing
- Advanced compliance features (audit logging, SSO)

## Success Metrics by Version

### v0.1.0
- 10,000+ downloads
- 99.5% API uptime
- < 5 minutes average rule installation time
- Zero critical security vulnerabilities

### v0.2.0
- 50,000+ downloads
- Auto-update adoption > 70%
- < 2 minutes average rule update time
- Support for 100+ enterprise customers

### v1.0.0
- 500,000+ downloads
- 99.9% API uptime (SLA)
- < 1 minute average rule installation time
- Trusted by Fortune 500 companies

---

## Dependencies & Blocking Items

| Phase | Dependency | Status |
|-------|-----------|--------|
| v0.1.0 | Go 1.23 toolchain | Ready |
| v0.1.0 | Cobra CLI framework | Ready |
| v0.1.0 | charmbracelet/huh TUI | Ready |
| v0.1.1 | UPX binary compression | Ready |
| v0.2.0 | Custom daemon implementation | Design in progress |
| v0.2.0 | GPG/Ed25519 signing library | Need research |
| v1.0.0 | SAML/OIDC provider integration | TBD |

---

## Stakeholder Feedback & Considerations

**Users (v0.1.0 Feedback Expected):**
- Feature request: Shell completion
- Feature request: JSON output for automation
- Issue: Cache cleanup on uninstall
- Issue: Clearer error messages for network failures

**Internal Teams:**
- API team: Need rate limiting guidelines
- Security team: Need security review milestone before v1.0
- Ops team: Need deployment automation (CI/CD scripts)

---

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|-----------|
| **Security vulnerability discovered** | Medium | Critical | Rapid patch release process; bug bounty program |
| **API performance degradation** | Low | High | Load testing; database optimization; caching |
| **Dependency vulnerability** | Medium | Medium | Automated dependency scanning; SCA tools |
| **User adoption slower than expected** | Low | Medium | Community outreach; documentation improvements |
| **Platform-specific bugs (macOS M1, Windows ARM)** | Medium | Medium | VM-based testing; hardware partnerships |

---

## Notes & Open Questions

- **v0.1.1 Security Audit:** Which third-party firm? Budget? Timeline?
- **v0.2.0 Daemon Implementation:** Should it use gRPC or simple HTTP server?
- **v1.0.0 Multi-Tenant:** Will licensing model change? Impact on pricing?
- **LTS Policy:** How many versions will we support simultaneously?
- **Backward Compatibility:** When can we break API contracts (major version only)?
