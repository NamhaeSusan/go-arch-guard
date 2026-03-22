# Add CheckVerticalSlice for cross-domain isolation

## Task
Implemented `CheckVerticalSlice` — HARD rule enforcing cross-domain isolation in vertical slice architecture.

## Files changed
- `rules/vertical.go` — New file with `CheckVerticalSlice`, `identifyVerticalDomain`, `isUsecasePkg`, `isAliasOrPort`
- `rules/vertical_test.go` — Tests: valid project, cross-domain violation detection, usecase exception
- `README.md` — Added Vertical Slice rule documentation

## Verification
- `go test ./...` — all pass
- `go vet ./...` — clean
- `make lint` — 0 issues
