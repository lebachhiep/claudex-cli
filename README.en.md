# ClaudeX CLI

> [🇻🇳 Tiếng Việt (default)](README.md) · 🇬🇧 English

ClaudeX CLI is a command-line tool to download, install, and manage Claude Code rules / skills / agents — machine-bound licensing, offline cache, multi-project sync.

**Version:** 0.1.0 (Beta) · **License:** Proprietary · **Go:** 1.23+

---

## Requirements

- Go 1.23+ installed, with `$GOPATH/bin` (or `$HOME/go/bin`) in your `PATH`.
- A valid license key (purchased or provided by the ClaudeX team).
- Internet connection for the first rules install (subsequent installs use the local cache).

---

## 1. Install

Install directly via Go:

```bash
go install github.com/lebachhiep/claudex-cli/cmd/claudex@latest
```

This builds and places the `claudex` binary in `$GOPATH/bin`. Make sure that directory is in your `PATH`:

```bash
# macOS / Linux
export PATH="$PATH:$(go env GOPATH)/bin"

# Windows (PowerShell)
$env:Path += ";$(go env GOPATH)\bin"
```

Verify:

```bash
claudex version
```

---

## 2. Log in

Log in to bind the license to the current machine. Binding uses a hardware fingerprint (OS + hostname + arch) — it prevents license sharing but still allows multiple machines within your plan's device limit.

```bash
claudex login --key=YOUR_LICENSE_KEY
```

Log out / unbind this machine:

```bash
claudex logout
```

---

## 3. Initialize a project

In the project where you want to install rules, run `claudex init`. This downloads the latest rules bundle, extracts it into `.claude/`, and writes a `.claudex.lock` file to track the installed version.

```bash
cd your-project
claudex init
```

Install a specific version instead of the latest:

```bash
claudex init 1.0.5
```

---

## 4. Get started

After `init`, the project has the full set of skills, agents, and hooks in `.claude/`. Open Claude Code and you're ready to go. Run `claudex status` to inspect your license, machine binding, and installed projects.

```bash
claudex status
```

See the `CLI reference` below for the rest of the commands.

---

## CLI reference

All CLI commands run in your terminal (not inside Claude Code).

| Command | Purpose |
|---------|---------|
| `claudex login --key=<KEY>` | Log in, bind this machine to the license |
| `claudex logout` | Log out, unbind this machine |
| `claudex init [version]` | Install / update rules in the current project |
| `claudex update` | Check for new rules and update all bound projects |
| `claudex status` | Show license, machine, and rules status for projects |
| `claudex versions [current]` | List available rules versions |
| `claudex projects` | List all installed projects, resync their config |
| `claudex version` | Show CLI version, commit, build date |
| `claudex config` | Open the config menu (language, coding level, notification, context7) |
| `claudex config language` | Switch CLI language (English / Tiếng Việt) |
| `claudex config coding-level` | Set coding level (0–3) to tune AI explanation depth |
| `claudex config notification` | Configure Telegram / Discord / Slack notifications |
| `claudex config context7` | Configure Context7 API key for docs-seeker |

---

## Configuration

### Environment variables

```bash
# Custom API server (default: https://api-dev.claudex.info)
export CLAUDEX_SERVER=https://custom-api.example.com

# Custom data directory (default: ~/.claudex/)
export CLAUDEX_DATA_DIR=/var/claudex
```

### Data directory

```
~/.claudex/
├── auth.json        # Session token + machine ID (encrypted)
├── config.json      # Notification, language, coding level config
├── projects.json    # Registry of installed projects
└── cache/           # Downloaded rules bundle cache
```

Sensitive files use `0600` permissions.

---

## Features

- **Machine-specific auth** — licenses are bound to a hardware fingerprint, preventing uncontrolled sharing while still supporting multiple devices within your plan.
- **Offline install** — downloaded bundles are cached locally; reinstalling the same version doesn't need network.
- **Lock file** — `.claudex.lock` records version, plan, checksum, and install date, so you always know which version a project is on.
- **Multi-project sync** — `claudex update` scans every bound project and updates them in one pass.
- **Notifications** — Telegram, Discord, or Slack.
- **i18n** — CLI UI available in English / Tiếng Việt, switchable via `claudex config language`.

---

## Security

- **Device binding** — hardware-derived machine ID prevents license duplication.
- **Encrypted storage** — session tokens stored with `0600` permissions.
- **HTTPS-only** — all API traffic and bundle downloads go over HTTPS.
- **Binary obfuscation** — production builds are obfuscated with `garble` to resist reverse engineering.
- **Integrity check** — bundles are verified with SHA-256 after download.

---

## Troubleshooting

### "Not logged in"

```bash
claudex login --key=YOUR_LICENSE_KEY
```

### "Device limit exceeded"

You've hit the device limit for this license. Unbind an older machine:

```bash
claudex logout
```

Then log in on the new machine.

### "Cannot reach API server"

Check your network + firewall. If you're using a custom server:

```bash
CLAUDEX_SERVER=https://your-server claudex status
```

### "Bundle integrity check failed"

Checksum mismatch — network corruption or tampering. Retry:

```bash
claudex init --force
```

---

## Development

### Prerequisites

- Go 1.23+
- Make
- git

### Clone & build

```bash
git clone https://github.com/lebachhiep/claudex-cli.git
cd claudex-cli
go mod download

# Dev build
make build
./dist/claudex version

# Production (obfuscated)
make build-prod

# Cross-platform
make build-all
```

### Test & lint

```bash
make test
make lint
```

---

## Project structure

```
claudex-cli/
├── cmd/claudex/          # CLI entry point
├── internal/
│   ├── api/              # HTTP client + API types
│   ├── auth/             # Auth + device binding
│   ├── cli/              # Command implementations
│   ├── config/           # Config management
│   ├── crypto/           # Encryption + key derivation
│   ├── i18n/             # CLI translations
│   ├── notification/     # Notification providers
│   ├── projects/         # Project registry
│   └── rules/            # Rules download / extract / cache
├── Makefile
├── go.mod
└── docs/
```

See [System Architecture](docs/system-architecture.md) for details.

---

## Documentation

- [Project Overview & PDR](docs/project-overview-pdr.md)
- [Codebase Summary](docs/codebase-summary.md)
- [Code Standards](docs/code-standards.md)
- [System Architecture](docs/system-architecture.md)
- [Project Roadmap](docs/project-roadmap.md)

---

## Support

- **GitHub Issues:** [claudex-cli/issues](https://github.com/lebachhiep/claudex-cli/issues)
- **Docs:** the [docs/](docs/) directory
- **Email:** support@claudex.dev

---

## License

Proprietary. See the `LICENSE` file.
