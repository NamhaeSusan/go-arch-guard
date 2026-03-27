# Blast Radius Guard

## Summary
Added `AnalyzeBlastRadius()` — a new analysis function that computes coupling metrics (Ca, Ce, Instability, Transitive Dependents) for internal packages and emits Warning violations for statistical outliers using IQR-based detection. Zero configuration required.

## Files Changed
- `rules/blast.go` — graph construction, metrics computation, IQR outlier detection
- `rules/blast_test.go` — unit tests with blast testdata
- `rules/testhelpers_test.go` — added loadBlast helper
- `testdata/blast/` — dedicated test project with star topology (19 packages, pkg as hub)
- `integration_test.go` — added blast radius integration tests
- `example_test.go` — added blast radius to example
- `README.md` — API documentation
- `SKILL.md` — rule description

## Verification
- All existing tests pass (`go test ./...`)
- Blast testdata correctly identifies `internal/pkg` (9 transitive dependents) and `order/core/model` (6 transitive dependents) as high-coupling outliers
- Edge cases covered: small projects (<5 packages), excluded packages, severity override
- `make lint` passes with 0 issues
