# Quality Improvements Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix silent failure on wrong module path, add module auto-extraction, deduplicate test loads, and merge findImportFile/findImportLine.

**Architecture:** Add a validation helper that compares `projectModule` with loaded packages' module paths. Introduce `""` sentinel for auto-extraction from `pkgs`. Refactor helpers.go to combine file+line lookup. Refactor test files to use package-level shared fixtures.

**Tech Stack:** Go, `golang.org/x/tools/go/packages`

---

## File Structure

| File | Action | Responsibility |
|------|--------|----------------|
| `rules/helpers.go` | Modify | Add `findImportPosition`, `resolveModule`, module validation |
| `rules/isolation.go` | Modify | Use `findImportPosition`, add module validation call |
| `rules/layer.go` | Modify | Use `findImportPosition`, add module validation call |
| `rules/helpers_test.go` | Create | Unit tests for new helpers |
| `rules/isolation_test.go` | Modify | Share package loading, add module mismatch test |
| `rules/layer_test.go` | Modify | Share package loading, add module mismatch test |
| `rules/naming_test.go` | Modify | Share package loading |
| `integration_test.go` | Modify | Add module mismatch integration test |
| `README.md` | Modify | Document auto-extraction feature |

---

### Task 1: Merge `findImportFile` + `findImportLine` into `findImportPosition`

**Files:**
- Modify: `rules/helpers.go:10-47`
- Create: `rules/helpers_test.go`
- Modify: `rules/isolation.go` (all `findImportFile`/`findImportLine` call sites)
- Modify: `rules/layer.go` (all `findImportFile`/`findImportLine` call sites)

- [ ] **Step 1: Write the failing test for `findImportPosition`**

```go
// rules/helpers_test.go
package rules

import (
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/analyzer"
)

func TestFindImportPosition(t *testing.T) {
	pkgs, err := analyzer.Load("../testdata/invalid", "internal/...")
	if err != nil {
		t.Fatal(err)
	}

	// Find the order/app package which imports core/model
	var appPkg *packages.Package
	for _, pkg := range pkgs {
		if pkg.PkgPath == "github.com/kimtaeyun/testproject-dc-invalid/internal/domain/order/app" {
			appPkg = pkg
			break
		}
	}
	if appPkg == nil {
		t.Fatal("order/app package not found")
	}

	importPath := "github.com/kimtaeyun/testproject-dc-invalid/internal/domain/order/core/model"
	file, line := findImportPosition(appPkg, importPath, "../testdata/invalid")

	if file == "" {
		t.Error("expected non-empty file")
	}
	if line == 0 {
		t.Error("expected non-zero line")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./rules/ -run TestFindImportPosition -v`
Expected: FAIL — `findImportPosition` undefined

- [ ] **Step 3: Implement `findImportPosition` in helpers.go**

Replace `findImportFile` and `findImportLine` with:

```go
func findImportPosition(pkg *packages.Package, importPath, projectRoot string) (string, int) {
	absRoot, _ := filepath.Abs(projectRoot)
	fset := pkg.Fset
	for _, file := range pkg.Syntax {
		for _, imp := range file.Imports {
			path := strings.Trim(imp.Path.Value, `"`)
			if path == importPath {
				pos := fset.Position(imp.Pos())
				rel, err := filepath.Rel(absRoot, pos.Filename)
				if err != nil {
					return pos.Filename, pos.Line
				}
				return filepath.ToSlash(rel), pos.Line
			}
		}
	}
	if len(pkg.GoFiles) > 0 {
		rel, err := filepath.Rel(absRoot, pkg.GoFiles[0])
		if err != nil {
			return pkg.GoFiles[0], 0
		}
		return filepath.ToSlash(rel), 0
	}
	return pkg.PkgPath, 0
}
```

Remove `findImportFile` and `findImportLine`.

- [ ] **Step 4: Update all call sites in `isolation.go`**

Replace all pairs of:
```go
File: findImportFile(pkg, impPath, projectRoot),
Line: findImportLine(pkg, impPath),
```
With:
```go
File: file, Line: line,
```
Where `file, line := findImportPosition(pkg, impPath, projectRoot)` is called once before the Violation literal.

- [ ] **Step 5: Update all call sites in `layer.go`**

Same pattern as Step 4 for `layer.go`.

- [ ] **Step 6: Run all tests**

Run: `go test ./... -v`
Expected: ALL PASS

- [ ] **Step 7: Commit**

```bash
git add rules/helpers.go rules/helpers_test.go rules/isolation.go rules/layer.go
git commit -m "refactor: merge findImportFile and findImportLine into findImportPosition"
```

---

### Task 2: Add module validation warning (P0)

**Files:**
- Modify: `rules/helpers.go` — add `validateModule` function
- Modify: `rules/isolation.go:10` — call validation at start
- Modify: `rules/layer.go:11` — call validation at start
- Modify: `rules/isolation_test.go` — add mismatch test
- Modify: `rules/layer_test.go` — add mismatch test

- [ ] **Step 1: Write failing test for module mismatch warning**

Add to `rules/isolation_test.go`:

```go
t.Run("warns when module path matches no packages", func(t *testing.T) {
	pkgs, err := analyzer.Load("../testdata/valid", "internal/...")
	if err != nil {
		t.Fatal(err)
	}
	violations := rules.CheckDomainIsolation(pkgs, "github.com/wrong/module", "../testdata/valid")
	found := false
	for _, v := range violations {
		if v.Rule == "meta.no-matching-packages" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected meta.no-matching-packages warning for wrong module path")
	}
})
```

Add same test to `rules/layer_test.go` for `CheckLayerDirection`.

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./rules/ -run "warns_when_module" -v`
Expected: FAIL

