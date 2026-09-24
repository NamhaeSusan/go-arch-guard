package types_test

import (
	"github.com/NamhaeSusan/go-arch-guard/analyzer"
	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/presets"
	"github.com/NamhaeSusan/go-arch-guard/rules/types"
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

func TestAuditStandardLoggerTermination(t *testing.T) {
	for _, call := range []string{"log.Panic(1)", "log.Panicf(\"%d\",1)", "log.Panicln(1)", "log.Default().Fatal(1)", "log.Default().Fatalf(\"%d\",1)", "log.Default().Fatalln(1)", "log.Default().Panic(1)", "log.Default().Panicf(\"%d\",1)", "log.Default().Panicln(1)"} {
		t.Run(call, func(t *testing.T) {
			ctx := auditContext(t, map[string]string{"internal/domain/order/app/a.go": "package app;import \"log\";func F(){" + call + "}"}, presets.DDD())
			if got := types.NewNoPanicInDomain().Check(ctx); len(got) != 1 {
				t.Fatalf("got %v", got)
			}
		})
	}
}
