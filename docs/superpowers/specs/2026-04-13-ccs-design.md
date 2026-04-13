# CCS — Claude Code Session Manager

## Overview

`ccs` is a Go CLI tool for managing Claude Code sessions. It provides listing, smart search, tagging, per-project grouping, and session resumption. It includes both a CLI interface and an interactive TUI.

## Problem

Claude Code's built-in session management (`--resume`, `--continue`) is limited:
- No tagging or labeling beyond `--name`
- No search across session content
- No per-project grouping or overview
- No interactive browser for picking sessions
- Opening a session requires knowing the UUID or being in the right directory

## Data Source

CC stores sessions as JSONL files at `~/.claude/projects/<sanitized-cwd-path>/<uuid>.jsonl`.

Each JSONL line is a JSON object with fields:
- `sessionId` — UUID
- `cwd` — original working directory
- `timestamp` — ISO 8601
- `gitBranch` — branch at time of session
- `type` — `user` or `assistant`
- `message.role` — `user` or `assistant`
- `message.content` — message text
- `version` — CC version
- Subagent sessions live in `<uuid>/subagents/` subdirectories (excluded from indexing)

Session names are stored as the `slug` field in JSONL entries (e.g. `"slug":"distributed-painting-dragonfly"`). This is either auto-generated or set via CC's `--name` flag.

## Data Model & Storage

### SQLite Database

Location: `~/.config/ccs/ccs.db`

```sql
CREATE TABLE sessions (
  id TEXT PRIMARY KEY,              -- UUID from filename
  project_dir TEXT NOT NULL,        -- sanitized project path key (directory name)
  cwd TEXT NOT NULL,                -- original working directory
  git_branch TEXT,
  name TEXT,                        -- from CC's --name flag if set
  first_message TEXT,               -- first user message (truncated to 200 chars)
  last_message TEXT,                -- last user/assistant message (truncated to 200 chars)
  message_count INTEGER DEFAULT 0,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,    -- last message timestamp
  file_size INTEGER NOT NULL,       -- for change detection
  file_mod_time TIMESTAMP NOT NULL  -- for change detection
);

CREATE TABLE tags (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT UNIQUE NOT NULL
);

CREATE TABLE session_tags (
  session_id TEXT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
  tag_id INTEGER NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
  PRIMARY KEY (session_id, tag_id)
);

CREATE INDEX idx_sessions_project ON sessions(project_dir);
CREATE INDEX idx_sessions_updated ON sessions(updated_at);
CREATE INDEX idx_sessions_name ON sessions(name);
```

### Indexing Strategy

On any list/search command:
1. Scan `~/.claude/projects/` for `*.jsonl` files (skip `subagents/` directories)
2. Compare each file's `size` + `mod_time` against SQLite records
3. Only re-parse changed or new files
4. Remove index entries for deleted files
5. `ccs reindex` forces a full re-parse of everything

When parsing a JSONL file, extract:
- `sessionId` from first entry
- `cwd`, `gitBranch`, `version` from first entry
- Session `name` from the `slug` field in any entry
- First user message content (truncated to 200 chars)
- Last message content (truncated to 200 chars)
- Total message count (user + assistant entries only, skip meta)
- First and last timestamps

## CLI Commands

```
ccs list [flags]                      # List sessions
  --all, -a                           # All projects (default: current project)
  --project, -p <path>                # Filter by project directory
  --tag, -t <tag>                     # Filter by tag (repeatable)
  --limit, -n <num>                   # Max results (default: 20)
  --sort <field>                      # Sort by: updated, created, name (default: updated)
  --json                              # JSON output

ccs search <query> [flags]            # Smart search across name, messages, cwd, branch
  --all, -a                           # Search all projects
  --tag, -t <tag>                     # Also filter by tag
  --json                              # JSON output

ccs show <id|name>                    # Show session details: metadata, tags, message preview
  --json                              # JSON output

ccs open <id|name> [flags]            # cd into project dir + resume CC session
  --new-terminal                      # Spawn in a new shell process

ccs tag <id|name> <tag> [tags...]     # Add tags to a session
ccs untag <id|name> <tag> [tags...]   # Remove tags from a session

ccs tags                              # List all tags with session counts

ccs projects                          # List all known projects with session counts

ccs delete <id|name> [flags]          # Delete session JSONL file + remove from index
  --force, -f                         # Skip confirmation prompt

ccs reindex                           # Force full re-index

ccs ui                                # Launch interactive TUI

ccs completion <bash|zsh|fish>        # Generate shell completions
```

### ID Resolution

Anywhere `<id|name>` is accepted, matching order:
1. Exact UUID match
2. UUID prefix match (minimum 4 chars)
3. Exact session name match
4. Fuzzy session name match

If multiple matches, prompt the user to pick (CLI) or show filtered list (TUI).

### Output

