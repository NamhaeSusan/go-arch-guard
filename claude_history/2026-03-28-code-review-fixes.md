# 2026-03-28 Code Review Fixes

## Task
Comprehensive code review of the entire go-arch-guard codebase across 5 iterations.

## Changes (Commit 1: fix)
- Fix `stutters()` UTF-8 safety (rune-based indexing) and suggested name casing bug
- Narrow `isDomainPackage` to `/internal/domain/`
- Rename `isRepoPackage`/`isRepoPackageByPath` for clarity
- Fix group tree node incorrectly inheriting leaf package data
- Remove unnecessary `violWithPath` wrapper in detail.go
- Separate error/warning counts in TUI status bar

## Changes (Commit 2: refactor)
- Separate `structure.misplaced-layer` rule from `structure.legacy-package`
- Extract `resolveIdentImportPath` shared helper
- Use `strings.HasPrefix`/`strings.TrimRight` in `matchPattern`
- Document new `structure.misplaced-layer` rule in README

## Changes (Commit 3: test)
- Assert `structure.misplaced-layer` in integration test
- Improve `assertHasRule` failure message with actual rule set

## Changes (Commit 4: test)
- Add missing `len(pkgs)==0` guard in `TestSkill_CrossDomainViolation`
- Add structure exclude verification in `TestSkill_ExcludeOption`

## Changes (Commit 5: style)
- Use `strings.HasPrefix` in `violations.go` walkPath for consistency

## Verification
- `go build ./...` — pass
- `go test ./...` — all pass
- `make lint` — 0 issues
