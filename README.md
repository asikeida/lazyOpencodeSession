# lazyOpencodeSession

`lazyOpencodeSession` is a fast terminal UI for browsing, searching, and resuming OpenCode sessions.

The command name is `lazyocs`.

## Technical Documentation

For an architecture-level walkthrough, implementation details, code reading order, and production roadmap, see [`opcode-summary/README.md`](./opcode-summary/README.md). The documentation covers the Bubble Tea state machine, SQLite safety and performance, configuration, testing, and release engineering.

## MVP Features

- Read OpenCode SQLite database with title editing enabled by default.
- List root OpenCode sessions by recent update time.
- Search session metadata and recent user messages with `/`.
- Multi-term search with AND semantics, for example `windows iso` matches sessions containing both words.
- Show session details without scanning large message payloads.
- Show whether the session directory still exists.
- Lazy-load recent user message preview with `p`.
- Resume selected session with `Enter`.
- Edit the selected session title with `e`.
- Delete the selected session with `d` after confirmation.
- Copy selected session id with `y`.
- Uses a terminal-friendly default color scheme inspired by lazygit.
- Chinese UI with `--language auto|en|zh-CN`.
- Auto-create the default config file at `~/.config/lazyocs/config.toml` on first run.
- Configurable details fields.

## Build

```bash
go build -o lazyocs ./cmd/lazyocs
```

## Run

```bash
./lazyocs
```

Use a custom OpenCode database:

```bash
./lazyocs --db ~/.local/share/opencode/opencode.db
```

Use Chinese UI:

```bash
./lazyocs --language zh-CN
```

`auto` is the default. It uses Chinese when your locale contains `zh`, `cn`, or `CN`.

Check database access without launching the TUI:

```bash
./lazyocs --check --limit 5
```

Use read-only mode when needed:

```bash
./lazyocs --read-only
```

Alternatively set `read_only = true` in `~/.config/lazyocs/config.toml`.

Print a sample config:

```bash
./lazyocs --print-config
```

Use a custom config file:

```bash
./lazyocs --config ~/lazyocs.toml
```

CLI flags override config file values. Custom config paths passed with `--config` must already exist.

## Config

Default config path:

```text
~/.config/lazyocs/config.toml
```

If this file does not exist, `lazyocs` creates it with the default sample content on first run. Existing config files are never overwritten.

Sample config:

```toml
# OpenCode SQLite database path. Default: ~/.local/share/opencode/opencode.db.
# Leave empty to use the default path.
# OpenCode SQLite 数据库路径。默认值：~/.local/share/opencode/opencode.db。
# 留空表示使用默认路径。
db = ""
# UI language: auto, en, or zh-CN. Default: auto.
# 界面语言：auto、en 或 zh-CN。默认值：auto。
language = "auto"
# Maximum number of sessions loaded into the list. Default: 500.
# 会话列表最多加载的数量。默认值：500。
limit = 500
# OpenCode executable or command path used by Enter. Default: opencode.
# 按 Enter 恢复会话时使用的 OpenCode 命令或路径。默认值：opencode。
opencode = "opencode"
# Database write protection. Default: false.
# 数据库写保护。默认值：false。使用 --read-only 或设为 true 可禁止修改标题。
read_only = false

# Optional built-in or local theme preset name. Example: lazygit-classic, moss, iris-night.
# 可选的内置或本地主题预设名称，例如 lazygit-classic、moss、iris-night。
theme_name = ""

# Optional external theme file. Relative paths are resolved from this config file.
# 可选的外部主题文件。相对路径将基于当前配置文件解析。
theme_file = ""

[search]
# User-message memory window in days. Use 0 to disable. Default: 7.
# 用户消息记忆搜索的时间范围（天）。设为 0 可关闭。默认值：7。
recent_days = 7

[preview]
# Number of recent user messages shown in the right-side preview. Default: 5.
# 右侧预览显示的最近用户消息条数。默认值：5。
recent_messages_limit = 5

[ui]
# Panel border style: rounded, single, double, hidden, or bold. Default: rounded.
# 面板边框样式：rounded、single、double、hidden 或 bold。默认值：rounded。
border_style = "rounded"
# Left panel width ratio in two-pane mode. Default: 0.45.
# 双栏模式下左侧面板宽度比例。默认值：0.45。
split_ratio = 0.45
# Minimum terminal width for two-pane mode. Default: 110.
# 进入双栏模式所需的最小终端宽度。默认值：110。
two_pane_min_width = 110

# Theme configuration. Uncomment and adjust any values you want to override.
# 主题配置。只需取消注释并修改你想覆盖的值。
#
# [theme]
# active_border_color = ["green", "bold"]
# inactive_border_color = ["green"]
# searching_active_border_color = ["cyan", "bold"]
# title_color = ["green", "bold"]
# accent_color = ["green", "bold"]
# options_text_color = ["blue"]
# default_fg_color = ["default"]
# muted_fg_color = ["250"]
# match_color = ["cyan", "bold", "underline"]
# focused_selected_match_color = ["229", "underline", "bold"]
# inactive_selected_match_color = ["cyan", "underline"]
# error_color = ["red", "bold"]
# warning_color = ["yellow"]
# selected_line_bg_color = ["blue"]
# selected_line_fg_color = ["white", "bold"]
# inactive_view_selected_line_color = ["bold"]
# status_mode_color = ["252"]
# status_text_color = ["252"]
# status_key_color = ["blue"]
# details_hint_color = ["blue"]
# preview_timestamp_color = ["250"]
# memory_snippet_color = ["250"]
#
# [theme.modal]
# background_color = "#202330"
# border_color = "#A8B47A"
# key_color = "#69AFC1"
# text_color = "#B9C2D0"
# muted_color = "#778195"
# icon_color = "#6096A3"
# warning_color = "#B5A06D"

[details.fields]
# Session title. Default: true.
# 会话标题。默认值：true。
title = true
# Session ID. Default: true.
# 会话 ID。默认值：true。
session = true
# Project ID. Default: true.
# 项目 ID。默认值：true。
project = true
# Session working directory. Default: true.
# 会话工作目录。默认值：true。
directory = true
# Whether the session directory currently exists. Default: true.
# 会话目录当前是否存在。默认值：true。
path_status = true
# Number of messages. Loaded lazily when enabled. Default: false.
# 消息数量。启用后按需加载。默认值：false。
message_count = false
# Number of parts. Loaded lazily when enabled. Default: false.
# 内容片段数量。启用后按需加载。默认值：false。
part_count = false
# Combined message and part payload size. Loaded lazily. Default: false.
# 消息和片段数据总大小。启用后按需加载。默认值：false。
size = false
# Whether payload size is at least 10 MB. Loaded lazily. Default: false.
# 数据总大小是否达到 10 MB。启用后按需加载。默认值：false。
large_session = false
# Last update time. Default: true.
# 最后更新时间。默认值：true。
updated = true
# Creation time. Default: true.
# 创建时间。默认值：true。
created = true
# Model information. Default: true.
# 模型信息。默认值：true。
model = true
# Agent name. Default: true.
# Agent 名称。默认值：true。
agent = true
# Session cost. Default: true.
# 会话费用。默认值：true。
cost = true
# Token usage. Default: true.
# Token 使用情况。默认值：true。
tokens = true
# Command used to resume the session. Default: false.
# 恢复会话时使用的命令。默认值：false。
resume_command = false
```

