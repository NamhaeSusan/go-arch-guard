# Design: `naming.no-handmock` Rule

## Purpose

Detect hand-rolled mock implementations in `_test.go` files to enforce mockery usage.

When vibe coding with AI, the AI often ignores mockery and hand-rolls interface implementations in test files. This rule catches that pattern.

## Detection Logic

1. For each loaded package, find `_test.go` files from `pkg.GoFiles`
2. Parse each `_test.go` with `go/parser.ParseFile` (since `pkg.Syntax` does not include test files — the loader does not set `Tests: true`)
3. Collect struct names whose name starts with `mock`, `fake`, or `stub` (case-insensitive)
4. If the same file has a method with a pointer or value receiver for that struct, flag as violation
5. Report violation at the struct definition line

### Violation Examples

```go
// order_test.go — VIOLATION: struct "mock" prefix + method receiver
type mockOrderRepo struct{ results []Order }
func (m *mockOrderRepo) FindByID(id string) (*Order, error) { return nil, nil }

// order_test.go — VIOLATION: value receiver also caught
type fakeNotifier struct{}
func (f fakeNotifier) Send(msg string) error { return nil }

// order_test.go — VIOLATION: func-field mock with methods
type mockReviewProvider struct {
    getReviewData func(ctx context.Context, reviewID string) (*ReviewData, error)
}
func (m *mockReviewProvider) GetReviewData(ctx context.Context, reviewID string) (*ReviewData, error) {
    return m.getReviewData(ctx, reviewID)
}
```

### Pass Examples

```go
// order_test.go — no "mock/fake/stub" prefix → pass
type testCase struct{ name string; want int }
func (tc testCase) run(t *testing.T) { ... }

// order_test.go — struct with prefix but no methods → pass
type mockData struct{ value string }
```

## Rule Output

- **Rule ID**: `naming.no-handmock`
- **Message**: `test file "order_test.go" defines hand-rolled mock "mockOrderRepo" with methods — use mockery instead`
- **Fix**: `generate mock with mockery and import from mocks/ package`
- **Severity**: Error (default), configurable via `WithSeverity`
- **Line**: struct definition line

## Placement

Added to `CheckNaming` in `rules/naming.go`, following existing patterns.

## Exclusions

Standard `WithExclude` pattern support. No special exceptions.

## Implementation Scope

- `rules/naming.go`: add `checkNoHandMock` function, wire into `CheckNaming`
- `testdata/invalid/`: add `_test.go` fixture with hand-rolled mock struct
- `testdata/valid/`: ensure existing test files pass
- `rules/naming_test.go`: add test cases
- `README.md`: document new rule
