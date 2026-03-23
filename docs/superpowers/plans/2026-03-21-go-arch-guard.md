# go-arch-guard Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a Go AST-based static analysis library that enforces DDD layered architecture rules, usable via `go test` in target projects.

**Architecture:** Independent Go module with three packages — `analyzer` (AST loading), `rules` (violation detection), `report` (formatting + test assertions). Rules check dependency direction, naming conventions, and directory structure. Supports gradual enforcement via severity levels and path exclusions.

**Tech Stack:** Go 1.22+, `go/packages`, `go/ast`

**Spec:** `docs/superpowers/specs/2026-03-21-go-arch-guard-design.md`

---

## File Structure

```
go-arch-guard/
├── go.mod
├── analyzer/
│   └── loader.go          # Load packages via go/packages
├── rules/
│   ├── rule.go            # Violation, Severity, Option types
│   ├── dependency.go      # Layer dependency checks
│   ├── naming.go          # Naming convention checks
│   └── structure.go       # Directory/package structure checks
├── report/
│   └── report.go          # Violation formatting + AssertNoViolations
├── testdata/               # Fake Go projects for testing
│   ├── valid/              # Project that passes all rules
│   │   └── internal/
│   │       ├── domain/user/{model.go,repo.go,service.go}
│   │       ├── app/user_service.go
│   │       ├── handler/http/user_handler.go
│   │       └── infra/postgres/user_repo.go
│   └── invalid/            # Project with intentional violations
│       └── internal/
│           ├── domain/user/{model.go,bad_import.go}
│           ├── domain/order/model.go
│           ├── app/user_service.go
│           ├── handler/http/user_handler.go
│           ├── infra/postgres/user_repo.go
│           ├── util/helpers.go
│           └── ...
├── analyzer/loader_test.go
├── rules/dependency_test.go
├── rules/naming_test.go
├── rules/structure_test.go
└── report/report_test.go
```

---

### Task 1: Project Scaffold + Core Types

**Files:**
- Create: `go.mod`
- Create: `rules/rule.go`
- Create: `rules/rule_test.go`

- [ ] **Step 1: Initialize Go module**

```bash
cd ~/go-arch-guard
go mod init github.com/kimtaeyun/go-arch-guard
```

- [ ] **Step 2: Write test for Violation.String()**

```go
// rules/rule_test.go
package rules_test

import (
	"testing"

	"github.com/kimtaeyun/go-arch-guard/rules"
)

func TestViolation_String(t *testing.T) {
	tests := []struct {
		name string
		v    rules.Violation
		want string
	}{
		{
			name: "error with line",
			v: rules.Violation{
				File:     "internal/domain/user/service.go",
				Line:     10,
				Rule:     "naming.no-stutter",
				Message:  `type "UserService" stutters with package "user"`,
				Fix:      `rename to "Service"`,
				Severity: rules.Error,
			},
			want: `[ERROR] violation: type "UserService" stutters with package "user" (file: internal/domain/user/service.go:10, rule: naming.no-stutter, fix: rename to "Service")`,
		},
		{
			name: "warning without line",
			v: rules.Violation{
				File:     "internal/util/",
				Line:     0,
				Rule:     "structure.banned-package",
				Message:  `package "util" is banned`,
				Fix:      "move to specific domain or pkg/",
				Severity: rules.Warning,
			},
			want: `[WARNING] violation: package "util" is banned (file: internal/util/, rule: structure.banned-package, fix: move to specific domain or pkg/)`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.v.String(); got != tt.want {
				t.Errorf("got:\n%s\nwant:\n%s", got, tt.want)
			}
		})
	}
}

func TestOptions(t *testing.T) {
	t.Run("default severity is Error", func(t *testing.T) {
		cfg := rules.NewConfig()
		if cfg.Sev != rules.Error {
			t.Errorf("got %v, want Error", cfg.Sev)
		}
	})
	t.Run("WithSeverity sets level", func(t *testing.T) {
		cfg := rules.NewConfig(rules.WithSeverity(rules.Warning))
		if cfg.Sev != rules.Warning {
			t.Errorf("got %v, want Warning", cfg.Sev)
		}
	})
	t.Run("WithExclude sets patterns", func(t *testing.T) {
		cfg := rules.NewConfig(rules.WithExclude("internal/legacy/..."))
		if len(cfg.ExcludePatterns) != 1 || cfg.ExcludePatterns[0] != "internal/legacy/..." {
			t.Errorf("got %v", cfg.ExcludePatterns)
		}
	})
}
```

- [ ] **Step 3: Run test to verify it fails**

Run: `go test ./rules/ -v`
Expected: FAIL — types not defined

- [ ] **Step 4: Implement rule.go**

