# lazyOpencodeSession

`lazyOpencodeSession` is a fast terminal UI for browsing, searching, and resuming OpenCode sessions.

The command name is `lazyocs`.

## MVP Features

- Read OpenCode SQLite database with title editing enabled by default.
- List root OpenCode sessions by recent update time.
- Search session metadata with `/`.
- Multi-term search with AND semantics, for example `windows iso` matches sessions containing both words.
- Show session details without scanning large message payloads.
- Show whether the session directory still exists.
- Lazy-load recent user message preview with `p`.
- Resume selected session with `Enter`.
- Edit the selected session title with `e`.
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

The default color scheme follows lazygit's general terminal style: green accent/border, blue selected line, blue options text, and terminal-provided background. Colors are currently built in rather than user-configurable.

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

Statistics fields are loaded only for the selected session when one of them is enabled. `large_session` means the session message and part payloads are at least 10 MB.

## Key Bindings

```text
q / Ctrl+C     Quit
↑/k ↓/j        Move selection
PageUp/Down    Jump list
/              Search metadata
↑/↓            Move selection while searching
Ctrl-J/K       Move selection while typing search text
Ctrl-P         Preview current session while typing search text
Esc            Clear search
Enter          Resume selected session
e              Edit current session title
p              Preview recent user messages
y              Copy session id
r              Reload sessions
?              Help
```

## Safety

The default mode opens the OpenCode database with SQLite `mode=rw` so title editing works directly. Use `--read-only` or `read_only = true` to prevent writes; read-write mode uses `mode=rw` and never creates a new database.

## Search Behavior

Search terms are split by spaces. All terms must match the session metadata, but they do not need to be adjacent.

```text
windows iso
```

This matches a session when both `windows` and `iso` appear somewhere in its title, directory, id, project, model, or agent.

While typing in the search box, plain `j`, `k`, and `p` are treated as text input. Use `↑/↓` or `Ctrl-J/Ctrl-K` to move the selection without leaving search input. Use `Ctrl-P` to preview the selected session without leaving search input.
