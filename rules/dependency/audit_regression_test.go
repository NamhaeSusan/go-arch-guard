package dependency_test

import (
	"github.com/NamhaeSusan/go-arch-guard/analyzer"
	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/presets"
	"github.com/NamhaeSusan/go-arch-guard/rules/dependency"
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

func TestAuditDomainCannotImportComposition(t *testing.T) {
	for _, dest := range []string{"app", "server/http"} {
		t.Run(dest, func(t *testing.T) {
			ctx := auditContext(t, map[string]string{
				"internal/domain/order/core/model/a.go": "package model;import _ \"example.com/audit/internal/" + dest + "\"",
				"internal/" + dest + "/a.go":            "package composition",
			}, presets.DDD())
			got := core.Run(ctx, core.NewRuleSet(dependency.NewIsolation(), dependency.NewLayerDirection()))
			if len(got) != 1 || got[0].Rule != "dependency.domain-imports-composition" {
				t.Fatalf("got %v", got)
			}
		})
	}
}
func TestAuditNestedPurityRoot(t *testing.T) {
	arch := presets.DDD()
	arch.Layout.InternalRoot = "src/internal"
	ctx := auditContext(t, map[string]string{"src/internal/domain/order/core/model/a.go": "package model;import \"time\";func F(){_ = time.Now()}"}, arch)
	if got := dependency.NewNoSideEffectCallInCore().Check(ctx); len(got) != 1 {
		t.Fatalf("got %v", got)
	}
}