```go
// rules/rule.go
package rules

import "fmt"

type Severity int

const (
	Error   Severity = iota
	Warning
)

type Violation struct {
	File     string
	Line     int
	Rule     string
	Message  string
	Fix      string
	Severity Severity
}

func (v Violation) String() string {
	sev := "ERROR"
	if v.Severity == Warning {
		sev = "WARNING"
	}
	fileStr := v.File
	if v.Line > 0 {
		fileStr = fmt.Sprintf("%s:%d", v.File, v.Line)
	}
	return fmt.Sprintf("[%s] violation: %s (file: %s, rule: %s, fix: %s)",
		sev, v.Message, fileStr, v.Rule, v.Fix)
}

type Option func(*Config)

type Config struct {
	Sev             Severity
	ExcludePatterns []string
}

func NewConfig(opts ...Option) Config {
	c := Config{Sev: Error}
	for _, o := range opts {
		o(&c)
	}
	return c
}

func WithSeverity(s Severity) Option {
	return func(c *Config) { c.Sev = s }
}

func WithExclude(patterns ...string) Option {
	return func(c *Config) { c.ExcludePatterns = append(c.ExcludePatterns, patterns...) }
}

// IsExcluded checks if a file path matches any exclude pattern.
// Patterns use "..." suffix as wildcard for directory trees.
func (c Config) IsExcluded(path string) bool {
	for _, p := range c.ExcludePatterns {
		if matchPattern(p, path) {
			return true
		}
	}
	return false
}

func matchPattern(pattern, path string) bool {
	if len(pattern) > 3 && pattern[len(pattern)-3:] == "..." {
		prefix := pattern[:len(pattern)-3]
		return len(path) >= len(prefix) && path[:len(prefix)] == prefix
	}
	return pattern == path
}
```

- [ ] **Step 5: Run tests**

Run: `go test ./rules/ -v`
Expected: PASS

- [ ] **Step 6: Initialize git and commit**

```bash
cd ~/go-arch-guard
git init
git add go.mod rules/
git commit -m "feat: add core types — Violation, Severity, Option"
```

---

### Task 2: Analyzer (Package Loader)

**Files:**
- Create: `analyzer/loader.go`
- Create: `analyzer/loader_test.go`
- Create: `testdata/valid/go.mod`
- Create: `testdata/valid/internal/domain/user/model.go`

- [ ] **Step 1: Create minimal testdata project**

```bash
mkdir -p ~/go-arch-guard/testdata/valid/internal/domain/user
```

```go
// testdata/valid/go.mod
module github.com/kimtaeyun/testproject

go 1.22
```

```go
// testdata/valid/internal/domain/user/model.go
package user

type User struct {
	ID   string
	Name string
}
```

- [ ] **Step 2: Write failing test for loader**

```go
// analyzer/loader_test.go
package analyzer_test

import (
	"testing"

	"github.com/kimtaeyun/go-arch-guard/analyzer"
)

func TestLoad(t *testing.T) {
	t.Run("loads valid project packages", func(t *testing.T) {
		pkgs, err := analyzer.Load("../testdata/valid", "internal/...")
		if err != nil {
			t.Fatal(err)
		}
		if len(pkgs) == 0 {
			t.Fatal("expected at least one package")
		}
		found := false
		for _, pkg := range pkgs {
			if pkg.Name == "user" {
				found = true
			}
		}
		if !found {
			t.Error("expected to find package 'user'")
		}
	})

	t.Run("returns error for nonexistent directory", func(t *testing.T) {
		_, err := analyzer.Load("/nonexistent", "internal/...")
		if err == nil {
			t.Error("expected error for nonexistent directory")
		}
	})
}
```

- [ ] **Step 3: Run test to verify it fails**

Run: `go test ./analyzer/ -v`
Expected: FAIL — package not found

- [ ] **Step 4: Implement loader.go**

```go
// analyzer/loader.go
package analyzer

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/tools/go/packages"
)

func Load(dir string, patterns ...string) ([]*packages.Package, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("resolve dir: %w", err)
	}
	if _, err := os.Stat(absDir); err != nil {
		return nil, fmt.Errorf("project root not found: %w", err)
	}

	prefixed := make([]string, len(patterns))
	for i, p := range patterns {
		prefixed[i] = "./" + p
	}

	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedImports | packages.NeedFiles |
			packages.NeedSyntax | packages.NeedTypes,
		Dir: absDir,
	}
	pkgs, err := packages.Load(cfg, prefixed...)
	if err != nil {
		return nil, fmt.Errorf("load packages: %w", err)
	}

	// Filter out packages with errors (best-effort)
	var result []*packages.Package
	for _, pkg := range pkgs {
		if len(pkg.Errors) == 0 {
			result = append(result, pkg)
		}
	}
	return result, nil
}
```

- [ ] **Step 5: Add golang.org/x/tools dependency**

```bash
cd ~/go-arch-guard
go get golang.org/x/tools/go/packages
go mod tidy
```

- [ ] **Step 6: Run tests**

Run: `go test ./analyzer/ -v`
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add analyzer/ testdata/ go.mod go.sum
git commit -m "feat: add analyzer package — Load via go/packages"
```

---

### Task 3: Structure Rules

**Files:**
- Create: `rules/structure.go`
- Create: `rules/structure_test.go`
- Create: testdata directories for invalid cases

Structure rules are filesystem-based (no AST), so they're the simplest to implement first.

- [ ] **Step 1: Create testdata for invalid structure**

```bash
mkdir -p ~/go-arch-guard/testdata/invalid/internal/domain/user
mkdir -p ~/go-arch-guard/testdata/invalid/internal/domain/order
mkdir -p ~/go-arch-guard/testdata/invalid/internal/util
mkdir -p ~/go-arch-guard/testdata/invalid/internal/app
```

```go
// testdata/invalid/internal/domain/user/model.go
package user
type User struct{ ID string }

