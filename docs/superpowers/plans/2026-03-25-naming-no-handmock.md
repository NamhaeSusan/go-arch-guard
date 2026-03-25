# naming.no-handmock Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Detect hand-rolled mock/fake/stub structs with methods in `_test.go` files, enforcing mockery usage.

**Architecture:** New check function `checkNoHandMock` added to `CheckNaming` in `rules/naming.go`. Since the loader does not set `Tests: true`, neither `pkg.Syntax` nor `pkg.GoFiles` include `_test.go` files. The function derives the package directory from `pkg.GoFiles[0]`, globs for `*_test.go`, and parses each with `go/parser.ParseFile`. Detection is scoped to structs whose name starts with `mock`, `fake`, or `stub` (case-insensitive) that also have pointer or value method receivers in the same file.

**Tech Stack:** Go `go/ast`, `go/parser`, `go/token`, `filepath.Glob`

**Spec:** `docs/superpowers/specs/2026-03-25-naming-no-handmock-design.md`

---

## File Structure

| Action | File | Responsibility |
|--------|------|----------------|
| Modify | `rules/naming.go` | Add `checkNoHandMock`, `collectMockStructs`, `receiverTypeName`, wire into `CheckNaming` |
| Create | `testdata/invalid/internal/domain/order/app/service_test.go` | Fixture: hand-rolled mock struct with methods |
| Create | `testdata/valid/internal/domain/order/app/service_test.go` | Fixture: clean test file (no hand-rolled mocks) |
| Modify | `rules/naming_test.go` | Add test cases for the new rule |
| Modify | `README.md` | Document `naming.no-handmock` in the Naming rules table |

---

### Task 1: Add testdata fixtures

**Files:**
- Create: `testdata/invalid/internal/domain/order/app/service_test.go`
- Create: `testdata/valid/internal/domain/order/app/service_test.go`

- [ ] **Step 1: Create the invalid fixture file**

```go
// testdata/invalid/internal/domain/order/app/service_test.go
package app

import "context"

type mockOrderRepo struct {
	findByID func(ctx context.Context, id string) (string, error)
}

func (m *mockOrderRepo) FindByID(ctx context.Context, id string) (string, error) {
	return m.findByID(ctx, id)
}
```

Hand-rolled mock struct (`mock` prefix) with a pointer receiver method — the rule must flag this.

- [ ] **Step 2: Create the valid fixture file**

```go
// testdata/valid/internal/domain/order/app/service_test.go
package app

type testInput struct {
	userID string
	amount int
}
```

Plain struct, no `mock`/`fake`/`stub` prefix, no methods. Must pass the rule AND all existing naming rules (snake_case filename, no stutter, no impl suffix — all clean).

- [ ] **Step 3: Verify fixtures compile**

Run: `cd /Users/kimtaeyun/go-arch-guard/testdata/invalid && go vet ./internal/domain/order/app/`
Expected: no errors from the new file

Run: `cd /Users/kimtaeyun/go-arch-guard/testdata/valid && go vet ./internal/domain/order/app/`
Expected: no errors

- [ ] **Step 4: Commit**

```bash
git add testdata/invalid/internal/domain/order/app/service_test.go \
       testdata/valid/internal/domain/order/app/service_test.go
git commit -m "test: add testdata fixtures for naming.no-handmock rule"
```

---

### Task 2: Implement `checkNoHandMock` with tests

Tests and implementation in one task to avoid a broken-CI commit.

**Files:**
- Modify: `rules/naming.go`
- Modify: `rules/naming_test.go`

- [ ] **Step 1: Add tests to `rules/naming_test.go`**

Add inside `TestCheckNaming`:

```go
t.Run("detects hand-rolled mock in test file", func(t *testing.T) {
	pkgs := loadInvalid(t)
	violations := rules.CheckNaming(pkgs)
	found := false
	for _, v := range violations {
		if v.Rule == "naming.no-handmock" {
			found = true
			if !strings.Contains(v.Message, "mockOrderRepo") {
				t.Errorf("expected message to mention mockOrderRepo, got %q", v.Message)
			}
			break
		}
	}
	if !found {
		t.Error("expected naming.no-handmock violation for hand-rolled mock")
	}
})

t.Run("valid project has no handmock violations", func(t *testing.T) {
	pkgs := loadValid(t)
	violations := rules.CheckNaming(pkgs)
	for _, v := range violations {
		if v.Rule == "naming.no-handmock" {
			t.Errorf("unexpected naming.no-handmock violation: %s", v.String())
		}
	}
})

t.Run("exclude skips handmock check", func(t *testing.T) {
	pkgs := loadInvalid(t)
	violations := rules.CheckNaming(pkgs, rules.WithExclude("internal/domain/order/app/..."))
	for _, v := range violations {
		if v.Rule == "naming.no-handmock" {
			t.Errorf("expected handmock violation to be excluded, got %s", v.String())
		}
	}
})
```

