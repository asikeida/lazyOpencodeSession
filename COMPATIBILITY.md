# Compatibility

`lazyocs` reads OpenCode's internal SQLite database. That schema is not a stable public API, so compatibility is determined from database capabilities at startup rather than only from an OpenCode version number.

## Supported Platforms

Release archives are built for:

- Linux amd64
- Linux arm64
- macOS amd64
- macOS arm64

Windows is not currently released or manually validated. Its clipboard integration is also unsupported.

## Database Capabilities

Browsing requires the `session` columns used by the list and details views, including identifiers, parent relationship, title, directory, timestamps, model, agent, cost, and token counters. If these fields are missing, startup fails with a compatibility error.

Capabilities are checked independently:

| Capability | Requirement | Failure behavior |
| --- | --- | --- |
| Browse | Required `session` columns | Refuse startup |
| Statistics | `message` and `part` session/data columns | Disable the operation with an error |
| Preview and memory search | Message/part identifiers, relationships, data, and timestamps | Disable the operation with an error |
| Rename | Session ID and title columns | Block title updates |
| Delete | Session parent relationship plus message-to-session and part-to-message `ON DELETE CASCADE` foreign keys | Block impact analysis and deletion |

Use `lazyocs --check` to inspect the current database:

```text
schema: browse=true stats=true preview=true rename=true delete=true
```

Unknown or changed schemas are never assumed safe for writes. Use `--read-only` when testing a new OpenCode release, and report the `--check` capability line without sharing session titles or message content.

## Clipboard

- Linux: tries `wl-copy`, `xclip`, then `xsel`.
- macOS: uses `pbcopy`.
- Windows: unsupported; the status line exposes the session ID for manual copying.
