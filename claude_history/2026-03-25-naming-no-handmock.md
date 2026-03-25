# naming.no-handmock rule

## Summary

Added `naming.no-handmock` rule to detect hand-rolled mock/fake/stub structs with methods in `_test.go` files, enforcing mockery usage.

## Detection Logic

- Derives package directory from `pkg.GoFiles[0]`, globs for `*_test.go`
- Parses each test file with `go/parser.ParseFile` (since loader doesn't include test ASTs)
- Flags structs with `mock`/`fake`/`stub` prefix (case-insensitive) that have pointer or value method receivers in the same file

## Files Changed

- `rules/naming.go` — added `checkNoHandMock`, `collectMockStructs`, `receiverTypeName`
- `rules/naming_test.go` — 3 new test cases (detect, valid-pass, exclude)
- `testdata/invalid/internal/domain/order/app/service_test.go` — invalid fixture
- `testdata/valid/internal/domain/order/app/service_test.go` — valid fixture
- `README.md` — documented rule in Naming table
- `SKILL.md` — added hand-rolled mocks to Banned Patterns
- `CLAUDE.md` — added SKILL.md sync requirement

## Verification

- `go test ./rules/ -run TestCheckNaming -v` — 8/8 PASS
- `go test ./...` — all packages PASS
- `make lint` — 0 issues
