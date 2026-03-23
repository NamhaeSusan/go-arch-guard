# Extract helpers, remove old dependency/vertical rules

## Task
Prepare codebase for domain-centric rule rewrite by extracting shared helpers and removing old rule implementations.

## Changes

### Created
- `rules/helpers.go` — extracted `findImportFile` and `findImportLine` from `dependency.go`

### Deleted
- `rules/dependency.go`, `rules/dependency_test.go` — old dependency rule
- `rules/vertical.go`, `rules/vertical_internal.go`, `rules/vertical_test.go` — old vertical slice rules
- `testdata/valid`, `testdata/invalid` — old testdata for dependency/naming/structure tests
- `testdata/vertical-valid`, `testdata/vertical-invalid` — old testdata for vertical tests

### Modified
- `integration_test.go` — cleared (placeholder, will be rewritten)
- `example_test.go` — cleared (placeholder, will be rewritten)
- `analyzer/loader_test.go` — removed test that depended on deleted testdata
- `rules/naming_test.go` — cleared (depended on deleted testdata)
- `rules/structure_test.go` — cleared (depended on deleted testdata)

## Verification
- `go build ./...` passes
- `go test ./... -count=1` all pass
- `make lint` passes (0 issues)
