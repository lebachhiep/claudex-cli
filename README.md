# ClaudeX CLI

> 🇻🇳 Tiếng Việt (mặc định) · [🇬🇧 English](README.en.md)

ClaudeX CLI là công cụ dòng lệnh để tải, cài đặt và quản lý bộ rules / skills / agents của Claude Code — license gắn theo máy, cache offline, đồng bộ đa dự án.

**Version:** 0.1.0 (Beta) · **License:** Proprietary · **Go:** 1.23+

---

## Yêu cầu

- Go 1.23+ đã cài đặt và `$GOPATH/bin` (hoặc `$HOME/go/bin`) nằm trong `PATH`.
- Một license key hợp lệ (mua hoặc xin từ đội ClaudeX).
- Kết nối Internet ở lần đầu cài rules (các lần sau dùng cache local).

---

## 1. Cài đặt

Cài trực tiếp từ Go:

```bash
go install github.com/lebachhiep/claudex-cli/cmd/claudex@latest
```

Lệnh này build và đặt binary `claudex` vào `$GOPATH/bin`. Đảm bảo thư mục đó nằm trong `PATH`:

```bash
# macOS / Linux
export PATH="$PATH:$(go env GOPATH)/bin"

# Windows (PowerShell)
$env:Path += ";$(go env GOPATH)\bin"
```

Kiểm tra:

```bash
claudex version
```

---

## 2. Đăng nhập

Đăng nhập để gắn license vào máy hiện tại. Bind theo hardware fingerprint (OS + hostname + arch), chặn chia sẻ license sai mục đích, nhưng vẫn cho phép đăng nhập trên nhiều máy nằm trong hạn mức gói.

```bash
claudex login --key=YOUR_LICENSE_KEY
```

Đăng xuất / gỡ bind máy hiện tại:

```bash
claudex logout
```

---

## 3. Khởi tạo dự án

Vào thư mục dự án cần cài rules, chạy `claudex init`. Lệnh này tải bản rules mới nhất, giải nén vào `.claude/`, và tạo `.claudex.lock` để theo dõi phiên bản đã cài.

```bash
cd your-project
claudex init
```

Cài một phiên bản cụ thể thay vì bản mới nhất:

```bash
claudex init 1.0.5
```

---

## 4. Bắt đầu

Sau khi init xong, dự án của bạn đã có đầy đủ skills, agents, hooks trong `.claude/`. Mở Claude Code lên là dùng được. Chạy `claudex status` để xem license, machine, và các dự án đã cài rules.

```bash
claudex status
```

Xem danh sách `CLI reference` bên dưới để biết các lệnh khác.

---

## CLI reference

Tất cả các lệnh CLI có thể chạy ở terminal (không phải trong Claude Code).

| Lệnh | Mục đích |
|------|----------|
| `claudex login --key=<KEY>` | Đăng nhập, gắn máy với license |
| `claudex logout` | Đăng xuất, gỡ bind máy hiện tại |
| `claudex init [version]` | Cài / cập nhật rules vào dự án hiện tại |
| `claudex update` | Kiểm tra rules mới và cập nhật tất cả dự án đã bind |
| `claudex status` | Xem trạng thái license, máy, rules của các dự án |
| `claudex versions [current]` | Liệt kê các phiên bản rules có sẵn |
| `claudex projects` | Liệt kê các dự án đã cài rules, sync lại config |
| `claudex version` | Xem version, commit, ngày build của CLI |
| `claudex config` | Mở menu cấu hình (ngôn ngữ, coding level, notification, context7) |
| `claudex config language` | Đổi ngôn ngữ CLI (English / Tiếng Việt) |
| `claudex config coding-level` | Đặt coding level (0–3) để điều chỉnh mức giải thích của AI |
| `claudex config notification` | Cấu hình thông báo Telegram / Discord / Slack |
| `claudex config context7` | Cấu hình Context7 API key cho docs-seeker |

---

## Cấu hình

### Thư mục dữ liệu

```
~/.claudex/
├── auth.json        # Session token + machine ID (đã mã hóa)
├── config.json      # Cấu hình notification, language, coding level
├── projects.json    # Danh sách dự án đã cài rules
└── cache/           # Cache các bundle rules đã tải
```

File nhạy cảm dùng quyền `0600`.

---

## Tính năng chính

- **Machine-specific auth** — license gắn theo hardware fingerprint, chặn share key lung tung, vẫn support nhiều máy trong hạn mức gói.
- **Offline install** — bundle rules đã tải được cache lại, lần sau cài cùng phiên bản không cần mạng.
- **Lock file** — `.claudex.lock` ghi version, plan, checksum, ngày cài → biết chính xác dự án đang ở phiên bản nào.
- **Multi-project sync** — `claudex update` quét mọi dự án đã bind và cập nhật đồng loạt.
- **Notifications** — Telegram, Discord, hoặc Slack.
- **i18n** — giao diện CLI English / Tiếng Việt, đổi qua `claudex config language`.

---

## Bảo mật

- **Device binding** — machine ID từ hardware, chống nhân bản license.
- **Encrypted storage** — session token lưu với permission 0600.
- **HTTPS-only** — toàn bộ giao tiếp API + tải bundle qua HTTPS.
- **Binary obfuscation** — bản production build qua `garble`, khó reverse.
- **Integrity check** — bundle verify SHA-256 sau khi tải.

---

## Troubleshooting

### "Not logged in"

```bash
claudex login --key=YOUR_LICENSE_KEY
```

### "Device limit exceeded"

Đã đạt giới hạn số máy cho license. Gỡ máy cũ:

```bash
claudex logout
```

Rồi đăng nhập ở máy mới.

### "Cannot reach API server"

Check mạng + firewall, thử lại sau ít phút.

### "Bundle integrity check failed"

Checksum không khớp — có thể mạng lỗi hoặc bundle bị tamper. Thử lại:

```bash
claudex init --force
```

---

## Phát triển

### Yêu cầu

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

Chi tiết kiến trúc xem [System Architecture](docs/system-architecture.md).

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
- **Docs:** thư mục [docs/](docs/)
- **Email:** support@claudex.dev

---

## License

Proprietary. Xem file `LICENSE`.
