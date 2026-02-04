# Burrow — Developer-First macOS Cleanup

<p align="center">
  <img src="assets/logo.svg" width="200" alt="Burrow Logo">
</p>

Burrow is a macOS-first command-line tool designed for developers to safely identify and remove unnecessary files left behind by development tools, SDKs, and package managers.

## Design Philosophy

- **Safety > Aggressiveness**: Never delete without explanation.
- **Explainable**: Tell the user *why* a file can be safely removed.
- **Reversible**: Move to trash instead of permanent deletion.
- **Zero Telemetry**: All operations are local and private.

## Commands

```bash
burrow scan      # Identify cleanup candidates
burrow clean     # Execute cleanup (dry-run by default)
burrow undo      # Restore last cleanup session
burrow list      # Detailed list of found files
burrow rules     # List all available cleanup rules
burrow stats     # Show disk reclaimable statistics
burrow doctor    # Check system health and permissions
burrow version   # Show version information
```

## Advanced Usage

Filter scans by category:

```bash
burrow scan --category "Developer Tools"
```

Explain why files are being flagged:

```bash
burrow scan --explain
```

Output results as JSON for automation:

```bash
burrow stats --json | jq .
```

Clean without confirmation (CI/CD mode):

```bash
burrow clean --yes
```

## Categories Covered

- **Package Managers**: Homebrew, npm, pip, Cargo registry index/cache.
- **Developer Tools**: Xcode DerivedData, Android Build Cache, Gradle Cache.
- **System**: General user caches and temporary files.

## Safety Guardrails

Burrow includes hard-coded safety checks to prevent deletion of:

- Paths containing Git metadata (`.git`).
- System-protected paths (SIP).
- User documents, desktop, and downloads.
- Home directory and root.

## Recovery

All deleted files are moved to `~/.burrow/trash/<timestamp>/`. You can restore the most recent session using:

```bash
burrow undo
```

## Installation

Install directly using Go:

```bash
go install github.com/ismailtsdln/burrow/cmd/burrow@latest
```

## Configuration

Customize Burrow by creating `~/.config/burrow/config.json`:

```json
{
  "disabled_categories": ["Developer Tools"],
  "excluded_paths": ["/Users/me/important_cache"],
  "size_threshold_mb": 100
}
```

## Project Structure

- `cmd/burrow/`: Entry point.
- `internal/scanner/`: Filesystem traversal logic.
- `internal/cleaner/`: Trash management and deletion logic.
- `internal/rules/`: Rules engine and definitions.
- `internal/safety/`: Hand-blocked safety guardrails.
- `internal/ui/`: CLI interface and output formatting.
