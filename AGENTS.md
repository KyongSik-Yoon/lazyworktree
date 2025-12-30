## Project

Textual-based TUI for managing Git worktrees. Python 3.12+.

## Building

- use go build -o /dev/null ./cmd/lazyworktree/... for testing build errors.

## Before Finishing

Always Run:

- `golangci-lint run --fix ./cmd/... ./internal/...`
- `gofumpt -w ./cmd/... ./internal/...`
