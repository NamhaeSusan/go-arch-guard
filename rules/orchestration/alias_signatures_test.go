package orchestration_test

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/analyzer"
	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/rules/orchestration"
)

func TestAliasSignaturesSpec(t *testing.T) {
	rule := orchestration.NewAliasSignatures(orchestration.WithSeverity(core.Error))

	spec := rule.Spec()
	if spec.ID != "orchestration.alias-signatures" {
		t.Fatalf("Spec().ID = %q, want orchestration.alias-signatures", spec.ID)
	}
	if spec.DefaultSeverity != core.Error {
		t.Fatalf("Spec().DefaultSeverity = %v, want Error", spec.DefaultSeverity)
	}
	if !slices.Contains(spec.ViolationIDs(), "orchestration.alias-signatures") {
		t.Fatalf("Spec().ViolationIDs() = %v, want orchestration.alias-signatures", spec.ViolationIDs())
	}
}

func TestAliasSignaturesDetectsDirectDomainSubpackageReturn(t *testing.T) {
	root := writeFixture(t, "example.com/direct-leak", map[string]string{
		"internal/domain/order/core/model/order.go": "package model\n\ntype Order struct{ ID string }\n",
		"internal/orchestration/checkout.go": `package orchestration

import "example.com/direct-leak/internal/domain/order/core/model"

func Place() (model.Order, error) {
	return model.Order{}, nil
}
`,
	})

	violations := orchestration.NewAliasSignatures().Check(loadContext(t, root, domainArchitecture(), "example.com/direct-leak"))

	assertHasRule(t, violations, "orchestration.alias-signatures")
	if !strings.Contains(violations[0].Message, "core/model") {
		t.Fatalf("Message should mention leaked sublayer, got %q", violations[0].Message)
	}
	if violations[0].EffectiveSeverity != core.Warning {
		t.Fatalf("EffectiveSeverity = %v, want Warning", violations[0].EffectiveSeverity)
	}
}

func TestAliasSignaturesDetectsDomainRootAliasMediatedLeak(t *testing.T) {
	root := writeFixture(t, "example.com/alias-leak", map[string]string{
		"internal/domain/order/alias.go": `package order

import "example.com/alias-leak/internal/domain/order/core/model"

type Order = model.Order
`,
		"internal/domain/order/core/model/order.go": "package model\n\ntype Order struct{ ID string }\n",
		"internal/orchestration/checkout.go": `package orchestration

import "example.com/alias-leak/internal/domain/order"

func Place() (order.Order, error) {
	return order.Order{}, nil
}
`,
	})

	violations := orchestration.NewAliasSignatures().Check(loadContext(t, root, domainArchitecture(), "example.com/alias-leak"))

	assertHasRule(t, violations, "orchestration.alias-signatures")
}

func TestAliasSignaturesDetectsParametersAndAppSublayerLeaks(t *testing.T) {
	root := writeFixture(t, "example.com/param-leak", map[string]string{
		"internal/domain/order/app/result.go":       "package app\n\ntype Result struct{ ID string }\n",
		"internal/domain/order/core/model/order.go": "package model\n\ntype Order struct{ ID string }\n",
		"internal/orchestration/checkout.go": `package orchestration

import (
	"example.com/param-leak/internal/domain/order/app"
	"example.com/param-leak/internal/domain/order/core/model"
)

func Place(order model.Order) (app.Result, error) {
	return app.Result{ID: order.ID}, nil
}
`,
	})

	violations := orchestration.NewAliasSignatures().Check(loadContext(t, root, domainArchitecture(), "example.com/param-leak"))

	assertHasRule(t, violations, "orchestration.alias-signatures")
	if len(violations) != 2 {
		t.Fatalf("got %d violations, want 2: %+v", len(violations), violations)
	}
}

func TestAliasSignaturesDetectsExportedInterfaceMethodLeaks(t *testing.T) {
	root := writeFixture(t, "example.com/interface-leak", map[string]string{
		"internal/domain/order/alias.go": `package order

import "example.com/interface-leak/internal/domain/order/core/model"

type Order = model.Order
`,
		"internal/domain/order/core/model/order.go": "package model\n\ntype Order struct{ ID string }\n",
		"internal/orchestration/checkout.go": `package orchestration

import "example.com/interface-leak/internal/domain/order"

type CheckoutPort interface {
	Place(order.Order) error
}
`,
	})

	violations := orchestration.NewAliasSignatures().Check(loadContext(t, root, domainArchitecture(), "example.com/interface-leak"))

	assertHasRule(t, violations, "orchestration.alias-signatures")
}