// testdata/invalid/internal/domain/user/dto.go  (violation: DTO in domain)
package user
type UserDTO struct{ ID string }

// testdata/invalid/internal/domain/order/service.go  (violation: missing model.go)
package order
func Process() {}

// testdata/invalid/internal/util/helpers.go  (violation: banned package name)
package util
func Help() string { return "help" }

// testdata/invalid/internal/app/dto.go  (OK: DTO in app is allowed)
package app
type Input struct{ Name string }

// testdata/invalid/go.mod
module github.com/kimtaeyun/testproject-invalid
go 1.22
```

- [ ] **Step 2: Write failing tests**

```go
// rules/structure_test.go
package rules_test

import (
	"testing"

	"github.com/kimtaeyun/go-arch-guard/rules"
)

func TestCheckStructure(t *testing.T) {
	t.Run("valid project has no violations", func(t *testing.T) {
		violations := rules.CheckStructure("../testdata/valid")
		if len(violations) > 0 {
			t.Errorf("expected no violations, got %d: %v", len(violations), violations)
		}
	})

	t.Run("detects banned package names", func(t *testing.T) {
		violations := rules.CheckStructure("../testdata/invalid")
		found := findViolation(violations, "structure.banned-package")
		if found == nil {
			t.Error("expected banned-package violation for 'util'")
		}
	})

	t.Run("detects missing model.go in domain", func(t *testing.T) {
		violations := rules.CheckStructure("../testdata/invalid")
		found := findViolation(violations, "structure.domain-model-required")
		if found == nil {
			t.Error("expected domain-model-required violation for 'order'")
		}
	})

	t.Run("detects DTO in domain", func(t *testing.T) {
		violations := rules.CheckStructure("../testdata/invalid")
		found := findViolation(violations, "structure.dto-placement")
		if found == nil {
			t.Error("expected dto-placement violation in domain/user/")
		}
	})

	t.Run("exclude skips matching paths", func(t *testing.T) {
		violations := rules.CheckStructure("../testdata/invalid",
			rules.WithExclude("internal/util/..."))
		for _, v := range violations {
			if v.Rule == "structure.banned-package" {
				t.Error("expected util violation to be excluded")
			}
		}
	})

	t.Run("warning severity sets violation level", func(t *testing.T) {
		violations := rules.CheckStructure("../testdata/invalid",
			rules.WithSeverity(rules.Warning))
		for _, v := range violations {
			if v.Severity != rules.Warning {
				t.Errorf("expected Warning, got %v", v.Severity)
			}
		}
	})
}

func findViolation(violations []rules.Violation, ruleID string) *rules.Violation {
	for _, v := range violations {
		if v.Rule == ruleID {
			return &v
		}
	}
	return nil
}
```

- [ ] **Step 3: Run test to verify it fails**

Run: `go test ./rules/ -run TestCheckStructure -v`
Expected: FAIL — CheckStructure not defined

- [ ] **Step 4: Implement structure.go**

```go
// rules/structure.go
package rules

import (
	"os"
	"path/filepath"
	"strings"
)

var bannedPackageNames = []string{"util", "common", "misc", "helper", "shared"}

func CheckStructure(projectRoot string, opts ...Option) []Violation {
	cfg := NewConfig(opts...)
	var violations []Violation

	internalDir := filepath.Join(projectRoot, "internal")
	if _, err := os.Stat(internalDir); err != nil {
		return nil
	}

	// Check banned package names
	violations = append(violations, checkBannedPackages(internalDir, cfg)...)

	// Check domain model.go requirement
	domainDir := filepath.Join(internalDir, "domain")
	violations = append(violations, checkDomainModelRequired(domainDir, cfg)...)

	// Check DTO placement
	violations = append(violations, checkDTOPlacement(internalDir, cfg)...)

	return violations
}

func checkBannedPackages(internalDir string, cfg Config) []Violation {
	var violations []Violation
	entries, err := os.ReadDir(internalDir)
	if err != nil {
		return nil
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		relPath := filepath.Join("internal", e.Name())
		if cfg.IsExcluded(relPath + "/") {
			continue
		}
		for _, banned := range bannedPackageNames {
			if e.Name() == banned {
				violations = append(violations, Violation{
					File:     relPath + "/",
					Rule:     "structure.banned-package",
					Message:  `package "` + e.Name() + `" is banned`,
					Fix:      "move to specific domain or pkg/",
					Severity: cfg.Sev,
				})
			}
		}
	}
	return violations
}

func checkDomainModelRequired(domainDir string, cfg Config) []Violation {
	var violations []Violation
	entries, err := os.ReadDir(domainDir)
	if err != nil {
		return nil
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		relPath := filepath.Join("internal", "domain", e.Name())
		if cfg.IsExcluded(relPath + "/") {
			continue
		}
		modelPath := filepath.Join(domainDir, e.Name(), "model.go")
		if _, err := os.Stat(modelPath); err != nil {
			violations = append(violations, Violation{
				File:     relPath + "/",
				Rule:     "structure.domain-model-required",
				Message:  `domain "` + e.Name() + `" missing required file "model.go"`,
				Fix:      "create model.go with domain entities",
				Severity: cfg.Sev,
			})
		}
	}
	return violations
}

