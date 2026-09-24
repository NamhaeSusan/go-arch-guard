package testpolicy_test

import (
	"github.com/NamhaeSusan/go-arch-guard/analyzer"
	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/presets"
	"github.com/NamhaeSusan/go-arch-guard/rules/testpolicy"
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

func TestAuditMockMethodsAcrossFiles(t *testing.T) {
	for _, tc := range []struct {
		name, methods string
		exclude       []string
		want          int
	}{
		{"split", "package p;func (*mockStore) Get(){}", nil, 1},
		{"excluded", "package p;func (*mockStore) Get(){}", []string{"p/methods_test.go"}, 0},
		{"external-package", "package p_test;type mockStore struct{};func (*mockStore) Get(){}", nil, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := auditContext(t, map[string]string{"p/a.go": "package p", "p/mock_test.go": "package p;type mockStore struct{}", "p/methods_test.go": tc.methods}, presets.DDD(), tc.exclude...)
			got := testpolicy.NewNoHandMock().Check(ctx)
			if len(got) != tc.want {
				t.Fatalf("got %v", got)
			}
			if tc.name == "external-package" && got[0].File != "p/methods_test.go" {
				t.Fatalf("cross-package match: %v", got)
			}
		})
	}
}
