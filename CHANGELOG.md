# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.2] - 2026-09-06

### Added

- `[resume.env]` support for passing proxy and other environment variables to resumed OpenCode sessions.

### Changed

- Document proxy configuration as environment variables instead of the unsupported `opencode --proxy` CLI flag.

## [0.1.1] - 2026-09-05

### Added

- Configurable `[resume]` command arguments for cases such as `opencode --proxy`.

## [0.1.0] - 2026-09-05

### Added

- Terminal UI for browsing, searching, previewing, and resuming root OpenCode sessions.
- Metadata and recent-user-message search with debounce and multi-term AND semantics.
- Session title editing and recursive deletion with impact analysis and transactional scope revalidation.
- Read-only mode, schema capability checks, and actionable database lock errors.
- Configurable details, preview depth, panel layout, language, themes, and external theme files.
- Linux and macOS clipboard integration with asynchronous timeout handling.
- Windows amd64/arm64 zip builds and `clip.exe` clipboard integration.
- Linux deb, rpm, and Arch Linux package artifacts.
- Generated `lazyocs-bin` AUR package metadata and bilingual release documentation.
- Generated-data performance benchmarks and GitHub Actions continuous integration.

### Security

- Write operations fail closed when required schema capabilities or delete cascades are absent.
- Title input strips control characters and deletion requires explicit confirmation.

[Unreleased]: https://github.com/asikeida/lazyOpencodeSession/compare/v0.1.2...HEAD
[0.1.2]: https://github.com/asikeida/lazyOpencodeSession/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/asikeida/lazyOpencodeSession/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/asikeida/lazyOpencodeSession/releases/tag/v0.1.0