func checkDTOPlacement(internalDir string, cfg Config) []Violation {
	var violations []Violation
	// Walk domain/ and infra/ looking for dto files
	for _, forbidden := range []string{"domain", "infra"} {
		dir := filepath.Join(internalDir, forbidden)
		if _, err := os.Stat(dir); err != nil {
			continue
		}
		filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}
			name := info.Name()
			if !strings.HasSuffix(name, ".go") {
				return nil
			}
			if name == "dto.go" || (strings.HasSuffix(name, "_dto.go") && !strings.HasSuffix(name, "_test.go")) {
				rel, _ := filepath.Rel(filepath.Dir(internalDir), path)
				if cfg.IsExcluded(rel) {
					return nil
				}
				violations = append(violations, Violation{
					File:     rel,
					Rule:     "structure.dto-placement",
					Message:  `"` + name + `" found in ` + forbidden + "/",
					Fix:      "DTOs belong in handler/ or app/",
					Severity: cfg.Sev,
				})
			}
			return nil
		})
	}
	return violations
}
```

- [ ] **Step 5: Run tests**

Run: `go test ./rules/ -run TestCheckStructure -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add rules/structure.go rules/structure_test.go testdata/invalid/
git commit -m "feat: add structure rules — banned packages, domain model, DTO placement"
```

---

### Task 4: Dependency Rules

**Files:**
- Create: `rules/dependency.go`
- Create: `rules/dependency_test.go`
- Extend: `testdata/invalid/` with cross-layer import violations

- [ ] **Step 1: Create testdata with dependency violations**

Files needed in `testdata/invalid/internal/`:

```go
// testdata/invalid/internal/handler/http/user_handler.go
// violation: handler imports infra
package http

import _ "github.com/kimtaeyun/testproject-invalid/internal/infra/postgres"

func Handle() {}

// testdata/invalid/internal/domain/user/bad_import.go
// violation: domain imports app
package user

import _ "github.com/kimtaeyun/testproject-invalid/internal/app"

func Bad() {}

// testdata/invalid/internal/domain/user/cross_import.go
// violation: domain/user imports domain/order (domain isolation)
package user

import _ "github.com/kimtaeyun/testproject-invalid/internal/domain/order"

func Cross() {}

// testdata/invalid/internal/infra/postgres/user_repo.go
// OK: infra imports domain
package postgres

import _ "github.com/kimtaeyun/testproject-invalid/internal/domain/user"

func Repo() {}
```

Also create handler/http dir and infra/postgres dir in `testdata/valid/` for a clean project:

```go
// testdata/valid/internal/app/user_service.go
package app

import _ "github.com/kimtaeyun/testproject/internal/domain/user"

func Serve() {}

// testdata/valid/internal/handler/http/user_handler.go
package http

import _ "github.com/kimtaeyun/testproject/internal/app"

func Handle() {}

// testdata/valid/internal/infra/postgres/user_repo.go
package postgres

import _ "github.com/kimtaeyun/testproject/internal/domain/user"

func Repo() {}
```

- [ ] **Step 2: Write failing tests**

```go
// rules/dependency_test.go
package rules_test

import (
	"testing"

	"github.com/kimtaeyun/go-arch-guard/analyzer"
	"github.com/kimtaeyun/go-arch-guard/rules"
)

func TestCheckDependency(t *testing.T) {
	t.Run("valid project has no violations", func(t *testing.T) {
		pkgs, err := analyzer.Load("../testdata/valid", "internal/...")
		if err != nil {
			t.Fatal(err)
		}
		violations := rules.CheckDependency(pkgs, "github.com/kimtaeyun/testproject", "../testdata/valid")
		if len(violations) > 0 {
			t.Errorf("expected no violations, got %d: %v", len(violations), violations)
		}
	})

	t.Run("detects handler importing infra", func(t *testing.T) {
		pkgs, err := analyzer.Load("../testdata/invalid", "internal/...")
		if err != nil {
			t.Fatal(err)
		}
		violations := rules.CheckDependency(pkgs, "github.com/kimtaeyun/testproject-invalid", "../testdata/invalid")
		found := findViolation(violations, "dependency.layer-direction")
		if found == nil {
			t.Error("expected layer-direction violation")
		}
	})

	t.Run("detects domain importing app", func(t *testing.T) {
		pkgs, err := analyzer.Load("../testdata/invalid", "internal/...")
		if err != nil {
			t.Fatal(err)
		}
		violations := rules.CheckDependency(pkgs, "github.com/kimtaeyun/testproject-invalid", "../testdata/invalid")
		found := findViolation(violations, "dependency.domain-purity")
		if found == nil {
			t.Error("expected domain-purity violation")
		}
	})

	t.Run("detects cross-domain import", func(t *testing.T) {
		pkgs, err := analyzer.Load("../testdata/invalid", "internal/...")
		if err != nil {
			t.Fatal(err)
		}
		violations := rules.CheckDependency(pkgs, "github.com/kimtaeyun/testproject-invalid", "../testdata/invalid")
		found := findViolation(violations, "dependency.domain-isolation")
		if found == nil {
			t.Error("expected domain-isolation violation for user importing order")
		}
	})
}
```

- [ ] **Step 3: Run test to verify it fails**

Run: `go test ./rules/ -run TestCheckDependency -v`
Expected: FAIL — CheckDependency not defined

- [ ] **Step 4: Implement dependency.go**

```go
// rules/dependency.go
package rules

