# Dead Code Cleanup & Coverage Improvement

## Task
Remove dead code branches and improve test coverage for uncovered functions.

## Changes

### Dead Code Removed
- **rules/isolation.go**: Removed two unreachable branches in the orchestration import check block:
  - `if srcIsOrchestration { continue }` — unreachable because orchestration source packages are fully handled earlier (lines 72-92) with continue
  - `if srcIsPkg { continue }` — unreachable because pkg packages importing orchestration are handled earlier (lines 107-117)
- **rules/structure.go**: Simplified `checkDTOPlacement` by removing the `"infra"` walk. `internal/infra/` cannot exist because `structure.internal-top-level` already blocks it. All actual infra directories live inside domain slices and are covered by the `"domain"` walk.

### Tests Added
- **rules/naming_test.go**: Added test for camelCase filename detection (`createOrder.go`) that validates the fix message contains the snake_case suggestion (`create_order.go`), exercising the previously untested `toSnakeCase` function
- **rules/structure_test.go**: Added 3 tests for the Go-file-at-internal-root branch of `checkInternalTopLevelPackages`:
  - Rejects `.go` files at `internal/` top level
  - Ignores `_test.go` files at `internal/` top level
  - Ignores non-Go files at `internal/` top level

## Coverage Impact
- `toSnakeCase`: 0% → 100%
- `checkInternalTopLevelPackages`: 68.4% → 89.5%
- Total: 86.0% → 88.9%

## Verification
- All tests pass (`go test ./...`)
- Lint clean (`make lint`)
