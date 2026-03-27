# TUI severity-based color scheme

## Summary
Changed TUI tree colors from layer-based (cmd=blue, domain=green, etc.)
to health-status-based: green (clean), yellow ⚠ (warnings only),
red ✗ (errors). blast-radius.high-coupling now shows as yellow warning
instead of red error.

## Files Changed
- `tui/violations.go` — Added Severity() method with walkPath, replaced HasViolations
- `tui/tree.go` — Replaced layerColor() with severityStyle(), green/yellow/red scheme

## Verification
- `make lint` — 0 issues
- `go test ./tui/...` — all pass
