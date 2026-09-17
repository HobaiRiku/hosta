# Changelog

All notable changes to Hosta will be documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and releases use semantic versioning.

## [Unreleased]

### Added

- Homebrew Tap publishing and an interactive shortcut to copy a resolved SSH command.
- `Ctrl+]` opens a Group-name filter in the launcher; it filters results while typing.
- The launcher displays a connection status and clears the terminal after OpenSSH establishes the SSH session.
- OpenSSH Config discovery with recursive Include graph and source diagnostics.
- Weighted Unicode Host search and interactive terminal launcher.
- `connect`, `list`, `show`, `config`, `doctor`, `completion`, and `version` commands.
- Linux, macOS, and Windows build and release configuration with injected version metadata.

### Changed

- Hosta metadata uses `# @hosta.*` comments instead of custom OpenSSH directives.
- Removed Tags from Hosta metadata, search, and CLI output. Existing bare `Tags` directives retain an OpenSSH compatibility guard but are no longer read.
