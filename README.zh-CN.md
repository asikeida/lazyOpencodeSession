# lazyOpencodeSession

[English](./README.md) | [简体中文](./README.zh-CN.md)

`lazyocs` 是一个快速的终端界面，用于浏览、搜索、预览和恢复本机 [OpenCode](https://opencode.ai/) 会话。

[![CI](https://github.com/asikeida/lazyOpencodeSession/actions/workflows/ci.yml/badge.svg)](https://github.com/asikeida/lazyOpencodeSession/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/asikeida/lazyOpencodeSession)](https://github.com/asikeida/lazyOpencodeSession/releases)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](./LICENSE)

> `lazyocs` 直接读取 OpenCode 的内部 SQLite 数据库。升级 OpenCode 后请先执行 `lazyocs --check`，对陌生 schema 启用写操作前请阅读 [COMPATIBILITY.md](./COMPATIBILITY.md)。

## 功能

- 按最近活动时间浏览根会话。
- 使用多关键词 AND 语义搜索会话元数据和近期用户消息。
- 无需加载完整消息 payload 即可预览最近提示词。
- 按 `Enter` 在 OpenCode 中恢复会话。
- 安全编辑标题，自动清洗单行输入并防止重复提交。
- 删除前分析完整影响范围、二次确认，并在事务中重新校验范围。
- 使用 `--read-only` 从 SQLite 数据源层进入真正的只读模式。
- 根据终端宽度自动切换单栏和双栏布局。
- 配置语言、详情字段、预览深度、边框、布局和主题。
- Linux 支持 Wayland/X11 剪贴板工具，macOS 支持 `pbcopy`，Windows 支持系统 `clip.exe`。
- 单个原生二进制运行，不依赖 CGO。

## 下载

所有正式产物统一发布到 [GitHub Releases](https://github.com/asikeida/lazyOpencodeSession/releases)。请使用随附的 `checksums.txt` 校验下载文件。

| 平台 | 架构 | 产物 |
| --- | --- | --- |
| Windows | x86-64 | `lazyocs_VERSION_windows_amd64.zip`，内含 `lazyocs.exe` |
| Windows | ARM64 | `lazyocs_VERSION_windows_arm64.zip`，内含 `lazyocs.exe` |
| macOS | Intel | `lazyocs_VERSION_darwin_amd64.tar.gz` |
| macOS | Apple 芯片 | `lazyocs_VERSION_darwin_arm64.tar.gz` |
| Linux | x86-64 / ARM64 | `.tar.gz`、`.deb`、`.rpm`、`.pkg.tar.zst` |

### Windows

下载对应 ZIP 并解压，在 PowerShell 或 Windows Terminal 中运行：

```powershell
.\lazyocs.exe --check
.\lazyocs.exe
```

发布产物是真正的原生 `.exe`，不需要安装 Go。需要全局调用时，把所在目录加入 `PATH`。Windows 构建目前属于实验支持且尚未签名；session ID 复制使用系统 `clip.exe`，失败时状态栏仍会显示可手动复制的 ID。

OpenCode 默认数据库位于 `%USERPROFILE%\.local\share\opencode\opencode.db`。如果你的安装位置不同，请传入 `--db`。

### macOS

Apple 芯片选择 `darwin_arm64`，Intel Mac 选择 `darwin_amd64`：

```bash
tar -xzf lazyocs_VERSION_darwin_arm64.tar.gz
./lazyocs --check
./lazyocs
```

当前 macOS 二进制尚未签名和公证，Gatekeeper 可能将其隔离。校验 `checksums.txt` 并确认信任下载来源后，可以移除隔离属性：

```bash
xattr -d com.apple.quarantine ./lazyocs
```

请保留二进制旁的 `themes/` 目录，或者把主题复制到 `~/.config/lazyocs/themes/`。要形成更成熟的 macOS 发布渠道，后续还需要 Apple Developer ID 签名、公证和 Homebrew tap。

### Linux 通用压缩包

```bash
tar -xzf lazyocs_VERSION_linux_amd64.tar.gz
./lazyocs --check
./lazyocs
```

### Debian 和 Ubuntu

```bash
pkexec apt install ./lazyocs_VERSION_amd64.deb
lazyocs --check
```

ARM64 设备请选择 `arm64.deb`。

### Fedora、RHEL 和 openSUSE

```bash
pkexec dnf install ./lazyocs_VERSION_amd64.rpm
lazyocs --check
```

没有 `dnf` 的系统请使用发行版对应的 RPM 包管理工具。

### Arch Linux

直接安装自动生成的 Arch Linux 包：

```bash
pkexec pacman -U ./lazyocs_VERSION_amd64.pkg.tar.zst
```

GoReleaser 还会生成 `lazyocs-bin` 的 AUR 元数据。在注册独立 AUR 包仓库并配置维护者 SSH 密钥之前，它不会自动上传。正式发布到 AUR 后，用户即可使用 `yay -S lazyocs-bin` 等方式安装。

## 从源码构建

当前模块要求 Go 1.25.6 或更高版本：

```bash
git clone https://github.com/asikeida/lazyOpencodeSession.git
cd lazyOpencodeSession
go build -o lazyocs ./cmd/lazyocs
./lazyocs --check
```

在 Linux 或 macOS 上交叉构建 Windows EXE：

```bash
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o lazyocs.exe ./cmd/lazyocs
```

交叉构建 macOS 二进制：

```bash
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o lazyocs-darwin-amd64 ./cmd/lazyocs
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o lazyocs-darwin-arm64 ./cmd/lazyocs
```

## 使用方法

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

命令行参数优先于配置文件。默认配置路径为 `~/.config/lazyocs/config.toml`，首次运行时自动创建，已有文件永远不会被覆盖。

最小配置示例：

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

执行 `lazyocs --print-config` 可查看带有完整中英文注释的配置。

## 快捷键

| 按键 | 操作 |
| --- | --- |
| `q`、`Ctrl+C` | 退出 |
| `j/k`、方向键 | 移动选择或滚动当前焦点面板 |
| `h/l` | 切换左右面板焦点 |
| `PgUp/PgDn`、`g/G` | 翻页或跳转到开头/结尾 |
| `/` | 搜索元数据和近期用户消息 |
| `Enter` | 恢复选中会话 |
| `e` | 修改标题 |
| `d` | 分析并确认删除 |
| `p` | 加载最近用户消息预览 |
| `y` | 复制 session ID |
| `r` | 重新加载会话 |
| `?` | 显示帮助 |

## 主题

内置预设包括 `lazygit-classic`、`moss`、`iris-night`、`tokyonight-storm`、`catppuccin-macchiato`、`nord-frost` 和三套 `retro-lime` 变体。

`theme_name` 按以下顺序查找：

1. 配置文件旁的 `themes/`。
2. 可执行文件旁的 `themes/`。
3. 系统包数据目录，Linux 通常为 `/usr/share/lazyocs/themes`。

内联 `[theme]` 会覆盖 `theme_name` 和 `theme_file`。颜色支持 ANSI 名称、256 色数字和十六进制值，样式支持 `bold`、`underline`、`faint`、`reverse`、`italic`。

## 安全与兼容性

- 读写模式使用 SQLite `mode=rw`，不会创建不存在的数据库。
- 只读模式从数据源层使用 SQLite `mode=ro`。
- TUI 启动前检查数据库 schema 能力。
- 缺少必要 schema 能力时，修改标题和删除操作默认拒绝执行。
- 删除确认框展示完整 session/message/part 影响范围。
- 删除事务内部会再次检查影响范围。
- 数据库锁定和超时错误会给出可执行的重试建议。
- 搜索和预览均有上限，避免无边界扫描大型 payload 表。

精确能力矩阵请查看 [COMPATIBILITY.md](./COMPATIBILITY.md)。

## 开发

```bash
go test ./...
go test -race ./...
go vet ./...
go test ./internal/opencode ./internal/tui -run '^$' -bench . -benchmem
```

架构和工程说明位于 [`opcode-summary/`](./opcode-summary/README.md)。

## 发布流程

推送语义化版本 tag 后，GitHub Actions 和 GoReleaser 会自动执行发布：

```bash
git tag -a v0.1.0 -m "v0.1.0"
git push origin v0.1.0
```

工作流会测试项目、构建所有目标、生成压缩包和 Linux 软件包、注入版本号，并把 SHA256 校验文件上传到 GitHub Releases。AUR、Homebrew、WinGet、Scoop、Microsoft 签名和 Apple 公证都需要额外的发布者账号、独立仓库或签名凭据，因此不会使用占位密钥假装启用。

## 许可证

[MIT](./LICENSE)
