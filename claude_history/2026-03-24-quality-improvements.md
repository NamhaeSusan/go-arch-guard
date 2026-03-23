# Quality Improvements

## Summary

Four quality improvements across the codebase:

1. **P0: Module validation warning** — `CheckDomainIsolation` and `CheckLayerDirection` now emit a `meta.no-matching-packages` warning when the `projectModule` argument matches no loaded package (previously all checks were silently skipped).

2. **P1: Module/root auto-extraction** — Passing `""` for `module` and `root` parameters auto-extracts values from the loaded packages' `Module` metadata. Eliminates the need to hard-code module paths.

3. **P2: Merged findImportFile + findImportLine** — Combined into a single `findImportPosition` function that returns `(file, line)` in one AST traversal instead of two.

4. **P2: Deduplicated test loads** — Created shared `loadValid`/`loadInvalid` helpers using `sync.Once` in `testhelpers_test.go`. Replaced ~24 redundant `analyzer.Load` calls across `isolation_test.go`, `layer_test.go`, and `naming_test.go`.

## Files Changed

- `rules/helpers.go` — added `findImportPosition`, `resolveModule`, `resolveRoot`, `validateModule`; removed `findImportFile`, `findImportLine`
- `rules/isolation.go` — updated call sites, added resolve/validate calls
- `rules/layer.go` — updated call sites, added resolve/validate calls
- `rules/helpers_test.go` (new) — internal tests for new helpers
- `rules/testhelpers_test.go` (new) — shared test loaders with sync.Once
- `rules/isolation_test.go` — deduplicated loads, added module mismatch and auto-extraction tests
- `rules/layer_test.go` — deduplicated loads, added module mismatch test
- `rules/naming_test.go` — deduplicated loads
- `README.md` — documented simplified usage, diagnostics rule, updated API reference

## Verification

- `go test ./... -count=1` — all pass
- `go build ./...` — clean