Use `true` to show a field and `false` to hide it. `1` and `0` are also accepted.

The default color scheme follows lazygit's general terminal style: green accent/border, blue selected line, blue options text, and terminal-provided background. These defaults can now be overridden through `[theme]` and `[theme.modal]` in the config file.

Theme overrides live under `[theme]` and `[theme.modal]`. Color entries accept named ANSI colors such as `green`, numeric 256-color strings such as `250`, and hex values such as `#A8B47A`. Style arrays can also include `bold`, `underline`, `faint`, `reverse`, and `italic`.

Layout and frame behaviour live under `[ui]`. Use this section for structural choices like border style, split ratio, and the minimum width required for two-pane mode.

If you want to switch among known presets quickly, set `theme_name`. lazyocs first looks in `~/.config/lazyocs/themes/`, then in `themes/` next to the executable. If you want to share or distribute a theme separately from your main config, set `theme_file` and move the whole `[theme]` block into another TOML file. Inline `[theme]` values still override both preset and file.

Bundled presets:

- `themes/lazygit-classic.toml`
- `themes/moss.toml`
- `themes/iris-night.toml`
- `themes/tokyonight-storm.toml`
- `themes/catppuccin-macchiato.toml`
- `themes/nord-frost.toml`
- `themes/retro-lime.toml`
- `themes/retro-lime-soft.toml`
- `themes/retro-lime-neon.toml`

Example:

```toml
theme_name = "moss"

[theme]
active_border_color = ["yellow", "bold"]
```

Or with an explicit file:

```toml
theme_file = "themes/moss.toml"

[theme]
active_border_color = ["yellow", "bold"]
```

The current details panel supports these fields:

- `title`
- `session`
- `project`
- `directory`
- `path_status`
- `message_count`
- `part_count`
- `size`
- `large_session`
- `updated`
- `created`
- `model`
- `agent`
- `cost`
- `tokens`
- `resume_command`

Statistics fields are loaded only for the selected session when one of them is enabled. `large_session` means the session message and part payloads are at least 10 MB. The right-side preview uses `preview.recent_messages_limit`; increasing it shows more user messages but also increases each preview query and render cost.

## Key Bindings

```text
q / Ctrl+C     Quit
↑/k ↓/j        Move selection
PageUp/Down    Jump list
/              Search metadata and recent user messages
↑/↓            Move selection while searching
Ctrl-J/K       Move selection while typing search text
Ctrl-P         Preview current session while typing search text
Esc            Clear search
Enter          Resume selected session
e              Edit current session title
d              Delete current session (confirmation required)
p              Preview recent user messages
y              Copy session id
r              Reload sessions
?              Help
```

## Safety

The default mode opens the OpenCode database with SQLite `mode=rw` so title editing works directly. Use `--read-only` or `read_only = true` to prevent writes; read-write mode uses `mode=rw` and never creates a new database.

## Search Behavior

Search terms are split by spaces. All terms must match the combined session metadata and recent user messages, but they do not need to be adjacent or come from the same source.

```text
windows iso
```

This can match `windows` in the title and `iso` in one of your recent prompts. When user-message memory contributes to a match, a muted one-line excerpt appears below the session. Child-session messages are attributed to the root session that can be resumed from the list.

`search.recent_days` defaults to 7, accepts 0-365, and uses `0` to disable memory search. Memory is loaded only into the current process, is never written to a sidecar index, and is capped at 5,000 messages with 4,000 runes per message.

While typing in the search box, plain `j`, `k`, and `p` are treated as text input. Use `↑/↓` or `Ctrl-J/Ctrl-K` to move the selection without leaving search input. Use `Ctrl-P` to preview the selected session without leaving search input.
