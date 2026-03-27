# Blast Radius Guard Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `AnalyzeBlastRadius()` that computes coupling metrics for internal packages and emits Warning violations for statistical outliers — zero configuration.

**Architecture:** Build a dependency graph from `[]*packages.Package`, compute Ca/Ce/Instability/TransitiveDependents per internal package, use IQR-based outlier detection to flag high-coupling packages. Same `[]Violation` return type as all other Check functions.

**Tech Stack:** Go, `golang.org/x/tools/go/packages`, standard library `sort`/`math`

---

### Task 1: Create testdata/blast project

**Files:**
- Create: `testdata/blast/go.mod`
- Create: `testdata/blast/internal/domain/order/core/model/order.go`
- Create: `testdata/blast/internal/domain/order/core/repo/repository.go`
- Create: `testdata/blast/internal/domain/order/core/svc/order.go`
- Create: `testdata/blast/internal/domain/order/app/service.go`
- Create: `testdata/blast/internal/domain/order/handler/http/handler.go`
- Create: `testdata/blast/internal/domain/order/event/events.go`
- Create: `testdata/blast/internal/domain/order/infra/persistence/store.go`
- Create: `testdata/blast/internal/domain/user/core/model/user.go`
- Create: `testdata/blast/internal/domain/user/core/repo/repository.go`
- Create: `testdata/blast/internal/domain/user/app/service.go`
- Create: `testdata/blast/internal/domain/user/handler/http/handler.go`
- Create: `testdata/blast/internal/domain/product/core/model/product.go`
- Create: `testdata/blast/internal/domain/product/app/service.go`
- Create: `testdata/blast/internal/domain/product/handler/http/handler.go`
- Create: `testdata/blast/internal/domain/shipping/core/model/shipping.go`
- Create: `testdata/blast/internal/domain/shipping/app/service.go`
- Create: `testdata/blast/internal/domain/payment/core/model/payment.go`
- Create: `testdata/blast/internal/domain/payment/app/service.go`
- Create: `testdata/blast/internal/pkg/shared.go`

The topology is designed so `internal/pkg` is imported by many packages (star hub) and `order/core/model` is imported heavily within the order domain. This creates a clear statistical outlier for `internal/pkg`.

- [ ] **Step 1: Create go.mod**

```
testdata/blast/go.mod
```
```go
module github.com/kimtaeyun/testproject-blast

go 1.23
```

- [ ] **Step 2: Create the shared hub package (internal/pkg)**

This is the star hub — many packages will import it, making it the statistical outlier.

```go
// testdata/blast/internal/pkg/shared.go
package pkg

func SharedHelper() string { return "shared" }
```

- [ ] **Step 3: Create order domain packages**

`order/core/model` — leaf, no imports:
```go
// testdata/blast/internal/domain/order/core/model/order.go
package model

type Order struct{ ID string }
```

`order/core/repo` — imports core/model:
```go
// testdata/blast/internal/domain/order/core/repo/repository.go
package repo

import "github.com/kimtaeyun/testproject-blast/internal/domain/order/core/model"

type Repository interface {
	Find(id string) (model.Order, error)
}
```

`order/core/svc` — imports core/model:
```go
// testdata/blast/internal/domain/order/core/svc/order.go
package svc

import "github.com/kimtaeyun/testproject-blast/internal/domain/order/core/model"

type Validator interface {
	Validate(o model.Order) error
}
```

`order/event` — imports core/model:
```go
// testdata/blast/internal/domain/order/event/events.go
package event

import "github.com/kimtaeyun/testproject-blast/internal/domain/order/core/model"

type OrderCreated struct{ Order model.Order }
```

