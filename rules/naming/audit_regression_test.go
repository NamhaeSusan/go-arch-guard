package naming_test

import (
	"github.com/NamhaeSusan/go-arch-guard/analyzer"
	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/presets"
	"github.com/NamhaeSusan/go-arch-guard/rules/naming"
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

func TestAuditDottedFilenames(t *testing.T) {
	for _, tc := range []struct {
		name string
		want int
	}{{"good.BadName.go", 1}, {"good.pb.go", 0}, {"good.vtproto.go", 0}, {"good..go", 1}, {"good_file.go", 0}} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := auditContext(t, map[string]string{"p/" + tc.name: "package p"}, core.Architecture{})
			if got := naming.NewSnakeCaseFiles().Check(ctx); len(got) != tc.want {
				t.Fatalf("got %v want %d", got, tc.want)
			}
		})
	}
}
func TestAuditNestedConstructorRoot(t *testing.T) {
	arch := presets.DDD()
	arch.Layout.InternalRoot = "src/internal"
	ctx := auditContext(t, map[string]string{"src/internal/domain/order/infra/a.go": "package infra;type T struct{};func NewThing()*T{return &T{}}"}, arch)
	if got := naming.NewConstructorName().Check(ctx); len(got) != 1 {
		t.Fatalf("got %v", got)
	}
}
