# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2026-09-05

### Added

- Terminal UI for browsing, searching, previewing, and resuming root OpenCode sessions.
- Metadata and recent-user-message search with debounce and multi-term AND semantics.
- Session title editing and recursive deletion with impact analysis and transactional scope revalidation.
- Read-only mode, schema capability checks, and actionable database lock errors.
- Configurable details, preview depth, panel layout, language, themes, and external theme files.
- Linux and macOS clipboard integration with asynchronous timeout handling.
- Generated-data performance benchmarks and GitHub Actions continuous integration.

### Security

- Write operations fail closed when required schema capabilities or delete cascades are absent.
- Title input strips control characters and deletion requires explicit confirmation.

[Unreleased]: https://github.com/asikeida/lazyOpencodeSession/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/asikeida/lazyOpencodeSession/releases/tag/v0.1.0