`order/app` — imports core/model, core/repo, core/svc, event, and pkg:
```go
// testdata/blast/internal/domain/order/app/service.go
package app

import (
	"github.com/kimtaeyun/testproject-blast/internal/domain/order/core/model"
	"github.com/kimtaeyun/testproject-blast/internal/domain/order/core/repo"
	"github.com/kimtaeyun/testproject-blast/internal/domain/order/core/svc"
	"github.com/kimtaeyun/testproject-blast/internal/domain/order/event"
	"github.com/kimtaeyun/testproject-blast/internal/pkg"
)

type Service struct{}

func (s Service) Create() {
	_ = model.Order{}
	_ = (*repo.Repository)(nil)
	_ = (*svc.Validator)(nil)
	_ = event.OrderCreated{}
	_ = pkg.SharedHelper()
}
```

`order/infra/persistence` — imports core/repo, core/model, and pkg:
```go
// testdata/blast/internal/domain/order/infra/persistence/store.go
package persistence

import (
	"github.com/kimtaeyun/testproject-blast/internal/domain/order/core/model"
	"github.com/kimtaeyun/testproject-blast/internal/domain/order/core/repo"
	"github.com/kimtaeyun/testproject-blast/internal/pkg"
)

type Store struct{}

func (s Store) Find(id string) (model.Order, error) {
	_ = pkg.SharedHelper()
	return model.Order{ID: id}, nil
}

var _ repo.Repository = Store{}
```

`order/handler/http` — imports app and pkg:
```go
// testdata/blast/internal/domain/order/handler/http/handler.go
package http

import (
	"github.com/kimtaeyun/testproject-blast/internal/domain/order/app"
	"github.com/kimtaeyun/testproject-blast/internal/pkg"
)

type Handler struct{ Svc app.Service }

func (h Handler) Handle() { _ = pkg.SharedHelper() }
```

- [ ] **Step 4: Create user domain packages**

`user/core/model`:
```go
// testdata/blast/internal/domain/user/core/model/user.go
package model

type User struct{ ID string }
```

`user/core/repo`:
```go
// testdata/blast/internal/domain/user/core/repo/repository.go
package repo

import "github.com/kimtaeyun/testproject-blast/internal/domain/user/core/model"

type Repository interface {
	Find(id string) (model.User, error)
}
```

`user/app` — imports core/model, core/repo, pkg:
```go
// testdata/blast/internal/domain/user/app/service.go
package app

import (
	"github.com/kimtaeyun/testproject-blast/internal/domain/user/core/model"
	"github.com/kimtaeyun/testproject-blast/internal/domain/user/core/repo"
	"github.com/kimtaeyun/testproject-blast/internal/pkg"
)

type Service struct{}

func (s Service) Get() {
	_ = model.User{}
	_ = (*repo.Repository)(nil)
	_ = pkg.SharedHelper()
}
```

`user/handler/http` — imports app, pkg:
```go
// testdata/blast/internal/domain/user/handler/http/handler.go
package http

import (
	"github.com/kimtaeyun/testproject-blast/internal/domain/user/app"
	"github.com/kimtaeyun/testproject-blast/internal/pkg"
)

type Handler struct{ Svc app.Service }

func (h Handler) Handle() { _ = pkg.SharedHelper() }
```

- [ ] **Step 5: Create product domain packages**

`product/core/model`:
```go
// testdata/blast/internal/domain/product/core/model/product.go
package model

type Product struct{ ID string }
```

`product/app` — imports core/model, pkg:
```go
// testdata/blast/internal/domain/product/app/service.go
package app

import (
	"github.com/kimtaeyun/testproject-blast/internal/domain/product/core/model"
	"github.com/kimtaeyun/testproject-blast/internal/pkg"
)

type Service struct{}

func (s Service) Get() {
	_ = model.Product{}
	_ = pkg.SharedHelper()
}
```

`product/handler/http` — imports app, pkg:
```go
// testdata/blast/internal/domain/product/handler/http/handler.go
package http

import (
	"github.com/kimtaeyun/testproject-blast/internal/domain/product/app"
	"github.com/kimtaeyun/testproject-blast/internal/pkg"
)

type Handler struct{ Svc app.Service }

func (h Handler) Handle() { _ = pkg.SharedHelper() }
```

- [ ] **Step 6: Create shipping and payment domain packages**

`shipping/core/model`:
```go
// testdata/blast/internal/domain/shipping/core/model/shipping.go
package model

type Shipping struct{ ID string }
```

