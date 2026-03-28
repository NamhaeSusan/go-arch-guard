# 2026-03-28 Code Review Fixes

## Task
Code review of the entire codebase and fix identified issues.

## Changes

### rules/naming.go
- **Fix `stutters()` UTF-8 safety**: Changed byte-indexing to rune-based indexing. Previous code used `typeName[:len(pkgName)]` which is unsafe for non-ASCII Go identifiers.
- **Fix `stutters()` suggested name bug**: Previous code lowercased the entire type name before trimming prefix, losing camelCase casing in the remainder (e.g., `UserOrderID` → `Orderid` instead of `OrderID`). Now slices by rune length to preserve original casing.
- **Fix `isDomainPackage` scope**: Changed from `/domain/` to `/internal/domain/` to avoid false positives on external dependencies whose import paths contain `/domain/`.
- **Rename ambiguous functions**: `isRepoPackage` → `isAnyRepoPackage`, `isRepoPackageByPath` → `isCoreRepoPackage` for clarity on their different matching scopes.

### tui/tree.go
- **Fix group node data leak**: Intermediate (non-leaf) tree nodes were incorrectly assigned the `Imports` and `FullPath` of the first leaf package processed. Now only leaf nodes get these fields.

### tui/detail.go
- **Remove `violWithPath` wrapper**: Eliminated unnecessary `violWithPath` struct that added no value over direct `rules.Violation` usage.

### tui/app.go
- **Distinguish error/warning counts in status bar**: Status bar now shows separate `errors` and `warnings` counts with appropriate colors instead of a single red `violations` count.

## Verification
- `go build ./...` — pass
- `go test ./...` — all pass
- `make lint` — 0 issues