- [ ] **Step 2: Add imports and wire into `CheckNaming`**

Add `"go/parser"`, `"go/token"`, `"path/filepath"` (if not present) to the import block in `rules/naming.go`.

In `CheckNaming`, after the `checkHandlerNoExportedInterface` call, add:

```go
violations = append(violations, checkNoHandMock(pkg, cfg)...)
```

- [ ] **Step 3: Implement `checkNoHandMock`**

Add to `rules/naming.go`:

```go
var handMockPrefixes = []string{"mock", "fake", "stub"}

func checkNoHandMock(pkg *packages.Package, cfg Config) []Violation {
	if len(pkg.GoFiles) == 0 {
		return nil
	}
	pkgDir := filepath.Dir(pkg.GoFiles[0])
	testFiles, err := filepath.Glob(filepath.Join(pkgDir, "*_test.go"))
	if err != nil || len(testFiles) == 0 {
		return nil
	}

	var violations []Violation
	fset := token.NewFileSet()
	seen := make(map[string]bool)
	for _, f := range testFiles {
		if seen[f] {
			continue
		}
		seen[f] = true
		relPath := relativePathForPackage(pkg, f)
		if cfg.IsExcluded(relPath) {
			continue
		}
		astFile, err := parser.ParseFile(fset, f, nil, 0)
		if err != nil {
			continue
		}
		structs := collectMockStructs(fset, astFile)
		if len(structs) == 0 {
			continue
		}
		baseName := filepath.Base(f)
		for _, decl := range astFile.Decls {
			fd, ok := decl.(*ast.FuncDecl)
			if !ok || fd.Recv == nil || len(fd.Recv.List) == 0 {
				continue
			}
			recvName := receiverTypeName(fd.Recv.List[0].Type)
			if line, ok := structs[recvName]; ok {
				violations = append(violations, Violation{
					File:     relPath,
					Line:     line,
					Rule:     "naming.no-handmock",
					Message:  `test file "` + baseName + `" defines hand-rolled mock "` + recvName + `" with methods — use mockery instead`,
					Fix:      "generate mock with mockery and import from mocks/ package",
					Severity: cfg.Sev,
				})
				delete(structs, recvName)
			}
		}
	}
	return violations
}

func collectMockStructs(fset *token.FileSet, file *ast.File) map[string]int {
	result := make(map[string]int)
	for _, decl := range file.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range gd.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			if _, ok := ts.Type.(*ast.StructType); !ok {
				continue
			}
			name := ts.Name.Name
			lower := strings.ToLower(name)
			for _, prefix := range handMockPrefixes {
				if strings.HasPrefix(lower, prefix) {
					result[name] = fset.Position(ts.Name.Pos()).Line
					break
				}
			}
		}
	}
	return result
}

func receiverTypeName(expr ast.Expr) string {
	if star, ok := expr.(*ast.StarExpr); ok {
		expr = star.X
	}
	if ident, ok := expr.(*ast.Ident); ok {
		return ident.Name
	}
	return ""
}
```

- [ ] **Step 4: Run tests**

Run: `cd /Users/kimtaeyun/go-arch-guard && go test ./rules/ -run TestCheckNaming -v`
Expected: PASS

- [ ] **Step 5: Run full test suite**

Run: `cd /Users/kimtaeyun/go-arch-guard && go test ./... -v`
Expected: PASS (no regressions — the existing `"valid project has no violations"` test must still pass)

- [ ] **Step 6: Commit**

```bash
git add rules/naming.go rules/naming_test.go
git commit -m "feat: add naming.no-handmock rule to detect hand-rolled mocks in test files"
```

---

### Task 3: Update README and SKILL.md

**Files:**
- Modify: `README.md`
- Modify: `SKILL.md`

- [ ] **Step 1: Add rule to the Naming table in README**

In the `### Naming` section, add a row to the table:

```markdown
| `naming.no-handmock` | test file defines a hand-rolled mock/fake/stub struct with methods |
```

- [ ] **Step 2: Update SKILL.md if naming rules are listed there**

Check SKILL.md for a naming rules section and add `naming.no-handmock` if applicable.

- [ ] **Step 3: Run lint**

Run: `cd /Users/kimtaeyun/go-arch-guard && make lint`
Expected: 0 issues

- [ ] **Step 4: Commit**

```bash
git add README.md SKILL.md
git commit -m "docs: document naming.no-handmock rule"
```

---

### Task 4: Write work log

**Files:**
- Create: `claude_history/2026-03-25-naming-no-handmock.md`

- [ ] **Step 1: Write work log**

Document: task summary, files changed, verification results.

- [ ] **Step 2: Commit**

```bash
git add claude_history/2026-03-25-naming-no-handmock.md
git commit -m "docs: add work log for naming.no-handmock

History: claude_history/2026-03-25-naming-no-handmock.md"
```
