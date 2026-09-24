package interfaces_test

import (
	"github.com/NamhaeSusan/go-arch-guard/analyzer"
	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/presets"
	"github.com/NamhaeSusan/go-arch-guard/rules/interfaces"
	"os"
	"path/filepath"
	"testing"
)

func auditContext(t *testing.T, files map[string]string, arch core.Architecture, exclude ...string) *core.Context {
	t.Helper()
	root := t.TempDir()
	files["go.mod"] = "module example.com/audit\n\ngo 1.26.1\n"
	for path, source := range files {
		full := filepath.Join(root, path)
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(source), 0644); err != nil {
			t.Fatal(err)
		}
	}
	pkgs, err := analyzer.Load(root, "./...")
	if err != nil {
		t.Fatal(err)
	}
	for _, pkg := range pkgs {
		if pkg.IllTyped || len(pkg.Errors) > 0 {
			t.Fatalf("invalid fixture: %v", pkg.Errors)
		}
	}
	return core.NewContext(pkgs, "example.com/audit", root, arch, exclude)
}

func TestAuditInterfaceExclusions(t *testing.T) {
	for _, exclude := range []string{"internal/domain/order/infra/...", "internal/domain/order/infra/a.go"} {
		ctx := auditContext(t, map[string]string{"internal/domain/order/infra/a.go": "package infra;type S interface{A();B();C()};func NewS() S{return nil}"}, presets.DDD(), exclude)
		for _, rule := range []core.Rule{interfaces.NewPattern(), interfaces.NewTooManyMethods(interfaces.WithMaxMethods(2))} {
			if got := rule.Check(ctx); len(got) != 0 {
				t.Fatalf("%s: %v", exclude, got)
			}
		}
	}
}
func TestAuditEmbeddedInterfaceMethodCount(t *testing.T) {
	ctx := auditContext(t, map[string]string{"internal/domain/order/infra/a.go": "package infra;type inner interface{A();B();C()};type S interface{inner}"}, presets.DDD())
	if got := interfaces.NewTooManyMethods(interfaces.WithMaxMethods(2)).Check(ctx); len(got) != 1 {
		t.Fatalf("got %v", got)
	}
}
func TestAuditGenericInterfaceConstructor(t *testing.T) {
	ctx := auditContext(t, map[string]string{"internal/domain/order/infra/a.go": "package infra;type S[T any] interface{Get() T};func New() S[int]{return nil}"}, presets.DDD())
	if got := interfaces.NewPattern().Check(ctx); len(got) != 0 {
		t.Fatalf("got %v", got)
	}
}

func TestAuditExcludedInterfaceDoesNotFlagIncludedImplementation(t *testing.T) {
	ctx := auditContext(t, map[string]string{
		"internal/domain/order/infra/contract.go": "package infra;type Service interface{Run()}",
		"internal/domain/order/infra/impl.go":     "package infra;type ServiceImpl struct{};func (ServiceImpl) Run(){}",
	}, presets.DDD(), "internal/domain/order/infra/contract.go")
	if got := interfaces.NewPattern().Check(ctx); len(got) != 0 {
		t.Fatalf("excluded contract affected implementation: %v", got)
	}
}

func TestAuditCrossDomainActualPackageName(t *testing.T) {
	ctx := auditContext(t, map[string]string{
		"internal/domain/order/infra/a.go":            "package infra;import `example.com/audit/internal/domain/customer/core/model/v2`;type Adapter struct{Port interface{Get() model.Customer}}",
		"internal/domain/customer/core/model/v2/a.go": "package model;type Customer struct{}",
	}, presets.DDD())
	if got := interfaces.NewCrossDomainAnonymous().Check(ctx); len(got) != 1 {
		t.Fatalf("got %v", got)
	}
}

func TestAuditOtherInterfaceRuleExclusions(t *testing.T) {
	ctx := auditContext(t, map[string]string{
		"internal/domain/order/infra/a.go":         "package infra;import `example.com/audit/internal/domain/customer/core/model`;type Service interface{Run()};type Adapter struct{Service Service;Port interface{Get() model.Customer}}",
		"internal/domain/customer/core/model/a.go": "package model;type Customer struct{}",
	}, presets.DDD(), "internal/domain/order/infra/a.go")
	for _, r := range []core.Rule{interfaces.NewContainer(), interfaces.NewCrossDomainAnonymous()} {
		if got := r.Check(ctx); len(got) != 0 {
			t.Fatalf("%s: %v", r.Spec().ID, got)
		}
	}
}

func TestAuditUntypedLocalInterfaceConstructor(t *testing.T) {
	ctx := auditContext(t, map[string]string{"internal/domain/order/infra/a.go": "package infra;type Service interface{Run()};func New() Service{return nil}"}, presets.DDD())
	for _, pkg := range ctx.Pkgs() {
		pkg.TypesInfo = nil
	}
	if got := interfaces.NewPattern().Check(ctx); len(got) != 0 {
		t.Fatalf("syntax-only local interface: %v", got)
	}
}

func TestAuditTypeParameterIsNotInterfaceReturn(t *testing.T) {
	ctx := auditContext(t, map[string]string{"internal/domain/order/infra/a.go": "package infra;type Service interface{Run()};func New[T any]() T{var value T;return value}"}, presets.DDD())
	got := interfaces.NewPattern().Check(ctx)
	if len(got) != 1 || got[0].Rule != "interfaces.constructor-returns-interface" {
		t.Fatalf("type parameter accepted as interface: %v", got)
	}
}
