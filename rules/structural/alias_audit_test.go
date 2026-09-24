package structural_test

import (
	"github.com/NamhaeSusan/go-arch-guard/analyzer"
	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/rules/structural"
	"os"
	"path/filepath"
	"testing"
)

func TestAliasRejectsGenericContractWithRealPackageName(t *testing.T) {
	root := t.TempDir()
	files := map[string]string{
		"go.mod":                         "module example.com/audit\n\ngo 1.26.1\n",
		"internal/domain/order/alias.go": "package order\nimport `example.com/audit/internal/domain/order/core/repo/v2`\ntype Export[T any] = repo.Store[T]\n",
		"internal/domain/order/core/repo/v2/store.go": "package repo\ntype Store[T any] struct{}\n",
	}
	for name, body := range files {
		p := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
	}
	pkgs, err := analyzer.Load(root, "./...")
	if err != nil {
		t.Fatal(err)
	}
	arch := core.Architecture{Layout: core.LayoutModel{InternalRoot: "internal", DomainDir: "domain"}, Structure: core.StructurePolicy{RequireAlias: true}, Layers: core.LayerModel{Sublayers: []string{"core/repo"}, Direction: map[string][]string{"core/repo": nil}, PortLayers: []string{"core/repo"}, ContractLayers: []string{"core/repo"}}}
	ctx := core.NewContext(pkgs, "example.com/audit", root, arch, nil)
	got := core.Run(ctx, core.NewRuleSet(structural.NewAlias()))
	if len(got) != 1 || got[0].Rule != "structural.domain-alias-contract-reexport" {
		t.Fatalf("got %+v", got)
	}
}
