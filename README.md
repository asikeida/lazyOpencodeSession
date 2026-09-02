# lazyOpencodeSession

`lazyOpencodeSession` is a fast terminal UI for browsing, searching, and resuming OpenCode sessions.

The command name is `lazyocs`.

## MVP Features

- Read OpenCode SQLite database in read-only mode.
- List root OpenCode sessions by recent update time.
- Search session metadata with `/`.
- Multi-term search with AND semantics, for example `windows iso` matches sessions containing both words.
- Show session details without scanning large message payloads.
- Show whether the session directory still exists.
- Lazy-load recent user message preview with `p`.
- Resume selected session with `Enter`.
- Copy selected session id with `y`.
- Built-in `default` and `dark` themes.
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

Use the dark theme:

```bash
./lazyocs --theme dark
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
# Path to OpenCode SQLite database. Leave empty to use ~/.local/share/opencode/opencode.db.
db = ""
language = "auto"
theme = "default"
limit = 500
opencode = "opencode"

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

Use `true` to show a field and `false` to hide it. `1` and `0` are also accepted.

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
p              Preview recent user messages
y              Copy session id
r              Reload sessions
?              Help
```

## Safety

The MVP opens the OpenCode database with `mode=ro`. It does not modify session data.

## Search Behavior

Search terms are split by spaces. All terms must match the session metadata, but they do not need to be adjacent.

```text
windows iso
```

This matches a session when both `windows` and `iso` appear somewhere in its title, directory, id, project, model, or agent.

While typing in the search box, plain `j`, `k`, and `p` are treated as text input. Use `↑/↓` or `Ctrl-J/Ctrl-K` to move the selection without leaving search input. Use `Ctrl-P` to preview the selected session without leaving search input.
