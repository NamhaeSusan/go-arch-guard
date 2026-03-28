# 2026-03-28 Code Review Fixes

## Task
Code review of the entire codebase and fix identified issues.

## Changes (Commit 1: fix)

### rules/naming.go
- Fix `stutters()` UTF-8 safety (rune-based indexing)
- Fix suggested name preserving original casing
- Narrow `isDomainPackage` to `/internal/domain/`
- Rename `isRepoPackage`/`isRepoPackageByPath` for clarity

### tui/tree.go
- Fix group node incorrectly inheriting leaf package data

### tui/detail.go
- Remove unnecessary `violWithPath` wrapper

### tui/app.go
- Separate error/warning counts in status bar

## Changes (Commit 2: refactor)

### rules/structure.go
- Separate `structure.misplaced-layer` rule from `structure.legacy-package`

### rules/helpers.go
- Extract `resolveIdentImportPath` shared helper

### rules/naming.go, rules/structure.go
- Use shared `resolveIdentImportPath` helper

### rules/rule.go
- Use `strings.HasPrefix`/`strings.TrimRight` in `matchPattern`

### README.md, README.ko.md
- Document new `structure.misplaced-layer` rule

## Changes (Commit 3: test)

### integration_test.go
- Assert `structure.misplaced-layer` in integration test
- Improve `assertHasRule` failure message with actual rule set

## Changes (Commit 4: test)

### skill_test.go
- Add missing `len(pkgs) == 0` guard in `TestSkill_CrossDomainViolation`
- Add structure exclude verification in `TestSkill_ExcludeOption`

## Verification
- `go build ./...` — pass
- `go test ./...` — all pass
- `make lint` — 0 issues
