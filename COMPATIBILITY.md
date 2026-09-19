# Compatibility

`lazyocs` reads OpenCode's internal SQLite database. That schema is not a stable public API, so compatibility is determined from database capabilities at startup rather than only from an OpenCode version number.

## Supported Platforms

Release archives are built for:

- Linux amd64
- Linux arm64
- macOS amd64
- macOS arm64
- Windows amd64
- Windows arm64

Windows builds are experimental until they complete native terminal testing. Clipboard integration uses the system `clip.exe`, and copy failures expose the session ID for manual copying.

## Database Discovery

When no explicit database is configured, lazyocs runs `opencode debug paths db`. This supports OpenCode release channels and `OPENCODE_DB`. If discovery is unavailable, lazyocs falls back to `OPENCODE_DB` and then `~/.local/share/opencode/opencode.db`.

Explicit path precedence is `--db` or config `db`, then `LAZYOCS_DB`.

## Database Capabilities

lazyocs supports both storage generations:

- V1: `session`, `message`, and `part`.
- V2: `session_v2` and `session_message`.

If both generations exist, V2 is always selected. An incomplete V2 schema is never silently treated as V1.

Capabilities are checked independently:

| Capability | V1 requirement | V2 requirement | Failure behavior |
| --- | --- | --- | --- |
| Browse | Required `session` columns | Required `session_v2` columns | Refuse startup |
| Statistics | `message` and `part` session/data columns | `session_message` type/data columns | Disable the operation with an error |
| Preview and memory search | Message/part identifiers, relationships, data, and timestamps | `session_message` identifiers, type, data, and timestamps | Disable the operation with an error |
| Rename | Session ID and title columns | `session_v2` identity plus the OpenCode API | Block title updates |
| Delete | Session relationship and required `ON DELETE CASCADE` foreign keys | V2 relationship/message columns plus the OpenCode API | Block impact analysis and deletion |

Use `lazyocs --check` to inspect the current database:

```text
schema: version=v2 browse=true stats=true preview=true rename=true delete=true
```

V1 title updates and deletion use guarded SQLite transactions. V2 title updates and deletion use the official OpenCode API; title editing supports both the newer session PATCH endpoint and the V2.0.3 rename endpoint. lazyocs never writes V2 session tables directly. The configured OpenCode executable is used for these calls.

Unknown or changed schemas are never assumed safe for writes. Use `--read-only` when testing a new OpenCode release, and report the `--check` capability line without sharing session titles or message content.

## Clipboard

- Linux: tries `wl-copy`, `xclip`, then `xsel`.
- macOS: uses `pbcopy`.
- Windows: uses the system `clip.exe`.

## Package Status

- Linux tar.gz, deb, rpm, and Arch Linux packages are generated automatically.
- Windows binaries are distributed as zip archives containing `lazyocs.exe`.
- macOS binaries are distributed as unsigned tar.gz archives and are not notarized.
- AUR metadata is generated as `lazyocs-bin`, but publishing requires a separately registered AUR package repository and maintainer SSH key.
