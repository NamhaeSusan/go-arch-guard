# Rename saga to orchestration

## Task
Rename all occurrences of "saga" to "orchestration" across the entire project.

## Changes

### Source code
- `rules/isolation.go`: renamed `isSagaPkg` -> `isOrchestrationPkg`, `isSagaHandler` -> `isOrchestrationHandler`, rule `isolation.saga-deep-import` -> `isolation.orchestration-deep-import`, updated all comments and error messages
- `rules/layer.go`: updated comment and function call from `isSagaPkg` to `isOrchestrationPkg`
- `rules/isolation_test.go`: updated test name and assertions

### Testdata
- Renamed directories `testdata/valid/internal/saga/` -> `orchestration/` and `testdata/invalid/internal/saga/` -> `orchestration/`
- Updated all package declarations, import paths, type names, and function names

### Documentation
- `README.md`: replaced all saga references with orchestration
- `claude_history/2026-03-23-dc-testdata.md`: updated saga references

## Verification
- `go build ./...` passes
- `go test ./... -count=1` passes (all 4 packages)
- `make lint` passes (0 issues)
- `grep -ri saga` returns 0 matches