import (
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
)

// Layer represents a DDD architecture layer.
type Layer int

const (
	LayerUnknown Layer = iota
	LayerDomain
	LayerApp
	LayerHandler
	LayerInfra
)

// allowedImports defines which layers each layer can import.
var allowedImports = map[Layer][]Layer{
	LayerHandler: {LayerApp, LayerDomain},
	LayerApp:     {LayerDomain},
	LayerInfra:   {LayerDomain},
	LayerDomain:  {}, // domain imports nothing within internal/
}

func CheckDependency(pkgs []*packages.Package, projectModule string, projectRoot string, opts ...Option) []Violation {
	cfg := NewConfig(opts...)
	var violations []Violation
	internalPrefix := projectModule + "/internal/"

	for _, pkg := range pkgs {
		pkgPath := pkg.PkgPath
		if !strings.HasPrefix(pkgPath, internalPrefix) {
			continue
		}

		relPath := strings.TrimPrefix(pkgPath, projectModule+"/")
		if cfg.IsExcluded(relPath + "/") {
			continue
		}

		srcLayer := classifyLayer(pkgPath, internalPrefix)
		if srcLayer == LayerUnknown {
			continue
		}
		srcDomain := extractDomain(pkgPath, internalPrefix)

		for importPath := range pkg.Imports {
			if !strings.HasPrefix(importPath, internalPrefix) {
				continue
			}

			dstLayer := classifyLayer(importPath, internalPrefix)
			if dstLayer == LayerUnknown {
				continue
			}

			// Check domain isolation
			if srcLayer == LayerDomain && dstLayer == LayerDomain {
				dstDomain := extractDomain(importPath, internalPrefix)
				if srcDomain != dstDomain {
					violations = append(violations, Violation{
						File:     findImportFile(pkg, importPath, projectRoot),
						Line:     findImportLine(pkg, importPath),
						Rule:     "dependency.domain-isolation",
						Message:  `domain "` + srcDomain + `" imports domain "` + dstDomain + `" directly`,
						Fix:      "use app layer to coordinate between domains",
						Severity: cfg.Sev,
					})
				}
				continue
			}

			// Check domain purity
			if srcLayer == LayerDomain && dstLayer != LayerDomain {
				violations = append(violations, Violation{
					File:     findImportFile(pkg, importPath, projectRoot),
					Line:     findImportLine(pkg, importPath),
					Rule:     "dependency.domain-purity",
					Message:  `"` + relPath + `" imports "` + strings.TrimPrefix(importPath, projectModule+"/") + `"`,
					Fix:      "domain must not depend on any other layer",
					Severity: cfg.Sev,
				})
				continue
			}

			// Check layer direction
			if !isAllowed(srcLayer, dstLayer) {
				violations = append(violations, Violation{
					File:     findImportFile(pkg, importPath, projectRoot),
					Line:     findImportLine(pkg, importPath),
					Rule:     "dependency.layer-direction",
					Message:  `"` + layerName(srcLayer) + `" imports "` + layerName(dstLayer) + `" directly`,
					Fix:      layerFix(srcLayer),
					Severity: cfg.Sev,
				})
			}
		}
	}
	return violations
}

func classifyLayer(pkgPath, internalPrefix string) Layer {
	rel := strings.TrimPrefix(pkgPath, internalPrefix)
	parts := strings.SplitN(rel, "/", 2)
	switch parts[0] {
	case "domain":
		return LayerDomain
	case "app":
		return LayerApp
	case "handler":
		return LayerHandler
	case "infra":
		return LayerInfra
	default:
		return LayerUnknown
	}
}

func extractDomain(pkgPath, internalPrefix string) string {
	rel := strings.TrimPrefix(pkgPath, internalPrefix)
	// rel = "domain/user/..." → domain name = "user"
	parts := strings.SplitN(rel, "/", 3)
	if len(parts) >= 2 {
		return parts[1]
	}
	return ""
}

func isAllowed(src, dst Layer) bool {
	for _, allowed := range allowedImports[src] {
		if dst == allowed {
			return true
		}
	}
	return false
}

func layerName(l Layer) string {
	switch l {
	case LayerDomain:
		return "domain"
	case LayerApp:
		return "app"
	case LayerHandler:
		return "handler"
	case LayerInfra:
		return "infra"
	default:
		return "unknown"
	}
}

func layerFix(src Layer) string {
	switch src {
	case LayerHandler:
		return "handler must only depend on app or domain"
	case LayerApp:
		return "app must only depend on domain"
	case LayerInfra:
		return "infra must only depend on domain"
	default:
		return "check layer dependency direction"
	}
}

