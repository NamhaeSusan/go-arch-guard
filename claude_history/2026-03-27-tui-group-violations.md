# TUI group node violations & trailing slash fix

## Summary
- Fixed trailing slash mismatch: CheckStructure reports paths like
  "internal/domain/settings/" but tree keys have no trailing slash.
  Now ViolationIndex normalizes paths on insert.
- Group nodes (domain, internal, pkg) now show aggregated violation
  summary when selected: error/warning counts + violations grouped
  by file path.

## Files Changed
- `tui/violations.go` — TrimRight trailing slash on index insert
- `tui/detail.go` — Added renderGroup() for non-leaf nodes with
  violation aggregation by file path

## Verification
- `make lint` — 0 issues
- `go test ./tui/...` — all pass