`shipping/app` — imports core/model, pkg:
```go
// testdata/blast/internal/domain/shipping/app/service.go
package app

import (
	"github.com/kimtaeyun/testproject-blast/internal/domain/shipping/core/model"
	"github.com/kimtaeyun/testproject-blast/internal/pkg"
)

type Service struct{}

func (s Service) Ship() {
	_ = model.Shipping{}
	_ = pkg.SharedHelper()
}
```

`payment/core/model`:
```go
// testdata/blast/internal/domain/payment/core/model/payment.go
package model

type Payment struct{ ID string }
```

`payment/app` — imports core/model, pkg:
```go
// testdata/blast/internal/domain/payment/app/service.go
package app

import (
	"github.com/kimtaeyun/testproject-blast/internal/domain/payment/core/model"
	"github.com/kimtaeyun/testproject-blast/internal/pkg"
)

type Service struct{}

func (s Service) Pay() {
	_ = model.Payment{}
	_ = pkg.SharedHelper()
}
```

- [ ] **Step 7: Verify testdata compiles**

Run: `cd testdata/blast && go build ./...`
Expected: no errors

- [ ] **Step 8: Commit**

```bash
git add testdata/blast/
git commit -m "test: add blast radius testdata project with star topology"
```

---

### Task 2: Implement graph building and metrics computation

**Files:**
- Create: `rules/blast.go`
- Create: `rules/blast_test.go`

- [ ] **Step 1: Write failing test for graph building**

```go
// rules/blast_test.go
package rules_test

import (
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/analyzer"
	"github.com/NamhaeSusan/go-arch-guard/rules"
)

func TestAnalyzeBlastRadius(t *testing.T) {
	t.Run("returns no violations for small project", func(t *testing.T) {
		pkgs := loadValid(t)
		violations := rules.AnalyzeBlastRadius(pkgs, "github.com/kimtaeyun/testproject-dc", "../testdata/valid")
		for _, v := range violations {
			if v.Severity == rules.Error {
				t.Errorf("unexpected error violation: %s", v.String())
			}
		}
	})
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd rules && go test -run TestAnalyzeBlastRadius -v`
Expected: FAIL — `AnalyzeBlastRadius` undefined

- [ ] **Step 3: Write minimal AnalyzeBlastRadius skeleton**

