# 2026-03-23 project-relative-excludes

## Summary

- Rewrote `README.md` so orchestration semantics match the actual rule scope: orchestration may use non-domain internal helpers, but it may not deep-import domains.
- Standardized `WithExclude` matching around project-relative paths only.
- Removed package-level support for module-qualified exclude paths.
- Added regression tests for project-relative exclude behavior and for rejecting module-qualified excludes.

## Files Changed

- `README.md`
- `rules/helpers.go`
- `rules/isolation_test.go`
- `rules/layer_test.go`
- `rules/naming_test.go`
- `rules/rule.go`
- `rules/structure_test.go`

## Verification

- `go test ./...`
- `make lint`