func TestAliasSignaturesAllowsLocalDTOsAndConstructorServiceDependencies(t *testing.T) {
	root := writeFixture(t, "example.com/allowed-orchestration", map[string]string{
		"internal/domain/order/alias.go": `package order

import "example.com/allowed-orchestration/internal/domain/order/app"

type Service = app.Service
`,
		"internal/domain/order/app/service.go": "package app\n\ntype Service struct{}\n",
		"internal/orchestration/checkout.go": `package orchestration

import "example.com/allowed-orchestration/internal/domain/order"

type Checkout struct {
	orders *order.Service
}

type Result struct {
	OrderID string
}

func NewCheckout(orders *order.Service) *Checkout {
	return &Checkout{orders: orders}
}

func (c *Checkout) Place() (Result, error) {
	return Result{OrderID: "order-1"}, nil
}
`,
	})

	violations := orchestration.NewAliasSignatures().Check(loadContext(t, root, domainArchitecture(), "example.com/allowed-orchestration"))

	assertLacksRule(t, violations, "orchestration.alias-signatures")
}

func TestAliasSignaturesCanDisallowConstructorServiceDependencies(t *testing.T) {
	root := writeFixture(t, "example.com/no-constructor-exception", map[string]string{
		"internal/domain/order/alias.go": `package order

import "example.com/no-constructor-exception/internal/domain/order/app"

type Service = app.Service
`,
		"internal/domain/order/app/service.go": "package app\n\ntype Service struct{}\n",
		"internal/orchestration/checkout.go": `package orchestration

import "example.com/no-constructor-exception/internal/domain/order"

type Checkout struct{}

func NewCheckout(orders *order.Service) *Checkout {
	return &Checkout{}
}
`,
	})

	violations := orchestration.NewAliasSignatures(
		orchestration.WithConstructorServiceAliases(false),
	).Check(loadContext(t, root, domainArchitecture(), "example.com/no-constructor-exception"))

	assertHasRule(t, violations, "orchestration.alias-signatures")
}

func TestAliasSignaturesDomainDirEmptyEmitsMetaDisabledByConfig(t *testing.T) {
	root := writeFixture(t, "example.com/flat", map[string]string{
		"internal/orchestration/checkout.go": "package orchestration\n",
	})
	arch := domainArchitecture()
	arch.Layout.DomainDir = ""

	violations := orchestration.NewAliasSignatures().Check(loadContext(t, root, arch, "example.com/flat"))

	assertHasRule(t, violations, "meta.rule-disabled-by-config")
}

func TestAliasSignaturesOrchestrationDirEmptyEmitsMetaDisabledByConfig(t *testing.T) {
	root := writeFixture(t, "example.com/no-orchestration", map[string]string{
		"internal/domain/order/core/model/order.go": "package model\n\ntype Order struct{ ID string }\n",
	})
	arch := domainArchitecture()
	arch.Layout.OrchestrationDir = ""

	violations := orchestration.NewAliasSignatures().Check(loadContext(t, root, arch, "example.com/no-orchestration"))

	assertHasRule(t, violations, "meta.rule-disabled-by-config")
}

func writeFixture(t *testing.T, module string, files map[string]string) string {
	t.Helper()

	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.26.1\n")
	for name, content := range files {
		writeFile(t, filepath.Join(root, name), content)
	}
	return root
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func loadContext(t *testing.T, root string, arch core.Architecture, module string) *core.Context {
	t.Helper()

	pkgs, err := analyzer.Load(root, "...")
	if err != nil {
		t.Fatalf("load packages: %v", err)
	}
	if len(pkgs) == 0 {
		t.Fatal("no packages loaded")
	}
	return core.NewContext(pkgs, module, root, arch, nil)
}

func domainArchitecture() core.Architecture {
	return core.Architecture{
		Layers: core.LayerModel{
			Sublayers: []string{"handler", "app", "core/model", "core/repo", "core/svc", "event", "infra"},
			Direction: map[string][]string{
				"handler":    {"app"},
				"app":        {"core/model", "core/repo", "core/svc", "event"},
				"core/model": {},
				"core/repo":  {"core/model"},
				"core/svc":   {"core/model"},
				"event":      {"core/model"},
				"infra":      {"core/repo", "core/model", "event"},
			},
		},
		Layout: core.LayoutModel{
			DomainDir:        "domain",
			OrchestrationDir: "orchestration",
			SharedDir:        "pkg",
		},
	}
}

func assertHasRule(t *testing.T, violations []core.Violation, id string) {
	t.Helper()

	for _, v := range violations {
		if v.Rule == id {
			return
		}
	}
	t.Fatalf("missing violation %q in %v", id, violations)
}

func assertLacksRule(t *testing.T, violations []core.Violation, id string) {
	t.Helper()

	for _, v := range violations {
		if v.Rule == id {
			t.Fatalf("unexpected violation %q in %v", id, violations)
		}
	}
}