- [ ] **Step 3: Implement `validateModule` in helpers.go**

```go
func validateModule(pkgs []*packages.Package, projectModule string) []Violation {
	if projectModule == "" {
		return nil
	}
	prefix := projectModule + "/"
	for _, pkg := range pkgs {
		if pkg.PkgPath == projectModule || strings.HasPrefix(pkg.PkgPath, prefix) {
			return nil
		}
	}
	return []Violation{{
		Rule:     "meta.no-matching-packages",
		Message:  fmt.Sprintf("module %q does not match any loaded package — all import checks will be skipped", projectModule),
		Fix:      "verify the module argument matches go.mod (e.g. pass the value from pkgs[0].Module.Path)",
		Severity: Warning,
	}}
}
```

- [ ] **Step 4: Call `validateModule` at the top of `CheckDomainIsolation`**

```go
func CheckDomainIsolation(pkgs []*packages.Package, projectModule string, projectRoot string, opts ...Option) []Violation {
	cfg := NewConfig(opts...)
	if warns := validateModule(pkgs, projectModule); len(warns) > 0 {
		return warns
	}
	// ... rest unchanged
```

- [ ] **Step 5: Call `validateModule` at the top of `CheckLayerDirection`**

Same pattern.

- [ ] **Step 6: Run all tests**

Run: `go test ./... -v`
Expected: ALL PASS

- [ ] **Step 7: Commit**

```bash
git add rules/helpers.go rules/isolation.go rules/layer.go rules/isolation_test.go rules/layer_test.go
git commit -m "feat: warn when projectModule matches no loaded packages"
```

---

### Task 3: Add module auto-extraction from packages (P1)

**Files:**
- Modify: `rules/helpers.go` — add `resolveModule` and `resolveRoot`
- Modify: `rules/isolation.go:10` — use `resolveModule`/`resolveRoot`
- Modify: `rules/layer.go:11` — use `resolveModule`/`resolveRoot`
- Create or modify: `rules/helpers_test.go` — test resolveModule
- Modify: `rules/isolation_test.go` — add auto-extraction test

- [ ] **Step 1: Write failing test for `resolveModule`**

Add to `rules/helpers_test.go`:

```go
func TestResolveModule(t *testing.T) {
	pkgs, err := analyzer.Load("../testdata/valid", "internal/...")
	if err != nil {
		t.Fatal(err)
	}
	mod := resolveModule(pkgs, "")
	if mod != "github.com/kimtaeyun/testproject-dc" {
		t.Errorf("got %q, want github.com/kimtaeyun/testproject-dc", mod)
	}

	// explicit value passes through
	mod = resolveModule(pkgs, "custom/module")
	if mod != "custom/module" {
		t.Errorf("got %q, want custom/module", mod)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./rules/ -run TestResolveModule -v`
Expected: FAIL — `resolveModule` undefined

- [ ] **Step 3: Implement `resolveModule` and `resolveRoot` in helpers.go**

```go
func resolveModule(pkgs []*packages.Package, explicit string) string {
	if explicit != "" {
		return explicit
	}
	for _, pkg := range pkgs {
		if pkg.Module != nil && pkg.Module.Path != "" {
			return pkg.Module.Path
		}
	}
	return ""
}

func resolveRoot(pkgs []*packages.Package, explicit string) string {
	if explicit != "" {
		return explicit
	}
	for _, pkg := range pkgs {
		if pkg.Module != nil && pkg.Module.Dir != "" {
			return pkg.Module.Dir
		}
	}
	return ""
}
```

- [ ] **Step 4: Wire into `CheckDomainIsolation`**

```go
func CheckDomainIsolation(pkgs []*packages.Package, projectModule string, projectRoot string, opts ...Option) []Violation {
	cfg := NewConfig(opts...)
	projectModule = resolveModule(pkgs, projectModule)
	projectRoot = resolveRoot(pkgs, projectRoot)
	if warns := validateModule(pkgs, projectModule); len(warns) > 0 {
		return warns
	}
	// ... rest unchanged
```

