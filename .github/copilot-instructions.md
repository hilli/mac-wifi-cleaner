# Copilot Instructions for mac-wifi-cleaner

## Build & Test

```sh
go build -o mac-wifi-cleaner .
go test ./...
go test -run TestParseSSIDs ./...   # single test
go vet ./...                        # lint
```

## Architecture

Single-binary CLI with manual arg parsing (no framework). Two source files:

- **wifi.go** — macOS `networksetup` interaction. Each function follows a wrapper/parser split: thin `exec.Command` wrappers (`detectWiFiInterface`, `listSSIDs`, `removeSSID`) call pure parsing functions (`parseWiFiInterface`, `parseSSIDs`) that are independently testable.
- **main.go** — CLI entry point, subcommands (`list`, `delete`, `keep`, `auto`), file I/O helpers, and arg parsing utilities.

Tests cover the pure functions only; anything calling `os/exec` or reading `os.Args` is untested by design.

## Conventions

- No external dependencies — stdlib only.
- Arg parsing uses hand-rolled `flagValue(args, "-f")` and `hasFlag(args, "--dry-run")` helpers instead of `flag` package. Follow this pattern when adding new flags.
- New `networksetup` parsing should follow the wrapper/parser split: exec wrapper calls a `parse*` pure function, tests go against the parser.
- SSID files are plain text, one per line. Lines starting with `#` are comments, blank lines are ignored.
- `$EDITOR` is split with `strings.Fields()` to support editors with flags (e.g. `zed --wait`).

## Release

Releases are via GoReleaser on tag push. macOS only (amd64 + arm64). The workflow uses `PACKAGES_TOKEN` for the release and the default `GITHUB_TOKEN` for issue creation on failure.
