# TUI Viewer MVP

## Summary
Added interactive TUI viewer as `cmd/tui` subcommand to go-arch-guard.
Uses tview (rivo/tview) with 2-panel layout: package tree + dependency detail.

## Files Changed
- `tui/tree.go` — Package tree builder (analyzer → TreeNode hierarchy, layer color-coding)
- `tui/detail.go` — Detail panel (imports + imported-by for selected package)
- `tui/app.go` — tview Application setup, 2-panel Flex layout, keyboard handling
- `tui/tree_test.go` — Unit tests for tree building and imported-by map
- `cmd/tui/main.go` — CLI entry point (accepts project directory argument)
- `go.mod`, `go.sum` — Added tview, tcell dependencies
- `README.md`, `README.ko.md` — Added TUI Viewer section

## Verification
- `make lint` — 0 issues
- `go test ./...` — all pass
- Manual test pending: `go run ./cmd/tui ./testdata/valid`