- [ ] **Step 5: Wire into `CheckLayerDirection`**

Same pattern.

- [ ] **Step 6: Write integration test for auto-extraction**

Add to `rules/isolation_test.go`:

```go
t.Run("auto-extracts module when empty string passed", func(t *testing.T) {
	pkgs, err := analyzer.Load("../testdata/valid", "internal/...", "cmd/...")
	if err != nil {
		t.Fatal(err)
	}
	violations := rules.CheckDomainIsolation(pkgs, "", "")
	if len(violations) > 0 {
		for _, v := range violations {
			t.Log(v.String())
		}
		t.Errorf("expected no violations with auto-extracted module, got %d", len(violations))
	}
})
```

- [ ] **Step 7: Run all tests**

Run: `go test ./... -v`
Expected: ALL PASS

- [ ] **Step 8: Commit**

```bash
git add rules/helpers.go rules/helpers_test.go rules/isolation.go rules/layer.go rules/isolation_test.go
git commit -m "feat: auto-extract module and root from packages when empty string passed"
```

---

### Task 4: Deduplicate test `analyzer.Load` calls (P2)

**Files:**
- Modify: `rules/isolation_test.go`
- Modify: `rules/layer_test.go`
- Modify: `rules/naming_test.go`

- [ ] **Step 1: Refactor `isolation_test.go`**

Add package-level shared loader at top of file:

```go
var (
	loadInvalidOnce  sync.Once
	invalidPkgs      []*packages.Package
	loadInvalidErr   error

	loadValidOnce    sync.Once
	validPkgs        []*packages.Package
	loadValidErr     error
)

func loadValid(t *testing.T) []*packages.Package {
	t.Helper()
	loadValidOnce.Do(func() {
		validPkgs, loadValidErr = analyzer.Load("../testdata/valid", "internal/...", "cmd/...")
	})
	if loadValidErr != nil {
		t.Fatal(loadValidErr)
	}
	return validPkgs
}

func loadInvalid(t *testing.T) []*packages.Package {
	t.Helper()
	loadInvalidOnce.Do(func() {
		invalidPkgs, loadInvalidErr = analyzer.Load("../testdata/invalid", "internal/...", "cmd/...")
	})
	if loadInvalidErr != nil {
		t.Fatal(loadInvalidErr)
	}
	return invalidPkgs
}
```

Since all three test files are in the same `rules_test` package, define these helpers **once** in a shared file.

- [ ] **Step 2: Create `rules/testhelpers_test.go`**

Move the shared `loadValid`/`loadInvalid` helpers into this file so all test files can use them.

- [ ] **Step 3: Update `isolation_test.go` to use shared loaders**

Replace every:
```go
pkgs, err := analyzer.Load("../testdata/invalid", "internal/...")
if err != nil {
    t.Fatal(err)
}
```
With:
```go
pkgs := loadInvalid(t)
```

Note: some tests load with `"cmd/..."` too. Since `loadInvalid` already includes `"cmd/..."`, this is safe — more packages loaded is fine, the rules filter by prefix.

- [ ] **Step 4: Update `layer_test.go` to use shared loaders**

Same replacement pattern.

- [ ] **Step 5: Update `naming_test.go` to use shared loaders**

Same replacement pattern.

- [ ] **Step 6: Run all tests**

Run: `go test ./... -v`
Expected: ALL PASS

- [ ] **Step 7: Commit**

```bash
git add rules/testhelpers_test.go rules/isolation_test.go rules/layer_test.go rules/naming_test.go
git commit -m "refactor: deduplicate analyzer.Load calls in rule tests"
```

---

### Task 5: Update README and documentation

**Files:**
- Modify: `README.md` — document auto-extraction, module validation warning
- Create: `claude_history/2026-03-24-quality-improvements.md`

- [ ] **Step 1: Update README Quick Start with auto-extraction option**

Add a note after the existing Quick Start showing the simplified API:

```markdown
### Simplified Usage

If you prefer not to hard-code the module path and root, pass empty strings
to auto-extract them from the loaded packages:

\```go
violations := rules.CheckDomainIsolation(pkgs, "", "")
\```
```

- [ ] **Step 2: Add `meta.no-matching-packages` to the Rules section**

Under a new "Diagnostics" sub-heading:

```markdown
### Diagnostics

| Rule | Meaning |
|------|---------|
| `meta.no-matching-packages` | the `projectModule` argument does not match any loaded package — usually a misconfiguration |
```

- [ ] **Step 3: Write work log**

Create `claude_history/2026-03-24-quality-improvements.md`.

- [ ] **Step 4: Run lint**

Run: `make lint`
Expected: 0 issues

- [ ] **Step 5: Commit**

```bash
git add README.md claude_history/2026-03-24-quality-improvements.md
git commit -m "docs: document module auto-extraction and diagnostics rule"
```
