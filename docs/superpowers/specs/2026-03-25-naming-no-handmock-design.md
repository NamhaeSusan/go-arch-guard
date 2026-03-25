# Design: `naming.no-handmock` Rule

## Purpose

Detect hand-rolled mock implementations in `_test.go` files to enforce mockery usage.

When vibe coding with AI, the AI often ignores mockery and hand-rolls interface implementations in test files. This rule catches that pattern.

## Detection Logic

1. Iterate `_test.go` files via AST (`pkg.Syntax`)
2. Collect struct names defined in the file (`type Foo struct{...}`)
3. If the same file has a method receiver (`func (f *Foo) Bar()`) for that struct, flag as violation

### Violation Example

```go
// order_test.go — VIOLATION
type mockOrderRepo struct{ results []Order }
func (m *mockOrderRepo) FindByID(id string) (*Order, error) { return nil, nil }
```

### Pass Examples

```go
// order_test.go — struct only, no methods → pass
type testCase struct{ name string; want int }

// order_test.go — import mockery-generated mock → pass
import "example.com/mocks"
```

## Rule Output

- **Rule ID**: `naming.no-handmock`
- **Message**: `test file "order_test.go" defines hand-rolled mock "mockOrderRepo" with methods — use mockery instead`
- **Fix**: `generate mock with mockery and import from mocks/ package`
- **Severity**: Error (default), configurable via `WithSeverity`

## Placement

Added to `CheckNaming` in `rules/naming.go`, following existing patterns.

## Exclusions

Standard `WithExclude` pattern support. No special exceptions (no suite handling).

## Implementation Scope

- `rules/naming.go`: add `checkNoHandMock` function, wire into `CheckNaming`
- `testdata/invalid/`: add `_test.go` fixture with hand-rolled mock
- `testdata/valid/`: ensure existing test files pass (struct-only, no methods)
- `rules/naming_test.go`: add test cases
- `README.md`: document new rule
