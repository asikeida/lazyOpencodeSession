# lazyOpencodeSession

[English](./README.md) | [简体中文](./README.zh-CN.md)

`lazyocs` is a fast terminal UI for browsing, searching, previewing, and resuming local [OpenCode](https://opencode.ai/) sessions.

[![CI](https://github.com/asikeida/lazyOpencodeSession/actions/workflows/ci.yml/badge.svg)](https://github.com/asikeida/lazyOpencodeSession/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/asikeida/lazyOpencodeSession)](https://github.com/asikeida/lazyOpencodeSession/releases)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](./LICENSE)

> `lazyocs` reads OpenCode's internal SQLite database. Run `lazyocs --check` after upgrading OpenCode, and review [COMPATIBILITY.md](./COMPATIBILITY.md) before enabling write operations on an unfamiliar schema.

## Features

- Browse root sessions ordered by recent activity.
- Search session metadata and recent user messages with multi-term AND matching.
- Preview recent prompts without loading entire message payloads.
- Resume a session in OpenCode with `Enter`.
- Edit titles with single-line input sanitization and duplicate-submit protection.
- Delete a session tree only after impact analysis, confirmation, and transactional scope revalidation.
- Use real SQLite read-only mode with `--read-only`.
- Switch between responsive one-pane and two-pane layouts.
- Configure language, fields, preview depth, borders, layout, and themes.
- Copy session IDs through Wayland/X11 tools on Linux, `pbcopy` on macOS, or `clip.exe` on Windows.
- Run as a single native binary without CGO.

## Downloads

All official artifacts are published on [GitHub Releases](https://github.com/asikeida/lazyOpencodeSession/releases). Verify downloads with the accompanying `checksums.txt`.

| Platform | Architecture | Artifact |
| --- | --- | --- |
| Windows | x86-64 | `lazyocs_VERSION_windows_amd64.zip` containing `lazyocs.exe` |
| Windows | ARM64 | `lazyocs_VERSION_windows_arm64.zip` containing `lazyocs.exe` |
| macOS | Intel | `lazyocs_VERSION_darwin_amd64.tar.gz` |
| macOS | Apple Silicon | `lazyocs_VERSION_darwin_arm64.tar.gz` |
| Linux | x86-64 / ARM64 | `.tar.gz`, `.deb`, `.rpm`, and `.pkg.tar.zst` |

### Windows

Download and extract the matching ZIP, then run it from PowerShell or Windows Terminal:

```powershell
.\lazyocs.exe --check
.\lazyocs.exe
```

The release is a real native `.exe`; no Go installation is required. Add its directory to `PATH` if you want to call `lazyocs` globally. Windows builds are currently experimental and unsigned. Session ID copying uses the system `clip.exe`; failures still show the ID for manual copying.

OpenCode stores its database at `%USERPROFILE%\.local\share\opencode\opencode.db`. Pass `--db` if your installation uses another path.

### macOS

Choose `darwin_arm64` for Apple Silicon or `darwin_amd64` for Intel:

```bash
tar -xzf lazyocs_VERSION_darwin_arm64.tar.gz
./lazyocs --check
./lazyocs
```

The current macOS binary is unsigned and not notarized. Gatekeeper may quarantine it. After verifying `checksums.txt` and only if you trust the download, remove the quarantine attribute:

```bash
xattr -d com.apple.quarantine ./lazyocs
```

Keep the extracted `themes/` directory beside the binary, or copy those files to `~/.config/lazyocs/themes/`. A polished public macOS channel should later add Apple Developer ID signing, notarization, and a Homebrew tap.

### Linux Archive

```bash
tar -xzf lazyocs_VERSION_linux_amd64.tar.gz
./lazyocs --check
./lazyocs
```

### Debian and Ubuntu

```bash
pkexec apt install ./lazyocs_VERSION_amd64.deb
lazyocs --check
```

Use the `arm64.deb` artifact on ARM64 systems.

### Fedora, RHEL, and openSUSE

```bash
pkexec dnf install ./lazyocs_VERSION_amd64.rpm
lazyocs --check
```

On systems without `dnf`, install the RPM with the distribution's normal package tool.

### Arch Linux

Install the generated package directly:

```bash
pkexec pacman -U ./lazyocs_VERSION_amd64.pkg.tar.zst
```

GoReleaser also generates `lazyocs-bin` AUR metadata. It is not uploaded automatically until an AUR package repository and maintainer SSH key are configured. Once published, users will be able to install it with an AUR helper such as `yay -S lazyocs-bin`.

## Build From Source

Go 1.25.6 or newer is required by the current module:

```bash
git clone https://github.com/asikeida/lazyOpencodeSession.git
cd lazyOpencodeSession
go build -o lazyocs ./cmd/lazyocs
./lazyocs --check
```

Cross-compile a Windows executable from Linux or macOS:

```bash
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o lazyocs.exe ./cmd/lazyocs
```

Cross-compile macOS binaries:

```bash
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o lazyocs-darwin-amd64 ./cmd/lazyocs
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o lazyocs-darwin-arm64 ./cmd/lazyocs
```

## Usage

```bash
lazyocs
lazyocs --check --limit 5
lazyocs --read-only
lazyocs --language zh-CN
lazyocs --db ~/.local/share/opencode/opencode.db
lazyocs --config ~/lazyocs.toml
lazyocs --print-config
lazyocs --version
```

CLI flags override config file values. The default config is `~/.config/lazyocs/config.toml`; it is created on first run and never overwritten.

Minimal configuration:

```toml
db = ""
language = "auto"
limit = 500
opencode = "opencode"
read_only = false
theme_name = "lazygit-classic"
theme_file = ""

[search]
recent_days = 7

[preview]
recent_messages_limit = 5

[ui]
border_style = "rounded"
split_ratio = 0.45
two_pane_min_width = 110

[details.fields]
title = true
session = true
project = true
directory = true
path_status = true
message_count = false
part_count = false
size = false
large_session = false
updated = true
created = true
model = true
agent = true
cost = true
tokens = true
resume_command = false
```

Run `lazyocs --print-config` for the fully commented bilingual configuration.

## Key Bindings

| Key | Action |
| --- | --- |
| `q`, `Ctrl+C` | Quit |
| `j/k`, arrows | Move selection or scroll the focused pane |
| `h/l` | Switch pane focus |
| `PgUp/PgDn`, `g/G` | Page or jump in the active view |
| `/` | Search metadata and recent user messages |
| `Enter` | Resume selected session |
| `e` | Edit title |
| `d` | Analyze and confirm deletion |
| `p` | Load recent user-message preview |
| `y` | Copy session ID |
| `r` | Reload sessions |
| `?` | Show help |

## Themes

Bundled presets include `lazygit-classic`, `moss`, `iris-night`, `tokyonight-storm`, `catppuccin-macchiato`, `nord-frost`, and three `retro-lime` variants.

`theme_name` is resolved from:

1. `themes/` beside the config file.
2. `themes/` beside the executable.
3. The package data directory, normally `/usr/share/lazyocs/themes` on Linux.

Inline `[theme]` values override both `theme_name` and `theme_file`. Colors accept ANSI names, 256-color numbers, and hex values. Styles accept `bold`, `underline`, `faint`, `reverse`, and `italic`.

## Safety and Compatibility

- Read-write mode uses SQLite `mode=rw` and never creates a missing database.
- Read-only mode uses SQLite `mode=ro` at the data-source level.
- Schema capabilities are inspected before the TUI starts.
- Rename and delete fail closed when required schema capabilities are missing.
- Delete confirmation shows the complete session/message/part scope.
- The scope is checked again inside the delete transaction.
- Database lock and timeout errors include a retry action.
- Search and previews are bounded to avoid scanning an entire large payload table.

See [COMPATIBILITY.md](./COMPATIBILITY.md) for the exact capability matrix.

## Development

```bash
go test ./...
go test -race ./...
go vet ./...
go test ./internal/opencode ./internal/tui -run '^$' -bench . -benchmem
```

Technical architecture and engineering notes are in [`opcode-summary/`](./opcode-summary/README.md).

## Release Process

Pushing a semantic-version tag triggers GitHub Actions and GoReleaser:

```bash
git tag -a v0.1.0 -m "v0.1.0"
git push origin v0.1.0
```

The workflow tests the project, builds all release targets, creates archives and Linux packages, injects the version, and publishes SHA256 checksums to GitHub Releases. AUR, Homebrew, WinGet, Scoop, Microsoft signing, and Apple notarization require separate publisher accounts, repositories, or signing credentials and are intentionally not enabled with placeholder secrets.

## License

[MIT](./LICENSE)