func findImportFile(pkg *packages.Package, importPath, projectRoot string) string {
	absRoot, _ := filepath.Abs(projectRoot)
	fset := pkg.Fset
	for _, file := range pkg.Syntax {
		for _, imp := range file.Imports {
			path := strings.Trim(imp.Path.Value, `"`)
			if path == importPath {
				pos := fset.Position(imp.Pos())
				rel, err := filepath.Rel(absRoot, pos.Filename)
				if err != nil {
					return pos.Filename
				}
				return rel
			}
		}
	}
	if len(pkg.GoFiles) > 0 {
		rel, err := filepath.Rel(absRoot, pkg.GoFiles[0])
		if err != nil {
			return pkg.GoFiles[0]
		}
		return rel
	}
	return pkg.PkgPath
}

func findImportLine(pkg *packages.Package, importPath string) int {
	fset := pkg.Fset
	for _, file := range pkg.Syntax {
		for _, imp := range file.Imports {
			path := strings.Trim(imp.Path.Value, `"`)
			if path == importPath {
				return fset.Position(imp.Pos()).Line
			}
		}
	}
	return 0
}
```

- [ ] **Step 5: Run tests**

Run: `go test ./rules/ -run TestCheckDependency -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add rules/dependency.go rules/dependency_test.go testdata/
git commit -m "feat: add dependency rules — layer direction, domain purity, domain isolation"
```

---

### Task 5: Naming Rules

**Files:**
- Create: `rules/naming.go`
- Create: `rules/naming_test.go`
- Extend: testdata with naming violations

- [ ] **Step 1: Add testdata with naming violations**

```go
// testdata/invalid/internal/domain/user/stutter.go — stutter violations (separate from model.go)
package user

type UserService struct{}  // violation: stutters
type Service struct{}      // OK
type PowerUser struct{}    // OK: "User" is not prefix

// testdata/invalid/internal/infra/postgres/impl.go — Impl suffix violation
package postgres

import _ "github.com/kimtaeyun/testproject-invalid/internal/domain/user"

type RepoImpl struct{}  // violation: Impl suffix

func Repo() {}
```

Create a file with bad filename:

```bash
# testdata/invalid/internal/app/userService.go  (violation: not snake_case)
```

```go
// testdata/invalid/internal/app/userService.go
package app

func UserServe() {}
```

- [ ] **Step 2: Write failing tests**

```go
// rules/naming_test.go
package rules_test

import (
	"testing"

	"github.com/kimtaeyun/go-arch-guard/analyzer"
	"github.com/kimtaeyun/go-arch-guard/rules"
)

func TestCheckNaming(t *testing.T) {
	t.Run("valid project has no violations", func(t *testing.T) {
		pkgs, err := analyzer.Load("../testdata/valid", "internal/...")
		if err != nil {
			t.Fatal(err)
		}
		violations := rules.CheckNaming(pkgs)
		if len(violations) > 0 {
			t.Errorf("expected no violations, got %d: %v", len(violations), violations)
		}
	})

	t.Run("detects package stutter", func(t *testing.T) {
		pkgs, err := analyzer.Load("../testdata/invalid", "internal/...")
		if err != nil {
			t.Fatal(err)
		}
		violations := rules.CheckNaming(pkgs)
		found := findViolation(violations, "naming.no-stutter")
		if found == nil {
			t.Error("expected no-stutter violation for UserService in package user")
		}
	})

	t.Run("detects Impl suffix", func(t *testing.T) {
		pkgs, err := analyzer.Load("../testdata/invalid", "internal/...")
		if err != nil {
			t.Fatal(err)
		}
		violations := rules.CheckNaming(pkgs)
		found := findViolation(violations, "naming.no-impl-suffix")
		if found == nil {
			t.Error("expected no-impl-suffix violation for RepoImpl")
		}
	})

	t.Run("detects non-snake-case filenames", func(t *testing.T) {
		pkgs, err := analyzer.Load("../testdata/invalid", "internal/...")
		if err != nil {
			t.Fatal(err)
		}
		violations := rules.CheckNaming(pkgs)
		found := findViolation(violations, "naming.snake-case-file")
		if found == nil {
			t.Error("expected snake-case-file violation for userService.go")
		}
	})
}
```

- [ ] **Step 3: Run test to verify it fails**

Run: `go test ./rules/ -run TestCheckNaming -v`
Expected: FAIL — CheckNaming not defined

- [ ] **Step 4: Implement naming.go**

```go
// rules/naming.go
package rules

import (
	"go/ast"
	"path/filepath"
	"strings"
	"unicode"

	"golang.org/x/tools/go/packages"
)

func CheckNaming(pkgs []*packages.Package, opts ...Option) []Violation {
	cfg := NewConfig(opts...)
	var violations []Violation

	for _, pkg := range pkgs {
		violations = append(violations, checkStutter(pkg, cfg)...)
		violations = append(violations, checkImplSuffix(pkg, cfg)...)
		violations = append(violations, checkSnakeCaseFiles(pkg, cfg)...)
	}
	return violations
}

