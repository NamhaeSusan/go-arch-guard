# 2026-03-24 Architecture Model Consolidation

## Task
Fix structure rule false positives on empty directories and consolidate duplicated architecture constants into a single source of truth.

## Changes

### Fix: Structure false positives on empty directories
- **File**: `rules/structure.go`
- **Problem**: `checkPackageNames` and `checkMiddlewarePlacement` flagged directories by name alone, even when they contained zero Go files. An empty `internal/domain/order/shared/` triggered `structure.banned-package`.
- **Fix**: Added `hasNonTestGoFiles(path)` guard to both walkers. Directories without non-test Go files are silently skipped.
- **Tests**: 2 new tests in `rules/structure_test.go` confirming empty banned-name and middleware directories are ignored.

### Refactor: Consolidate architecture constants into `rules/arch.go`
- **File**: `rules/arch.go` (new), `rules/arch_test.go` (new)
- **Problem**: Sublayer lists, banned names, and layer-dir names were duplicated across `layer.go`, `naming.go`, and `structure.go`. The `policy` entry existed in naming's `layerDirs` but was absent from layer's `knownSublayers` — producing inconsistent diagnostics.
- **Fix**: Extracted all architecture model constants into `rules/arch.go` as the single source of truth. Removed all local duplicates from `layer.go`, `naming.go`, `structure.go`.
- **Removed**: `policy` from `layerDirNames` (was never a known sublayer).
- **Added**: `isKnownSublayer()` helper, `TestArchModelConsistency` preventing future drift (forward + reverse checks).

### Files changed
- `rules/arch.go` — new: consolidated constants
- `rules/arch_test.go` — new: consistency test
- `rules/layer.go` — removed local maps, removed `knownSublayerList()`
- `rules/naming.go` — removed `layerDirs` (with stale `policy`)
- `rules/structure.go` — removed local vars, added empty-dir guards

## Verification
- `go test ./...` — all pass
- `make lint` — 0 issues
