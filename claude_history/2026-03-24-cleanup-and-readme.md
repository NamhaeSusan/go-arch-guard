# 2026-03-24 Cleanup and README update

## Summary

Minor cleanup: stale comments, code simplification, and README clarification.

## Changes

### README.md
- Rewrote "Future Considerations > External Import Restrictions" to "External Import Hygiene — NOT PLANNED"
- Stronger language: will not be added, copy constraints into AI tool instructions instead

### rules/helpers.go
- Removed stale `//nolint:unused` comments on `findImportFile` and `findImportLine`
- These functions are actively used by isolation.go and layer.go

### rules/isolation.go
- Simplified orchestration branch (lines 72-107 → 72-93)
- Collapsed 3 overlapping `isSrcHandler` conditions into unified logic
- Same behavior: domain alias allowed, sub-package denied for all orchestration

## Verification

- `go test ./...` — all pass
- `make lint` — 0 issues