```go
// rules/blast.go
package rules

import (
	"fmt"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"
)

// AnalyzeBlastRadius computes coupling metrics for internal packages and emits
// Warning violations for packages whose transitive dependents count is a
// statistical outlier (IQR method). No configuration required.
func AnalyzeBlastRadius(pkgs []*packages.Package, projectModule string, projectRoot string, opts ...Option) []Violation {
	cfg := NewConfig(opts...)
	if cfg.Sev == Error {
		cfg.Sev = Warning // default to Warning for blast radius
	}
	projectModule = resolveModule(pkgs, projectModule)
	projectRoot = resolveRoot(pkgs, projectRoot)
	if warns := validateModule(pkgs, projectModule); len(warns) > 0 {
		return warns
	}

	internalPrefix := projectModule + "/"

	// 1. Collect internal packages
	internalPkgs := make(map[string]bool)
	for _, pkg := range pkgs {
		if strings.HasPrefix(pkg.PkgPath, internalPrefix) {
			if !isExcludedPackage(cfg, pkg.PkgPath, projectModule) {
				internalPkgs[pkg.PkgPath] = true
			}
		}
	}

	if len(internalPkgs) < 5 {
		return nil
	}

	// 2. Build forward adjacency (pkg -> internal imports)
	forward := make(map[string]map[string]bool)
	for _, pkg := range pkgs {
		if !internalPkgs[pkg.PkgPath] {
			continue
		}
		deps := make(map[string]bool)
		for impPath := range pkg.Imports {
			if internalPkgs[impPath] {
				deps[impPath] = true
			}
		}
		forward[pkg.PkgPath] = deps
	}

	// 3. Build reverse adjacency (pkg -> internal importers)
	reverse := make(map[string]map[string]bool)
	for pkg := range internalPkgs {
		reverse[pkg] = make(map[string]bool)
	}
	for src, deps := range forward {
		for dep := range deps {
			reverse[dep][src] = true
		}
	}

	// 4. Compute metrics per package
	type pkgMetrics struct {
		path                string
		ca                  int     // afferent coupling
		ce                  int     // efferent coupling
		instability         float64
		transitiveDependents int
	}

	metrics := make([]pkgMetrics, 0, len(internalPkgs))
	for pkg := range internalPkgs {
		ca := len(reverse[pkg])
		ce := len(forward[pkg])
		var instability float64
		if ca+ce > 0 {
			instability = float64(ce) / float64(ca+ce)
		}

		// BFS for transitive dependents
		visited := make(map[string]bool)
		queue := []string{pkg}
		visited[pkg] = true
		for len(queue) > 0 {
			cur := queue[0]
			queue = queue[1:]
			for dep := range reverse[cur] {
				if !visited[dep] {
					visited[dep] = true
					queue = append(queue, dep)
				}
			}
		}
		td := len(visited) - 1 // exclude self

		metrics = append(metrics, pkgMetrics{
			path:                pkg,
			ca:                  ca,
			ce:                  ce,
			instability:         instability,
			transitiveDependents: td,
		})
	}

	// 5. IQR outlier detection on transitive dependents
	tdValues := make([]int, len(metrics))
	for i, m := range metrics {
		tdValues[i] = m.transitiveDependents
	}
	sort.Ints(tdValues)

	q1 := percentile(tdValues, 25)
	q3 := percentile(tdValues, 75)
	iqr := q3 - q1
	if iqr == 0 {
		return nil
	}
	threshold := q3 + 1.5*iqr
	if threshold < 2 {
		threshold = 2
	}

	// 6. Emit violations
	var violations []Violation
	for _, m := range metrics {
		if float64(m.transitiveDependents) > threshold {
			relPath := projectRelativePackagePath(m.path, projectModule)
			violations = append(violations, Violation{
				File: relPath,
				Rule: "blast-radius.high-coupling",
				Message: fmt.Sprintf(
					"package %q has %d transitive dependents (Ca:%d Ce:%d Instability:%.2f, threshold:%.0f)",
					relPath, m.transitiveDependents, m.ca, m.ce, m.instability, threshold,
				),
				Fix:      "consider breaking this package into smaller units or introducing an interface boundary",
				Severity: cfg.Sev,
			})
		}
	}

	sort.Slice(violations, func(i, j int) bool {
		return violations[i].File < violations[j].File
	})

	return violations
}

func percentile(sorted []int, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	k := (p / 100) * float64(len(sorted)-1)
	f := int(k)
	c := f + 1
	if c >= len(sorted) {
		return float64(sorted[f])
	}
	d := k - float64(f)
	return float64(sorted[f])*(1-d) + float64(sorted[c])*d
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd rules && go test -run TestAnalyzeBlastRadius -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add rules/blast.go rules/blast_test.go
git commit -m "feat: add AnalyzeBlastRadius with graph metrics and IQR outlier detection"
```

---

### Task 3: Test with blast testdata (outlier detection)

**Files:**
- Modify: `rules/blast_test.go`
- Modify: `rules/testhelpers_test.go`

- [ ] **Step 1: Add blast testdata loader to testhelpers**

```go
// Add to rules/testhelpers_test.go
var (
	blastOnce sync.Once
	blastPkgs []*packages.Package
	blastErr  error
)

func loadBlast(t *testing.T) []*packages.Package {
	t.Helper()
	blastOnce.Do(func() {
		blastPkgs, blastErr = analyzer.Load("../testdata/blast", "internal/...")
	})
	if blastErr != nil {
		t.Fatal(blastErr)
	}
	return blastPkgs
}
```

- [ ] **Step 2: Write test that expects outlier detection**

