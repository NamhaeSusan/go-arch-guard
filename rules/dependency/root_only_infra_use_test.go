package dependency_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/analyzer"
	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/presets"
	"github.com/NamhaeSusan/go-arch-guard/rules/dependency"
	"golang.org/x/tools/go/packages"
)

const rootOnlyInfraUseRule = "composition.root-only-infra-use"

func TestRootOnlyInfraUseSpec(t *testing.T) {
	rule := dependency.NewRootOnlyInfraUse()
	spec := rule.Spec()

	if spec.ID != rootOnlyInfraUseRule {
		t.Fatalf("Spec().ID = %q, want %q", spec.ID, rootOnlyInfraUseRule)
	}
	if spec.DefaultSeverity != core.Warning {
		t.Fatalf("Spec().DefaultSeverity = %v, want Warning", spec.DefaultSeverity)
	}
	got := spec.ViolationIDs()
	if len(got) != 1 || got[0] != rootOnlyInfraUseRule {
		t.Fatalf("ViolationIDs() = %v, want [%s]", got, rootOnlyInfraUseRule)
	}

	strict := dependency.NewRootOnlyInfraUse(dependency.WithSeverity(core.Error))
	if strict.Spec().DefaultSeverity != core.Error {
		t.Fatalf("WithSeverity(Error) default = %v, want Error", strict.Spec().DefaultSeverity)
	}
}

func TestRootOnlyInfraUseFlagsNonCompositionInfraImports(t *testing.T) {
	module := "example.com/shop"
	root := writeFixture(t, module, map[string]string{
		"internal/domain/order/infra/memory/store.go": `package memory

type Store struct{}
`,
		"internal/domain/order/handler/http/handler.go": `package http

import _ "example.com/shop/internal/domain/order/infra/memory"
`,
		"internal/orchestration/orders.go": `package orchestration

import _ "example.com/shop/internal/domain/order/infra/memory"
`,
		"internal/domain/payment/app/service.go": `package app

import _ "example.com/shop/internal/domain/order/infra/memory"
`,
		"internal/server/http/router.go": `package http

import _ "example.com/shop/internal/domain/order/infra/memory"
`,
	})

	violations := dependency.NewRootOnlyInfraUse().Check(loadRootOnlyInfraContext(t, root, module, false))

	assertViolationCount(t, violations, rootOnlyInfraUseRule, 4)
	assertViolationAt(t, violations, "internal/domain/order/handler/http/handler.go")
	assertViolationAt(t, violations, "internal/orchestration/orders.go")
	assertViolationAt(t, violations, "internal/domain/payment/app/service.go")
	assertViolationAt(t, violations, "internal/server/http/router.go")
}

func TestRootOnlyInfraUseAllowsCompositionRootsAndSameDomainInfra(t *testing.T) {
	module := "example.com/shop"
	root := writeFixture(t, module, map[string]string{
		"internal/domain/order/infra/memory/store.go": `package memory

import _ "example.com/shop/internal/domain/order/infra/sql"

type Store struct{}
`,
		"internal/domain/order/infra/sql/store.go": `package sql

type Store struct{}
`,
		"internal/app/container.go": `package app

import _ "example.com/shop/internal/domain/order/infra/memory"
`,
		"cmd/api/main.go": `package main

import _ "example.com/shop/internal/domain/order/infra/sql"
`,
	})

	violations := dependency.NewRootOnlyInfraUse().Check(loadRootOnlyInfraContext(t, root, module, false))

	assertNoRule(t, violations, rootOnlyInfraUseRule)
}

func TestRootOnlyInfraUseAllowsSameDomainAliasBootstrapFacade(t *testing.T) {
	module := "example.com/shop"
	root := writeFixture(t, module, map[string]string{
		"internal/domain/order/alias.go": `package order

import "example.com/shop/internal/domain/order/infra/memory"

var NewRepository = memory.New
`,
		"internal/domain/order/infra/memory/store.go": `package memory

type Store struct{}

func New() *Store { return &Store{} }
`,
	})

	violations := dependency.NewRootOnlyInfraUse().Check(loadRootOnlyInfraContext(t, root, module, false))

	assertNoRule(t, violations, rootOnlyInfraUseRule)
}

func TestRootOnlyInfraUseFlagsNonAliasDomainRootInfraImports(t *testing.T) {
	module := "example.com/shop"
	root := writeFixture(t, module, map[string]string{
		"internal/domain/order/bootstrap.go": `package order

import _ "example.com/shop/internal/domain/order/infra/memory"
`,
		"internal/domain/order/infra/memory/store.go": `package memory

type Store struct{}
`,
	})

	violations := dependency.NewRootOnlyInfraUse().Check(loadRootOnlyInfraContext(t, root, module, false))

	assertViolationCount(t, violations, rootOnlyInfraUseRule, 1)
	assertViolationAt(t, violations, "internal/domain/order/bootstrap.go")
}

