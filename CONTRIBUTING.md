# Contributing to ccs

Thanks for your interest in contributing to ccs! This guide covers everything you need to get started.

## Development Environment Setup

### Prerequisites

- Go 1.22 or later
- Git
- Make (optional, for convenience targets)
- A working [Claude Code](https://claude.ai/claude-code) installation (for testing with real session data)

### Getting Started

```bash
# Fork and clone the repository
git clone https://github.com/<your-username>/ccs.git
cd ccs

# Download dependencies
go mod download

# Build
make build        # Output: ./bin/ccs

# Verify everything works
go vet ./...
go test ./...
```

## Running Tests

```bash
# Run all tests
go test ./...

# Verbose output
go test -v ./...

# Run tests for a specific package
go test -v ./internal/indexer/

# Run a single test
go test -v -run TestParseSession ./internal/indexer/

# Run tests with race detector
go test -race ./...

# Check test coverage
go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out
```

All new code should include tests. Tests use real SQLite (via `modernc.org/sqlite`) — do not mock the database layer.

## Code Style

- Follow standard Go conventions (`gofmt`, `go vet`)
- Format all code with `gofmt -w .` before committing
- Use `golangci-lint run` if available for additional checks
- Keep functions focused and small
- Use meaningful variable and function names
- Add comments only where the logic is non-obvious
- No CGO dependencies — we use pure Go SQLite (`modernc.org/sqlite`)

## Branch Naming

Use the following branch naming convention:

```
feat/<short-description>     # New features
fix/<short-description>      # Bug fixes
docs/<short-description>     # Documentation changes
refactor/<short-description> # Code refactoring
chore/<short-description>    # Maintenance tasks
```

Examples: `feat/session-export`, `fix/tui-crash-on-empty-list`, `docs/install-guide`

## Submitting a Pull Request

1. **Fork** the repository and create your branch from `main`
2. **Make your changes** — keep commits focused and atomic
3. **Write/update tests** for any changed functionality
4. **Run verification** before pushing:
   ```bash
   gofmt -w .
   go vet ./...
   go build ./...
   go test ./...
   ```
5. **Push** your branch and open a pull request
6. **Fill out the PR template** — describe what changed and why
7. **Link related issues** — use "Closes #123" in the PR description

### PR Guidelines

- Keep PRs focused on a single change
- Update documentation if your change affects user-facing behavior
- Ensure all tests pass and no new `go vet` warnings are introduced
- Add a clear description of what the PR does and why
- If the PR is a work in progress, mark it as a draft

## Commit Messages

Use conventional commit style:

```
type: short description

Optional longer description explaining the motivation
for the change.
```

Types: `feat`, `fix`, `docs`, `refactor`, `test`, `chore`

## Reporting Issues

- Use the [bug report template](.github/ISSUE_TEMPLATE/bug_report.md) for bugs
- Use the [feature request template](.github/ISSUE_TEMPLATE/feature_request.md) for new ideas
- Search existing issues before creating a new one

## AI Agents Welcome

This project was built entirely by AI agents and we welcome AI-assisted contributions. If you're using Claude Code, Copilot, Cursor, or any other AI coding tool to prepare your PR — great, no need to hide it. Just make sure the code compiles, tests pass, and the PR description accurately describes the changes.

## Questions?

Open a discussion or issue if something is unclear. We're happy to help!