```go
// Add to rules/blast_test.go
func TestAnalyzeBlastRadius_DetectsOutlier(t *testing.T) {
	pkgs := loadBlast(t)
	violations := rules.AnalyzeBlastRadius(pkgs, "github.com/kimtaeyun/testproject-blast", "../testdata/blast")

	if len(violations) == 0 {
		t.Fatal("expected at least one blast-radius violation for the hub package")
	}

	foundPkg := false
	for _, v := range violations {
		if v.Rule != "blast-radius.high-coupling" {
			t.Errorf("unexpected rule: %s", v.Rule)
		}
		if v.Severity != rules.Warning {
			t.Errorf("expected Warning severity, got %v", v.Severity)
		}
		if strings.Contains(v.File, "internal/pkg") {
			foundPkg = true
		}
		t.Logf("violation: %s", v.String())
	}
	if !foundPkg {
		t.Error("expected internal/pkg to be flagged as high-coupling outlier")
	}
}
```

Add `"strings"` to the import block at the top of the test file.

- [ ] **Step 3: Run test to verify it passes**

Run: `cd rules && go test -run TestAnalyzeBlastRadius -v`
Expected: PASS — `internal/pkg` detected as outlier (imported by 9+ packages out of ~20)

- [ ] **Step 4: Add edge case tests**

```go
// Add to rules/blast_test.go
func TestAnalyzeBlastRadius_SkipsTooFewPackages(t *testing.T) {
	// valid testdata has fewer than 5 internal packages? Check — if not, use
	// the existing test above as proof that small projects return no errors.
	pkgs := loadValid(t)
	violations := rules.AnalyzeBlastRadius(pkgs, "github.com/kimtaeyun/testproject-dc", "../testdata/valid")
	for _, v := range violations {
		if v.Severity == rules.Error {
			t.Errorf("unexpected error-severity violation: %s", v.String())
		}
	}
}

func TestAnalyzeBlastRadius_RespectsExclude(t *testing.T) {
	pkgs := loadBlast(t)
	violations := rules.AnalyzeBlastRadius(pkgs,
		"github.com/kimtaeyun/testproject-blast", "../testdata/blast",
		rules.WithExclude("internal/pkg/..."),
	)
	for _, v := range violations {
		if strings.Contains(v.File, "internal/pkg") {
			t.Error("excluded package should not appear in violations")
		}
	}
}

func TestAnalyzeBlastRadius_RespectsSeverity(t *testing.T) {
	pkgs := loadBlast(t)
	violations := rules.AnalyzeBlastRadius(pkgs,
		"github.com/kimtaeyun/testproject-blast", "../testdata/blast",
		rules.WithSeverity(rules.Error),
	)
	for _, v := range violations {
		if v.Severity != rules.Error {
			t.Errorf("expected Error severity, got %v", v.Severity)
		}
	}
}
```

- [ ] **Step 5: Run all tests**

Run: `cd rules && go test -run TestAnalyzeBlastRadius -v`
Expected: all PASS

- [ ] **Step 6: Commit**

```bash
git add rules/blast_test.go rules/testhelpers_test.go
git commit -m "test: add blast radius outlier detection and edge case tests"
```

---

### Task 4: Add integration test and update example

**Files:**
- Modify: `integration_test.go`
- Modify: `example_test.go`

- [ ] **Step 1: Add blast radius to integration test**

Add to `TestIntegration_Valid`:
```go
t.Run("blast radius", func(t *testing.T) {
	report.AssertNoViolations(t, rules.AnalyzeBlastRadius(pkgs, "github.com/kimtaeyun/testproject-dc", "testdata/valid"))
})
```

Add a new integration test for the blast testdata:
```go
func TestIntegration_BlastRadius(t *testing.T) {
	pkgs, err := analyzer.Load("testdata/blast", "internal/...")
	if err != nil {
		t.Fatal(err)
	}

	violations := rules.AnalyzeBlastRadius(pkgs, "github.com/kimtaeyun/testproject-blast", "testdata/blast")
	if len(violations) == 0 {
		t.Error("expected blast radius violations for hub package")
	}
	assertHasRule(t, violations, "blast-radius.high-coupling")
}
```