func TestRootOnlyInfraUseAllowsConfiguredExtraRoot(t *testing.T) {
	module := "example.com/shop"
	root := writeFixture(t, module, map[string]string{
		"internal/domain/order/infra/memory/store.go": `package memory

type Store struct{}
`,
		"internal/bootstrap/wire.go": `package bootstrap

import _ "example.com/shop/internal/domain/order/infra/memory"
`,
	})

	violations := dependency.NewRootOnlyInfraUse(
		dependency.WithCompositionRoots("internal/bootstrap/..."),
	).Check(loadRootOnlyInfraContext(t, root, module, false))

	assertNoRule(t, violations, rootOnlyInfraUseRule)
}

func TestRootOnlyInfraUseTestFilesAreConfigurable(t *testing.T) {
	module := "example.com/shop"
	root := writeFixture(t, module, map[string]string{
		"internal/domain/order/infra/memory/store.go": `package memory

type Store struct{}
`,
		"internal/domain/order/handler/http/handler.go": `package http
`,
		"internal/domain/order/handler/http/handler_test.go": `package http

import _ "example.com/shop/internal/domain/order/infra/memory"
`,
	})

	defaultViolations := dependency.NewRootOnlyInfraUse().Check(loadRootOnlyInfraContext(t, root, module, true))
	assertNoRule(t, defaultViolations, rootOnlyInfraUseRule)

	includeTests := dependency.NewRootOnlyInfraUse(dependency.WithTestFiles(true)).
		Check(loadRootOnlyInfraContext(t, root, module, true))
	assertViolationCount(t, includeTests, rootOnlyInfraUseRule, 1)
	assertViolationAt(t, includeTests, "internal/domain/order/handler/http/handler_test.go")
}

func TestRootOnlyInfraUseIsOptInForRecommendedDDD(t *testing.T) {
	for _, rule := range presets.RecommendedDDD().Rules() {
		if rule.Spec().ID == rootOnlyInfraUseRule {
			t.Fatalf("%s must stay opt-in, found in RecommendedDDD()", rootOnlyInfraUseRule)
		}
	}
}

func writeFixture(t *testing.T, module string, files map[string]string) string {
	t.Helper()

	root := t.TempDir()
	writeFixtureFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.26.1\n")
	for name, content := range files {
		writeFixtureFile(t, filepath.Join(root, name), content)
	}
	return root
}

func writeFixtureFile(t *testing.T, path, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func loadRootOnlyInfraContext(t *testing.T, root, module string, includeTests bool) *core.Context {
	t.Helper()

	var pkgs []*packages.Package
	if includeTests {
		pkgs = loadPackagesWithTests(t, root)
	} else {
		var err error
		pkgs, err = analyzer.Load(root, loadPatterns(root)...)
		if err != nil {
			t.Fatalf("load packages: %v", err)
		}
	}
	return core.NewContext(pkgs, module, root, presets.DDD(), nil)
}

func loadPackagesWithTests(t *testing.T, root string) []*packages.Package {
	t.Helper()

	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedImports | packages.NeedFiles |
			packages.NeedSyntax | packages.NeedModule |
			packages.NeedTypes | packages.NeedTypesInfo | packages.NeedDeps,
		Dir:   root,
		Tests: true,
	}
	pkgs, err := packages.Load(cfg, prefixedLoadPatterns(root)...)
	if err != nil {
		t.Fatalf("load packages with tests: %v", err)
	}
	if packages.PrintErrors(pkgs) > 0 {
		t.Fatalf("packages with tests had load errors")
	}
	return pkgs
}

func loadPatterns(root string) []string {
	patterns := []string{"internal/..."}
	if _, err := os.Stat(filepath.Join(root, "cmd")); err == nil {
		patterns = append(patterns, "cmd/...")
	}
	return patterns
}

func prefixedLoadPatterns(root string) []string {
	patterns := loadPatterns(root)
	for i, pattern := range patterns {
		patterns[i] = "./" + pattern
	}
	return patterns
}

func assertViolationCount(t *testing.T, violations []core.Violation, rule string, want int) {
	t.Helper()

	var got int
	for _, v := range violations {
		if v.Rule == rule {
			got++
		}
	}
	if got != want {
		t.Fatalf("%s violation count = %d, want %d; violations: %+v", rule, got, want, violations)
	}
}

func assertViolationAt(t *testing.T, violations []core.Violation, file string) {
	t.Helper()

	for _, v := range violations {
		if v.Rule == rootOnlyInfraUseRule && v.File == file {
			return
		}
	}
	t.Fatalf("missing %s violation at %s; got %+v", rootOnlyInfraUseRule, file, violations)
}

func assertNoRule(t *testing.T, violations []core.Violation, rule string) {
	t.Helper()

	for _, v := range violations {
		if v.Rule == rule {
			t.Fatalf("unexpected %s violation: %+v", rule, v)
		}
		if strings.HasPrefix(v.Rule, "meta.") {
			t.Fatalf("unexpected meta violation: %+v", v)
		}
	}
}