func checkStutter(pkg *packages.Package, cfg Config) []Violation {
	var violations []Violation
	pkgName := pkg.Name

	for _, file := range pkg.Syntax {
		filePath := pkg.Fset.Position(file.Pos()).Filename
		if cfg.IsExcluded(filePath) {
			continue
		}
		for _, decl := range file.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			for _, spec := range gd.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok || !ts.Name.IsExported() {
					continue
				}
				name := ts.Name.Name
				if stutters(pkgName, name) {
					suggested := strings.TrimPrefix(strings.ToLower(name), strings.ToLower(pkgName))
					// Capitalize first letter
					if len(suggested) > 0 {
						suggested = strings.ToUpper(suggested[:1]) + suggested[1:]
					}
					pos := pkg.Fset.Position(ts.Name.Pos())
					violations = append(violations, Violation{
						File:     pos.Filename,
						Line:     pos.Line,
						Rule:     "naming.no-stutter",
						Message:  `type "` + name + `" stutters with package "` + pkgName + `"`,
						Fix:      `rename to "` + suggested + `"`,
						Severity: cfg.Sev,
					})
				}
			}
		}
	}
	return violations
}

// stutters checks if typeName starts with pkgName (case-insensitive).
func stutters(pkgName, typeName string) bool {
	if len(typeName) <= len(pkgName) {
		return false
	}
	prefix := strings.ToLower(typeName[:len(pkgName)])
	if prefix != strings.ToLower(pkgName) {
		return false
	}
	// Ensure the character after prefix is uppercase (word boundary)
	next := rune(typeName[len(pkgName)])
	return unicode.IsUpper(next)
}

func checkImplSuffix(pkg *packages.Package, cfg Config) []Violation {
	var violations []Violation
	for _, file := range pkg.Syntax {
		filePath := pkg.Fset.Position(file.Pos()).Filename
		if cfg.IsExcluded(filePath) {
			continue
		}
		for _, decl := range file.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			for _, spec := range gd.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok || !ts.Name.IsExported() {
					continue
				}
				if strings.HasSuffix(ts.Name.Name, "Impl") {
					pos := pkg.Fset.Position(ts.Name.Pos())
					violations = append(violations, Violation{
						File:     pos.Filename,
						Line:     pos.Line,
						Rule:     "naming.no-impl-suffix",
						Message:  `type "` + ts.Name.Name + `" uses banned suffix "Impl"`,
						Fix:      "rename without Impl suffix",
						Severity: cfg.Sev,
					})
				}
			}
		}
	}
	return violations
}

func checkSnakeCaseFiles(pkg *packages.Package, cfg Config) []Violation {
	var violations []Violation
	seen := make(map[string]bool)

	for _, f := range pkg.GoFiles {
		if seen[f] {
			continue
		}
		seen[f] = true

		if cfg.IsExcluded(f) {
			continue
		}
		base := filepath.Base(f)
		if !isSnakeCase(base) {
			violations = append(violations, Violation{
				File:     f,
				Rule:     "naming.snake-case-file",
				Message:  `filename "` + base + `" must be snake_case`,
				Fix:      `rename to "` + toSnakeCase(base) + `"`,
				Severity: cfg.Sev,
			})
		}
	}
	return violations
}

// isSnakeCase checks if a Go filename is snake_case.
// Allows: lowercase, digits, underscores, _test.go suffix, .go extension.
func isSnakeCase(filename string) bool {
	name := strings.TrimSuffix(filename, ".go")
	name = strings.TrimSuffix(name, "_test")
	for _, r := range name {
		if !unicode.IsLower(r) && !unicode.IsDigit(r) && r != '_' {
			return false
		}
	}
	return len(name) > 0
}

func toSnakeCase(filename string) string {
	ext := filepath.Ext(filename)
	name := strings.TrimSuffix(filename, ext)
	var result []rune
	for i, r := range name {
		if unicode.IsUpper(r) {
			if i > 0 {
				result = append(result, '_')
			}
			result = append(result, unicode.ToLower(r))
		} else {
			result = append(result, r)
		}
	}
	return string(result) + ext
}
```

- [ ] **Step 5: Run tests**

Run: `go test ./rules/ -run TestCheckNaming -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add rules/naming.go rules/naming_test.go testdata/
git commit -m "feat: add naming rules — stutter, Impl suffix, snake_case filenames"
```

---

### Task 6: Report Package

**Files:**
- Create: `report/report.go`
- Create: `report/report_test.go`

- [ ] **Step 1: Write failing test**

```go
// report/report_test.go
package report_test

import (
	"fmt"
	"testing"

	"github.com/kimtaeyun/go-arch-guard/report"
	"github.com/kimtaeyun/go-arch-guard/rules"
)

type fakeTB struct {
	testing.TB
	errors []string
	failed bool
}

func (f *fakeTB) Errorf(format string, args ...interface{}) {
	f.errors = append(f.errors, fmt.Sprintf(format, args...))
	f.failed = true
}

func (f *fakeTB) Helper() {}

