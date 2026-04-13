# ccs — Claude Code Session Manager

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/mihaiandrei12234/ccs)](https://goreportcard.com/report/github.com/mihaiandrei12234/ccs)

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

### From source

Requires [Go 1.22+](https://go.dev/dl/).

```bash
# Clone the repository
git clone https://github.com/mihaiandrei12234/ccs.git
cd ccs

# Build the binary
make build        # Output: ./bin/ccs

# Or install directly to your GOPATH
make install      # Installs to ~/go/bin/ccs
```

### Manual build (without Make)

```bash
go build -o bin/ccs ./cmd/ccs/
```

Make sure `~/go/bin` (or your `GOPATH/bin`) is in your `PATH`.

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

## Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details on:

- Setting up the development environment
- Running tests
- Code style guidelines
- Branch naming conventions
- Submitting pull requests

## License

This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.
