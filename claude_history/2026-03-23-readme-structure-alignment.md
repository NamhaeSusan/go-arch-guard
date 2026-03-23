# 2026-03-23 readme-structure-alignment

## Summary

- Tightened `structure.middleware-placement` so only `internal/pkg/middleware/` is accepted.
- Added regression coverage for nested `pkg/middleware` paths and for `core/model/` trees that only contain nested Go files.
- Updated `README.md` and rule messages to describe the exact middleware location and the direct-file `core/model/` requirement.

## Files Changed

- `README.md`
- `rules/structure.go`
- `rules/structure_test.go`
- `docs/superpowers/plans/2026-03-23-readme-structure-alignment.md`

## Verification

- `go test ./rules -run 'TestCheckStructure/detects_middleware_nested_under_non-root_pkg_path' -v`
- `go test ./rules -run 'TestCheckStructure/(detects_middleware_nested_under_non-root_pkg_path|detects_nested-only_core_model_files)' -v`
- `go test ./rules -run TestCheckStructure -v`
- `go test ./...`
- `make lint`
