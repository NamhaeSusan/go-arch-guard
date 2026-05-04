package structural_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/presets"
	"github.com/NamhaeSusan/go-arch-guard/rules/structural"
)

const nonEmptyAliasRule = "domain.non-empty-alias"

func TestNonEmptyAliasSpec(t *testing.T) {
	rule := structural.NewNonEmptyAlias()
	spec := rule.Spec()

	if spec.ID != nonEmptyAliasRule {
		t.Fatalf("Spec().ID = %q, want %q", spec.ID, nonEmptyAliasRule)
	}
	if spec.DefaultSeverity != core.Warning {
		t.Fatalf("Spec().DefaultSeverity = %v, want Warning", spec.DefaultSeverity)
	}
	got := spec.ViolationIDs()
	if len(got) != 1 || got[0] != nonEmptyAliasRule {
		t.Fatalf("ViolationIDs() = %v, want [%s]", got, nonEmptyAliasRule)
	}

	strict := structural.NewNonEmptyAlias(structural.WithSeverity(core.Error))
	if strict.Spec().DefaultSeverity != core.Error {
		t.Fatalf("WithSeverity(Error) default = %v, want Error", strict.Spec().DefaultSeverity)
	}
}

func TestNonEmptyAliasFlagsAppDomainWithEmptyAlias(t *testing.T) {
	root := writeFixture(t, map[string]string{
		"internal/domain/order/alias.go":       "package order\n",
		"internal/domain/order/app/service.go": "package app\n\ntype Service struct{}\n",
	})

	violations := structural.NewNonEmptyAlias().Check(ctx(root))

	assertViolationCount(t, violations, nonEmptyAliasRule, 1)
	assertViolationAt(t, violations, "internal/domain/order/alias.go")
}

func TestNonEmptyAliasAllowsExportedAliasSurface(t *testing.T) {
	root := writeFixture(t, map[string]string{
		"internal/domain/order/alias.go": `package order

import "example.com/shop/internal/domain/order/app"

type Service = app.Service

var NewService = app.NewService
`,
		"internal/domain/order/app/service.go": `package app

type Service struct{}

func NewService() *Service { return &Service{} }
`,
	})

	violations := structural.NewNonEmptyAlias().Check(ctx(root))

	assertNoRule(t, violations, nonEmptyAliasRule)
}

func TestNonEmptyAliasAllowsExportedConstsVarsAndFuncs(t *testing.T) {
	cases := map[string]string{
		"const": "const Version = 1\n",
		"var":   "var NewService = app.NewService\n",
		"func":  "func NewService() *app.Service { return app.NewService() }\n",
	}
	for name, decl := range cases {
		t.Run(name, func(t *testing.T) {
			root := writeFixture(t, map[string]string{
				"internal/domain/order/alias.go": `package order

import "example.com/shop/internal/domain/order/app"

` + decl,
				"internal/domain/order/app/service.go": `package app

type Service struct{}

func NewService() *Service { return &Service{} }
`,
			})

			violations := structural.NewNonEmptyAlias().Check(ctx(root))
			assertNoRule(t, violations, nonEmptyAliasRule)
		})
	}
}

func TestNonEmptyAliasPlaceholderDomainsAreConfigurable(t *testing.T) {
	root := writeFixture(t, map[string]string{
		"internal/domain/placeholder/alias.go":           "package placeholder\n",
		"internal/domain/placeholder/core/model/item.go": "package model\n\ntype Item struct{}\n",
	})

	defaultViolations := structural.NewNonEmptyAlias().Check(ctx(root))
	assertNoRule(t, defaultViolations, nonEmptyAliasRule)

	requireAll := structural.NewNonEmptyAlias(structural.WithRequirePlaceholderAliases(true)).
		Check(ctx(root))
	assertViolationCount(t, requireAll, nonEmptyAliasRule, 1)
	assertViolationAt(t, requireAll, "internal/domain/placeholder/alias.go")
}

func TestNonEmptyAliasIsOptInForRecommendedDDD(t *testing.T) {
	for _, rule := range presets.RecommendedDDD().Rules() {
		if rule.Spec().ID == nonEmptyAliasRule {
			t.Fatalf("%s must stay opt-in, found in RecommendedDDD()", nonEmptyAliasRule)
		}
	}
}

func ctx(root string) *core.Context {
	return core.NewContext(nil, "example.com/shop", root, presets.DDD(), nil)
}

func writeFixture(t *testing.T, files map[string]string) string {
	t.Helper()

	root := t.TempDir()
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
		if v.Rule == nonEmptyAliasRule && v.File == file {
			return
		}
	}
	t.Fatalf("missing %s violation at %s; got %+v", nonEmptyAliasRule, file, violations)
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
