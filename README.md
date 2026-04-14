# ClaudeX CLI

ClaudeX CLI is a secure, cross-platform command-line tool for managing Claude Code skills distribution and license binding. It enables users to authenticate with machine-specific device binding, download rules bundles, and track project installations.

**Version:** 0.1.0 (Beta)  
**License:** Proprietary  
**Go Version:** 1.23+

## Quick Start

### Installation

Download the pre-built binary for your platform from the [releases page](https://github.com/claudex/claudex-cli/releases):

- **Windows:** `claudex-windows-amd64.exe`
- **macOS (Intel):** `claudex-darwin-amd64`
- **macOS (Apple Silicon):** `claudex-darwin-arm64`
- **Linux:** `claudex-linux-amd64`

Extract the binary and add it to your PATH:

```bash
# macOS/Linux
chmod +x claudex
sudo mv claudex /usr/local/bin/

# Windows
# Copy claudex-windows-amd64.exe to a directory in PATH (e.g., C:\Program Files\claudex\)
```

### Authenticate

```bash
claudex login --key=YOUR_LICENSE_KEY
```

This binds your machine (hardware-specific) to the license. You can authenticate on up to the limit specified in your license tier.

### Install Rules

```bash
cd /path/to/your/project
claudex init
```

This downloads and installs the latest rules bundle, creating a `.claudex.lock` file to track the installation.

### Check Status

```bash
claudex status
```

Displays authentication status, plan tier, and installed projects.

## Commands

| Command | Purpose |
|---------|---------|
| `claudex login --key=<KEY>` | Authenticate with license key and bind machine |
| `claudex logout` | Unbind machine and remove authentication |
| `claudex init [version]` | Install or update rules to latest or specified version |
| `claudex update` | Deprecated; use `init` instead |
| `claudex status` | Check authentication and project status |
| `claudex versions` | List available rule bundle versions |
| `claudex version` | Show CLI version and build info |
| `claudex notification` | Interactive setup for notifications (Telegram/Discord/Slack) |

## Configuration

### Environment Variables

Override defaults via environment variables:

```bash
# Custom API server (default: https://api-dev.claudex.info)
export CLAUDEX_SERVER=https://custom-api.example.com

# Custom data directory (default: ~/.claudex/)
export CLAUDEX_DATA_DIR=/var/claudex
```

### Data Directory

ClaudeX stores configuration and cached data in `~/.claudex/`:

```
~/.claudex/
├── auth.json        # Encrypted session token and machine ID
├── config.json      # Notification provider configuration
├── projects.json    # Registry of installed projects
└── cache/           # Downloaded rule bundle cache
```

All configuration files are created with secure permissions (0600 for sensitive files).

## Features

### Machine-Specific Authentication

Your license is bound to your machine using hardware fingerprinting (OS, hostname, architecture). This prevents license sharing across devices while allowing multi-device support within your license limits.

### Offline Installation

Rule bundles are cached locally after download. Reinstalling the same version uses the cache, eliminating the need for network connectivity on subsequent installations.

### Rules Management

Track installed projects and versions via `.claudex.lock`:

```json
{
  "version": "1.0.5",
  "plan": "pro",
  "installed_at": "2026-04-13T10:30:00Z",
  "checksum": "sha256:abc123...",
  "cli_version": "0.1.0"
}
```

### Notifications

Configure alerts via Telegram, Discord, or Slack:

```bash
claudex notification
```

Choose your provider and enter credentials. Credentials are stored securely in `~/.claudex/config.json` (0600 permissions).

## Security

- **Device Binding:** Hardware-based machine ID prevents license duplication
- **Encrypted Storage:** Session tokens stored with secure file permissions (0600)
- **HTTPS-Only:** All API communication and bundle downloads over HTTPS
- **Binary Obfuscation:** Production binaries obfuscated with garble to resist reverse engineering
- **Integrity Checking:** Downloaded bundles verified via SHA-256 checksums

## Development

### Prerequisites

- Go 1.23 or later
- Make
- git

### Clone & Setup

```bash
git clone https://github.com/claudex/claudex-cli.git
cd claudex-cli
go mod download
```

### Build

```bash
# Development build (fast, symbols included)
make build
./dist/claudex version

# Production build (obfuscated)
make build-prod

# Cross-platform builds
make build-all
```

### Test

```bash
make test
# or
go test ./...
```

### Lint

```bash
make lint
# or
golangci-lint run ./...
```

## Project Structure

```
claudex-cli/
├── cmd/claudex/          # CLI entry point
├── internal/
│   ├── api/              # HTTP client and API types
│   ├── auth/             # Authentication and device binding
│   ├── cli/              # Command implementations
│   ├── config/           # Configuration management
│   ├── crypto/           # Encryption and key derivation
│   ├── notification/     # Notification provider setup
│   ├── projects/         # Project registry
│   └── rules/            # Rules download, extraction, and caching
├── Makefile              # Build targets
├── go.mod               # Go module definition
└── docs/                # Documentation
```

For detailed architecture and design information, see [System Architecture](docs/system-architecture.md).

## Documentation

- **[Project Overview & PDR](docs/project-overview-pdr.md)** — Vision, requirements, success metrics
- **[Codebase Summary](docs/codebase-summary.md)** — Directory structure, packages, testing
- **[Code Standards](docs/code-standards.md)** — Naming conventions, patterns, guidelines
- **[System Architecture](docs/system-architecture.md)** — Design decisions, data flows, security model
- **[Project Roadmap](docs/project-roadmap.md)** — Version plans, features, timeline

## Troubleshooting

### "Not logged in" Error

```bash
claudex login --key=YOUR_LICENSE_KEY
```

### "Device limit exceeded"

You've bound this license to the maximum number of devices. Unbind another device:

```bash
claudex logout
```

Then authenticate on a different machine.

### "Cannot reach API server"

Check your internet connection and firewall settings. If using a custom server:

```bash
CLAUDEX_SERVER=https://your-server claudex status
```

### "Bundle integrity check failed"

The downloaded bundle's checksum doesn't match. This may indicate network corruption or tampering. Try again:

```bash
claudex init --force
```

## Contributing

Contributions are welcome. Please ensure:

1. Code follows the standards in [Code Standards](docs/code-standards.md)
2. Tests pass: `make test`
3. Linting passes: `make lint`
4. Commit messages follow conventional commits format

## Support

For issues, questions, or feature requests:

- **GitHub Issues:** [claudex-cli/issues](https://github.com/claudex/claudex-cli/issues)
- **Documentation:** See [docs/](docs/) directory
- **Email:** support@claudex.dev

## License

Proprietary. See LICENSE file for details.

## Changelog

### v0.1.0 (April 2026)

**Initial Release**

- Core authentication with machine-specific device binding
- Rules bundle download, validation, and extraction
- Local caching for offline installations
- `.claudex.lock` file tracking
- Configuration management (`~/.claudex/`)
- Notification providers (Telegram, Discord, Slack)
- Cross-platform support (Windows, macOS, Linux)
- Binary obfuscation for production builds

For detailed release notes, see [Project Roadmap](docs/project-roadmap.md).
