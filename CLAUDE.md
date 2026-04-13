# CCS — Claude Code Session Manager

## What This Is

A Go CLI tool (`ccs`) for managing Claude Code sessions. Provides listing, smart search, tagging, per-project grouping, and session resumption via both CLI subcommands and an interactive TUI.

## Stack

- **Language:** Go (1.22+)
- **CLI framework:** `github.com/spf13/cobra`
- **TUI:** `github.com/charmbracelet/bubbletea` + `lipgloss` + `bubbles`
- **Database:** SQLite via `modernc.org/sqlite` (pure Go, no CGO)
- **Build:** Makefile + standard `go build`

## Project Structure

```
cmd/ccs/main.go          — Entry point, cobra root command setup
internal/cli/             — Cobra subcommand implementations (list, search, open, tag, etc.)
internal/db/              — SQLite connection, schema migrations, query functions
internal/indexer/         — JSONL file scanner and parser, incremental re-indexing
internal/model/           — Data structs (Session, Tag)
internal/opener/          — Session open/resume logic (same terminal + new terminal)
internal/tui/             — Bubbletea TUI (app model, views, components, styles)
```

## Build & Run

```bash
# Build
make build                # Output: ./bin/ccs

# Install to GOPATH
make install              # Copies to ~/go/bin/ccs

# Run directly during development
go run ./cmd/ccs/         # Run without building

# Build for specific platform
GOOS=linux GOARCH=amd64 go build -o bin/ccs ./cmd/ccs/
```

## Test

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests for a specific package
go test -v ./internal/indexer/

# Run a single test
go test -v -run TestParseSession ./internal/indexer/

# Run tests with race detector
go test -race ./...

# Test coverage
go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out
```

## Lint & Format

```bash
# Format all Go files
gofmt -w .

# Vet (built-in static analysis)
go vet ./...

# If golangci-lint is installed
golangci-lint run
```

## Dependencies

```bash
# Download dependencies
go mod download

# Tidy (remove unused, add missing)
go mod tidy

# Add a new dependency
go get <package>@latest
```

## How It Works

### Data Flow

1. Claude Code stores sessions as JSONL files at `~/.claude/projects/<sanitized-path>/<uuid>.jsonl`
2. `ccs` scans these directories and indexes session metadata into a local SQLite database at `~/.config/ccs/ccs.db`
3. Indexing is incremental — only re-parses files whose size or mod time changed since last index
4. All queries (list, search, filter) run against SQLite for speed
5. Tags and custom labels are stored in SQLite only (not in CC's JSONL files)

### JSONL Format

Each session file contains one JSON object per line with these key fields:
- `sessionId` — UUID identifying the session
- `cwd` — working directory where the session was started
- `timestamp` — ISO 8601 timestamp
- `gitBranch` — git branch at time of message
- `slug` — session name (auto-generated or set via `claude --name`)
- `type` — entry type: `user`, `assistant`, `file-history-snapshot`, `permission-mode`
- `message.role` — `user` or `assistant`
- `message.content` — message text (string or array of content blocks)

### Session Opening

When opening a session, `ccs`:
1. Looks up `cwd` and `sessionId` from SQLite
2. Spawns `bash -c "cd <cwd> && claude --resume <sessionId>"`
3. In `--new-terminal` mode, detaches the process with `Setsid`

### ID Resolution

Session identifiers are resolved in order: exact UUID > UUID prefix (4+ chars) > exact name > fuzzy name match. Ambiguous matches prompt the user to pick.

## Verification (for CI / automated agents)

```bash
go vet ./...            # Static analysis (replaces tsc --noEmit)
go build ./...          # Confirm compilation
go test ./...           # Run all tests
```

Do NOT use `npm`, `tsc`, `npx`, or any Node.js tooling — this is a pure Go project.

## Key Design Decisions

- **Pure Go SQLite** (`modernc.org/sqlite`) — no CGO dependency, cross-compilation works out of the box
- **Incremental indexing** — compare file size + mod time, only re-parse changed files
- **No daemon required** — indexing happens lazily on each command invocation; fast enough for hundreds of sessions
- **Tags in SQLite only** — we don't modify CC's session files; our metadata is separate

## Spec

Full design spec: `docs/superpowers/specs/2026-04-13-ccs-design.md`
