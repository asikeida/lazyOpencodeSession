# lazyOpencodeSession

`lazyOpencodeSession` is a fast terminal UI for browsing, searching, and resuming OpenCode sessions.

The command name is `lazyocs`.

## MVP Features

- Read OpenCode SQLite database in read-only mode.
- List root OpenCode sessions by recent update time.
- Search session metadata with `/`.
- Show session details without scanning large message payloads.
- Lazy-load recent user message preview with `p`.
- Resume selected session with `Enter`.
- Copy selected session id with `y`.
- Built-in `default` and `dark` themes.
- Chinese UI with `--language auto|en|zh-CN`.

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

## Key Bindings

```text
q / Ctrl+C     Quit
↑/k ↓/j        Move selection
PageUp/Down    Jump list
/              Search metadata
↑/↓            Move selection while searching
Esc            Clear search
Enter          Resume selected session
p              Preview recent user messages
y              Copy session id
r              Reload sessions
?              Help
```

## Safety

The MVP opens the OpenCode database with `mode=ro`. It does not modify session data.