Default: human-readable table with columns adapted to terminal width.
`--json`: machine-readable JSON array.

## TUI Design

Built with `bubbletea` + `lipgloss` + `bubbles`.

### Layout

```
+-- ccs -------------------------------------------------+
| [Search: _______________] [Filter: tag:bugfix] [All v] |
+---------+---------------+-----------+--------+---------+
|  #      | Name/ID       | Project   | Branch | Updated |
|  ----   | ------------- | --------- | ------ | ------- |
|  > 1    | auth-refactor | nexus     | NEX-73 | 2h ago  |
|    2    | fix-login-bug | nexus     | main   | 1d ago  |
|    3    | 490ed12d      | cc-session| -      | 3d ago  |
|    4    | dashboard-v2  | dashboard | feat/  | 5d ago  |
+---------+---------------+-----------+--------+---------+
| Tags: [bugfix] [wip]                                   |
| CWD:  /mnt/c/Users/mihai/JsProjects/nexus              |
| First: "refactor the auth middleware to use..."         |
| Last:  "Done. All tests passing."                       |
+---------+---------------+-----------+--------+---------+
| [Enter] Open [t] Tag [d] Delete [/] Search [q] Quit    |
+---------------------------------------------------------+
```

### Key Bindings

- `up/down` or `j/k` — navigate session list
- `Enter` — open selected session
- `/` — focus search input (real-time filtering)
- `t` — add/remove tags on selected session
- `d` — delete with confirmation
- `Tab` — cycle project filter
- `p` — project grouping view
- `?` — help overlay
- `q` / `Esc` — quit

### Detail Pane

Bottom pane shows metadata for the highlighted session, updating on navigation.

## Session Opening

### Same Terminal (default)

```go
cmd := exec.Command("bash", "-c",
    fmt.Sprintf("cd %q && claude --resume %q", cwd, sessionId))
cmd.Stdin = os.Stdin
cmd.Stdout = os.Stdout
cmd.Stderr = os.Stderr
cmd.Run()
```

### New Terminal (`--new-terminal`)

```go
cmd := exec.Command("bash", "-c",
    fmt.Sprintf("cd %q && claude --resume %q", cwd, sessionId))
cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
cmd.Start()
```

### Edge Cases

- If `cwd` no longer exists: warn and offer to open in current directory
- If JSONL file deleted externally: remove from index, report error
- TUI exits cleanly before shell handoff

## Project Structure

```
ccs/
├── cmd/
│   └── ccs/
│       └── main.go                # Entry point, cobra root command
├── internal/
│   ├── cli/                       # Cobra subcommands
│   │   ├── list.go
│   │   ├── search.go
│   │   ├── show.go
│   │   ├── open.go
│   │   ├── tag.go
│   │   ├── untag.go
│   │   ├── tags.go
│   │   ├── projects.go
│   │   ├── delete.go
│   │   ├── reindex.go
│   │   ├── ui.go
│   │   └── completion.go
│   ├── db/                        # SQLite schema, migrations, queries
│   │   ├── db.go
│   │   ├── schema.go
│   │   └── queries.go
│   ├── indexer/                   # JSONL scanning, parsing, incremental indexing
│   │   ├── scanner.go
│   │   └── parser.go
│   ├── model/                     # Session, Tag structs
│   │   └── model.go
│   ├── opener/                    # Session open/resume logic
│   │   └── opener.go
│   └── tui/                       # Bubbletea TUI
│       ├── app.go                 # Main model, Update, View
│       ├── list.go                # Session list component
│       ├── detail.go              # Detail pane component
│       ├── search.go              # Search input component
│       ├── tag.go                 # Tag dialog component
│       ├── help.go                # Help overlay
│       └── styles.go              # Lipgloss styles
├── CLAUDE.md                      # Project knowledge for CC sessions
├── go.mod
├── go.sum
└── Makefile
```

## Dependencies

- `github.com/spf13/cobra` — CLI framework
- `github.com/charmbracelet/bubbletea` — TUI framework
- `github.com/charmbracelet/lipgloss` — TUI styling
- `github.com/charmbracelet/bubbles` — TUI components (table, textinput, viewport)
- `modernc.org/sqlite` — pure Go SQLite driver (no CGO)

## Enrichments

### Included in initial build:
- **Relative timestamps** — "2h ago", "3d ago" in list/TUI output
- **Staleness indicators** — visual highlight for sessions older than 30 days
- **Shell completions** — bash/zsh/fish via cobra

### Stretch goals (post-initial):
- **Bulk operations** — `ccs delete --tag <tag>`, `ccs tag --all-in-project <path> <tag>`
- **Export** — `ccs show <id> --full` dumps conversation as readable markdown
- **`ccs stats`** — total sessions, per-project counts, tag usage, oldest/newest

## Build & Install

```bash
make build        # Builds to ./bin/ccs
make install      # Copies to ~/go/bin/ccs
```
