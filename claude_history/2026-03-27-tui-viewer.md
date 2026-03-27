# TUI Viewer

## Summary
Added interactive TUI viewer as `cmd/tui` subcommand to go-arch-guard.
Uses tview (rivo/tview) with 2-panel layout: package tree + dependency detail.

### MVP (commit 1)
- Package tree with layer color-coding
- Imports / imported-by detail panel

### v2 Upgrade (commit 2)
- Violation highlighting (red ✗ on violated packages)
- Violation details in detail panel (severity, rule, message, fix)
- Blast radius metrics (Ca, Ce, Instability, Transitive Dependents)
- Search/filter with `/` key
- Status bar with package/violation counts

## Files Changed
- `tui/tree.go` — Package tree builder with violation marking
- `tui/detail.go` — Detail panel with violations + metrics sections
- `tui/app.go` — Application setup, search input, status bar
- `tui/violations.go` — Run all rules and index by package path
- `tui/metrics.go` — Coupling metrics computation (Ca, Ce, Instability, BFS)
- `tui/tree_test.go` — Tests for tree, violations, metrics
- `cmd/tui/main.go` — CLI entry point
- `README.md`, `README.ko.md` — Updated TUI Viewer section

## Verification
- `make lint` — 0 issues
- `go test ./...` — all pass
- Manual test: `go run ./cmd/tui ./testdata/valid`