- [ ] **Step 2: Update example_test.go**

Add blast radius to the example:
```go
violations = append(violations, rules.AnalyzeBlastRadius(pkgs, "github.com/kimtaeyun/testproject-dc", "testdata/valid")...)
```

- [ ] **Step 3: Run full test suite**

Run: `go test ./... -v`
Expected: all PASS

- [ ] **Step 4: Commit**

```bash
git add integration_test.go example_test.go
git commit -m "test: add blast radius integration test and example"
```

---

### Task 5: Update documentation

**Files:**
- Modify: `README.md`
- Modify: `plugins/go-arch-guard/skills/go-arch-guard/SKILL.md`

- [ ] **Step 1: Add blast radius section to README.md**

After the Naming rules section and before Options, add:

```markdown
### Blast Radius

`rules.AnalyzeBlastRadius(pkgs, module, root, opts...)`

Purpose:

- surface internal packages with abnormally high coupling
- zero configuration — uses IQR-based statistical outlier detection
- default severity is Warning (does not fail tests unless overridden)

| Rule | Meaning |
|------|---------|
| `blast-radius.high-coupling` | package has statistically outlying transitive dependents |

Metrics computed per internal package:

| Metric | Definition |
|--------|-----------|
| Ca (Afferent Coupling) | count of internal packages that import this package |
| Ce (Efferent Coupling) | count of internal packages this package imports |
| Instability | Ce / (Ca + Ce) — 0 = stable, 1 = unstable |
| Transitive Dependents | full reverse-reachable set via BFS |

Outlier detection uses median + 1.5 × IQR on the Transitive Dependents distribution. Projects with fewer than 5 internal packages skip analysis.
```

Update the Quick Start `architecture_test.go` to include:
```go
t.Run("blast radius", func(t *testing.T) {
	report.AssertNoViolations(t, rules.AnalyzeBlastRadius(pkgs, module, root))
})
```

Update the API Reference table to add:
```
| `rules.AnalyzeBlastRadius(pkgs, module, root, opts...)` | coupling outlier detection (default Warning) |
```

Update the opening sentence to include "blast-radius":
```
Define isolation, layer-direction, structure, naming, and blast-radius rules, then fail regular tests when the project shape drifts.
```

- [ ] **Step 2: Update SKILL.md**

Add blast radius to the architecture_test.go template and add a brief description section in the appropriate location within SKILL.md.

- [ ] **Step 3: Run lint**

Run: `make lint`
Expected: 0 issues

- [ ] **Step 4: Commit**

```bash
git add README.md plugins/go-arch-guard/skills/go-arch-guard/SKILL.md
git commit -m "docs: add blast radius guard to README and SKILL"
```

---

### Task 6: Work log

**Files:**
- Create: `claude_history/2026-03-27-blast-radius-guard.md`

- [ ] **Step 1: Write work log**

```markdown
# Blast Radius Guard

## Summary
Added `AnalyzeBlastRadius()` — a new analysis function that computes coupling metrics (Ca, Ce, Instability, Transitive Dependents) for internal packages and emits Warning violations for statistical outliers using IQR-based detection. Zero configuration required.

## Files Changed
- `rules/blast.go` — graph construction, metrics computation, IQR outlier detection
- `rules/blast_test.go` — unit tests with blast testdata
- `rules/testhelpers_test.go` — added loadBlast helper
- `testdata/blast/` — dedicated test project with star topology (20 packages, pkg as hub)
- `integration_test.go` — added blast radius integration tests
- `example_test.go` — added blast radius to example
- `README.md` — API documentation
- `SKILL.md` — rule description

## Verification
- All existing tests pass (`go test ./...`)
- Blast testdata correctly identifies `internal/pkg` as high-coupling outlier
- Edge cases covered: small projects, excluded packages, severity override
- `make lint` passes
```

- [ ] **Step 2: Commit**

```bash
git add claude_history/2026-03-27-blast-radius-guard.md
git commit -m "docs: add blast radius guard work log"
```