func TestAssertNoViolations(t *testing.T) {
	t.Run("no violations passes", func(t *testing.T) {
		tb := &fakeTB{}
		report.AssertNoViolations(tb, nil)
		if tb.failed {
			t.Error("expected test to pass with no violations")
		}
	})

	t.Run("error violations fails test", func(t *testing.T) {
		tb := &fakeTB{}
		report.AssertNoViolations(tb, []rules.Violation{
			{Rule: "test.rule", Message: "bad", Severity: rules.Error},
		})
		if !tb.failed {
			t.Error("expected test to fail with error violations")
		}
	})

	t.Run("warning-only violations passes test", func(t *testing.T) {
		tb := &fakeTB{}
		report.AssertNoViolations(tb, []rules.Violation{
			{Rule: "test.rule", Message: "warn", Severity: rules.Warning},
		})
		if tb.failed {
			t.Error("expected test to pass with warning-only violations")
		}
	})
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./report/ -v`
Expected: FAIL — package not found

- [ ] **Step 3: Implement report.go**

```go
// report/report.go
package report

import (
	"fmt"
	"testing"

	"github.com/kimtaeyun/go-arch-guard/rules"
)

// AssertNoViolations prints all violations.
// Fails the test only if any ERROR-level violations exist.
func AssertNoViolations(t testing.TB, violations []rules.Violation) {
	t.Helper()
	hasErrors := false
	for _, v := range violations {
		fmt.Println(v.String())
		if v.Severity == rules.Error {
			hasErrors = true
		}
	}
	if hasErrors {
		t.Errorf("found %d architecture violation(s)", countErrors(violations))
	}
}

func countErrors(violations []rules.Violation) int {
	n := 0
	for _, v := range violations {
		if v.Severity == rules.Error {
			n++
		}
	}
	return n
}
```

- [ ] **Step 4: Run tests**

Run: `go test ./report/ -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add report/
git commit -m "feat: add report package — AssertNoViolations with severity support"
```

---

### Task 7: Integration Test + Polish

**Files:**
- Create: `integration_test.go` (root level)
- Verify all tests pass together

- [ ] **Step 1: Write integration test**

```go
// integration_test.go
package goarchguard_test

import (
	"testing"

	"github.com/kimtaeyun/go-arch-guard/analyzer"
	"github.com/kimtaeyun/go-arch-guard/report"
	"github.com/kimtaeyun/go-arch-guard/rules"
)

func TestIntegration_ValidProject(t *testing.T) {
	pkgs, err := analyzer.Load("testdata/valid", "internal/...")
	if err != nil {
		t.Fatal(err)
	}

	t.Run("dependency", func(t *testing.T) {
		report.AssertNoViolations(t, rules.CheckDependency(pkgs, "github.com/kimtaeyun/testproject", "testdata/valid"))
	})
	t.Run("naming", func(t *testing.T) {
		report.AssertNoViolations(t, rules.CheckNaming(pkgs))
	})
	t.Run("structure", func(t *testing.T) {
		report.AssertNoViolations(t, rules.CheckStructure("testdata/valid"))
	})
}

func TestIntegration_InvalidProject(t *testing.T) {
	pkgs, err := analyzer.Load("testdata/invalid", "internal/...")
	if err != nil {
		t.Fatal(err)
	}

	t.Run("dependency violations found", func(t *testing.T) {
		violations := rules.CheckDependency(pkgs, "github.com/kimtaeyun/testproject-invalid", "testdata/invalid")
		if len(violations) == 0 {
			t.Error("expected dependency violations")
		}
	})
	t.Run("naming violations found", func(t *testing.T) {
		violations := rules.CheckNaming(pkgs)
		if len(violations) == 0 {
			t.Error("expected naming violations")
		}
	})
	t.Run("structure violations found", func(t *testing.T) {
		violations := rules.CheckStructure("testdata/invalid")
		if len(violations) == 0 {
			t.Error("expected structure violations")
		}
	})
}

func TestIntegration_WarningMode(t *testing.T) {
	pkgs, err := analyzer.Load("testdata/invalid", "internal/...")
	if err != nil {
		t.Fatal(err)
	}

	// Warning mode: violations exist but test passes
	violations := rules.CheckDependency(pkgs, "github.com/kimtaeyun/testproject-invalid", "testdata/invalid",
		rules.WithSeverity(rules.Warning))
	if len(violations) == 0 {
		t.Error("expected violations even in warning mode")
	}
	for _, v := range violations {
		if v.Severity != rules.Warning {
			t.Errorf("expected Warning severity, got %v", v.Severity)
		}
	}
	// AssertNoViolations should pass with warnings only
	report.AssertNoViolations(t, violations)
}
```

- [ ] **Step 2: Run all tests**

Run: `go test ./... -v`
Expected: ALL PASS

- [ ] **Step 3: Commit**

```bash
git add integration_test.go
git commit -m "test: add integration tests for valid/invalid projects + warning mode"
```

---

### Task 8: Final Cleanup

- [ ] **Step 1: Create .gitignore**

```
# .gitignore
/bin/
*.exe
*.test
*.out
```

- [ ] **Step 2: Run full test suite**

```bash
go test ./... -v -count=1
```
Expected: ALL PASS

- [ ] **Step 3: Run go vet**

```bash
go vet ./...
```
Expected: No issues

- [ ] **Step 4: Final commit**

```bash
git add .gitignore
git commit -m "chore: add .gitignore"
```
