## Summary

- Added a new `structure.internal-top-level` rule so only `internal/domain`, `internal/orchestration`, and `internal/pkg` are allowed at the `internal/` top level.
- Replaced the previous integration proof that neutral support packages were allowed with regression coverage showing `internal/config`, `internal/platform`, `internal/system`, and `internal/foundation` are now rejected.
- Updated README wording to state the top-level whitelist explicitly and align orchestration guidance with the stricter structure rule.

## Files Changed

- `rules/structure.go`
- `rules/structure_test.go`
- `integration_test.go`
- `README.md`
- `docs/superpowers/plans/2026-03-24-internal-top-level-whitelist.md`
- `claude_history/2026-03-24-internal-top-level-whitelist.md`

## Verification

- `go test ./... -run 'Test(CheckStructure|Integration_RejectsUnexpectedInternalTopLevelPackages)' -v`
- `make lint`
- `go test ./...`
