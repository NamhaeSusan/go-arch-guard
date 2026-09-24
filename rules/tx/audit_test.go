package tx_test

import (
	"github.com/NamhaeSusan/go-arch-guard/analyzer"
	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/rules/tx"
	"os"
	"path/filepath"
	"testing"
)

func TestBoundaryCallShapesNestedTypesAndFileExclusion(t *testing.T) {
	root := t.TempDir()
	files := map[string]string{
		"go.mod": "module example.com/audit\n\ngo 1.26.1\n",
		"internal/model/model.go": `package model
 type Tx struct{}
 type Alias = Tx
 func Start[T any](v T) {}
 func Plain() {}
 func Use(a Alias, b map[*Tx]string, c func() *Tx) { Start(1); Start[int](1); (Plain)(); Plain() }
 `,
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
	arch := core.Architecture{Layers: core.LayerModel{Sublayers: []string{"model"}, Direction: map[string][]string{"model": nil}}}
	rule := tx.New(tx.Config{StartSymbols: []string{"example.com/audit/internal/model.Start", "example.com/audit/internal/model.Plain"}, Types: []string{"example.com/audit/internal/model.Tx"}})
	for _, exclude := range [][]string{nil, {"internal/model/model.go"}} {
		ctx := core.NewContext(pkgs, "example.com/audit", root, arch, exclude)
		got := core.Run(ctx, core.NewRuleSet(rule))
		want := 7
		if exclude != nil {
			want = 0
		}
		if len(got) != want {
			t.Fatalf("exclude %v: got %d want %d: %+v", exclude, len(got), want, got)
		}
	}
}
