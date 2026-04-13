# ccs — Claude Code Session Manager

A fast CLI tool for managing [Claude Code](https://claude.ai/claude-code) sessions. List, search, tag, group by project, and resume sessions — all from your terminal.

## Features

- **List sessions** across all projects or filter by project/tag
- **Smart search** across session names, messages, working directories, and git branches
- **Tagging** — add custom labels to sessions for easy organization
- **Per-project grouping** — see all sessions for a project at a glance
- **Interactive TUI** — browse, filter, tag, and open sessions visually
- **Quick open** — resume any session in two steps: cd into its project dir + `claude --resume`
- **Incremental indexing** — blazing fast, only re-parses changed files

## Install

```bash
# Clone and build
git clone <repo-url> && cd cc-session
make build      # Output: ./bin/ccs
make install    # Copies to ~/go/bin/ccs
```

Requires Go 1.22+.

## Usage

### List sessions

```bash
ccs list              # Current project's sessions (most recent first)
ccs list --all        # All sessions across all projects
ccs list -t bugfix    # Filter by tag
ccs list -p nexus     # Filter by project (partial match)
ccs list -n 50        # Show more results (default: 20)
```

### Search

```bash
ccs search "auth middleware"   # Searches name, messages, cwd, branch
ccs search "NEX-73" --all     # Search across all projects
```

### Show details

```bash
ccs show <name-or-id>         # Full metadata, tags, message preview
ccs show distributed-painting-dragonfly
```

### Open / resume a session

```bash
ccs open <name-or-id>               # cd + claude --resume in current terminal
ccs open <name-or-id> --new-terminal # Open in a new shell process
```

Session identifiers can be: full UUID, UUID prefix (4+ chars), session name (exact or fuzzy).

### Tag management

```bash
ccs tag <session> bugfix wip         # Add tags
ccs untag <session> wip              # Remove a tag
ccs tags                             # List all tags with counts
```

### Projects

```bash
ccs projects                         # List all projects with session counts
```

### Interactive TUI

```bash
ccs ui
```

Key bindings:
- `↑/↓` or `j/k` — navigate
- `Enter` — open session
- `/` — search (real-time filter)
- `t` — add/remove tags
- `d` — delete session
- `Tab` — cycle project filter
- `?` — help
- `q` — quit

### Other

```bash
ccs reindex                          # Force full re-index
ccs completion bash > /etc/bash_completion.d/ccs  # Shell completions
```

### Global flags

```
--db-path <path>      Path to SQLite database (default: ~/.config/ccs/ccs.db)
--claude-dir <path>   Path to Claude data dir (default: ~/.claude)
--json                Machine-readable JSON output (on list/search/show)
```

## How it works

Claude Code stores session data as JSONL files in `~/.claude/projects/`. `ccs` scans these files, extracts metadata (name, messages, timestamps, branch), and indexes everything into a local SQLite database. The index is updated incrementally — only changed files are re-parsed.

Tags and labels are stored in the SQLite database only; `ccs` never modifies Claude Code's files.

## License

MIT
