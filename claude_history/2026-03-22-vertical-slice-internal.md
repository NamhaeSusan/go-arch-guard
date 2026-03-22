# Vertical Slice Internal Layer Direction Rule

## Task
Implement `CheckVerticalSliceInternal` — a SOFT rule that enforces intra-domain layer direction within vertical slices.

## Files Changed
- `rules/vertical_internal.go` — new file with `CheckVerticalSliceInternal` and `identifySublayer` helper
- `rules/vertical_test.go` — added `TestCheckVerticalSliceInternal` with valid/invalid test cases
- `README.md` — added documentation for the new rule

## Implementation
- `allowedInternalImports` map defines which sublayers can import which
- `identifySublayer` extracts the first path segment after `internal/<domain>/`
- Cross-domain and shared imports are skipped (handled by `CheckVerticalSlice`)
- Same-sublayer imports are always allowed
- Violation rule name: `vertical.internal-layer-direction`

## Verification
- `go test ./...` — all tests pass
- `go vet ./...` — no issues
- `make lint` — 0 issues
