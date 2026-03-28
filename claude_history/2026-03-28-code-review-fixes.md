# 2026-03-28 Code Review Fixes

## Task
Code review of the entire codebase and fix identified issues.

## Changes (Commit 1)

### rules/naming.go
- **Fix `stutters()` UTF-8 safety**: Changed byte-indexing to rune-based indexing
- **Fix `stutters()` suggested name bug**: Preserve original casing instead of lowercasing
- **Fix `isDomainPackage` scope**: `/domain/` → `/internal/domain/` to avoid false positives
- **Rename ambiguous functions**: `isRepoPackage` → `isAnyRepoPackage`, `isRepoPackageByPath` → `isCoreRepoPackage`

### tui/tree.go
- **Fix group node data leak**: Only leaf nodes get `Imports`/`FullPath`

### tui/detail.go
- **Remove `violWithPath` wrapper**: Direct `rules.Violation` usage

### tui/app.go
- **Separate error/warning counts in status bar**

## Changes (Commit 2)

### rules/structure.go
- **Separate `structure.misplaced-layer` rule** from `structure.legacy-package`

### rules/helpers.go
- **Extract `resolveIdentImportPath` helper** for shared alias import resolution

### rules/naming.go, rules/structure.go
- **Use shared `resolveIdentImportPath` helper**

### rules/rule.go
- **Use `strings.HasPrefix`/`strings.TrimRight`** in `matchPattern`

### README.md, README.ko.md
- **Document new `structure.misplaced-layer` rule**

## Changes (Commit 3)

### integration_test.go
- **Add `structure.misplaced-layer` to integration test** rule surface check
- **Improve `assertHasRule` error messages**: Show actual rule set on failure

## Verification
- `go build ./...` — pass
- `go test ./...` — all pass
- `make lint` — 0 issues
