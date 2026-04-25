package naming_test

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/analyzer"
	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/rules/naming"
	"golang.org/x/tools/go/packages"
)

var (
	validOnce   sync.Once
	validPkgs   []*packages.Package
	validErr    error
	invalidOnce sync.Once
	invalidPkgs []*packages.Package
	invalidErr  error
)

func loadValid(t *testing.T) []*packages.Package {
	t.Helper()
	validOnce.Do(func() {
		validPkgs, validErr = analyzer.Load("../../testdata/valid", "internal/...")
	})
	if validErr != nil {
		t.Fatal(validErr)
	}
	return validPkgs
}

func loadInvalid(t *testing.T) []*packages.Package {
	t.Helper()
	invalidOnce.Do(func() {
		invalidPkgs, invalidErr = analyzer.Load("../../testdata/invalid", "internal/...")
	})
	if invalidErr != nil {
		t.Fatal(invalidErr)
	}
	return invalidPkgs
}

func newContext(pkgs []*packages.Package, module, root string, arch core.Architecture, exclude []string) *core.Context {
	return core.NewContext(pkgs, module, root, arch, exclude)
}

func validContext(t *testing.T) *core.Context {
	t.Helper()
	return newContext(loadValid(t), "github.com/kimtaeyun/testproject-dc", "../../testdata/valid", dddArch(), nil)
}

func invalidContext(t *testing.T, exclude []string) *core.Context {
	t.Helper()
	return newContext(loadInvalid(t), "github.com/kimtaeyun/testproject-dc-invalid", "../../testdata/invalid", dddArch(), exclude)
}

func tempContext(t *testing.T, files map[string]string, arch core.Architecture) *core.Context {
	t.Helper()
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), "module example.com/namingtest\n\ngo 1.25.0\n")
	for name, content := range files {
		writeFile(t, filepath.Join(root, name), content)
	}
	pkgs, err := analyzer.Load(root, "internal/...")
	if err != nil {
		t.Fatal(err)
	}
	return newContext(pkgs, "example.com/namingtest", root, arch, nil)
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func dddArch() core.Architecture {
	return core.Architecture{
		Layers: core.LayerModel{
			Sublayers:      []string{"handler", "app", "core", "core/model", "core/repo", "core/svc", "event", "infra"},
			PortLayers:     []string{"core/repo"},
			ContractLayers: []string{"core/repo", "core/svc"},
			LayerDirNames: map[string]bool{
				"handler": true, "app": true, "core": true, "model": true,
				"repo": true, "svc": true, "event": true, "infra": true,
			},
		},
		Layout: core.LayoutModel{
			DomainDir:        "domain",
			OrchestrationDir: "orchestration",
			SharedDir:        "pkg",
		},
		Structure: core.StructurePolicy{
			RequireAlias: true,
		},
	}
}

func TestNoStutterSpec(t *testing.T) {
	spec := naming.NewNoStutter(naming.WithSeverity(core.Error)).Spec()

	if spec.ID != "naming.no-stutter" {
		t.Fatalf("ID = %q, want naming.no-stutter", spec.ID)
	}
	if spec.DefaultSeverity != core.Error {
		t.Fatalf("DefaultSeverity = %v, want Error", spec.DefaultSeverity)
	}
}

func TestNoStutterFlagsExportedTypePrefix(t *testing.T) {
	ctx := tempContext(t, map[string]string{
		"internal/domain/order/order.go": "package order\n\ntype OrderService struct{}\ntype localService struct{}\n",
	}, dddArch())
	got := naming.NewNoStutter().Check(ctx)

	var found bool
	for _, v := range got {
		if v.Rule == "naming.no-stutter" && strings.Contains(v.Message, `"OrderService"`) {
			found = true
			if v.DefaultSeverity != core.Warning || v.EffectiveSeverity != core.Warning {
				t.Fatalf("severity = default %v effective %v, want Warning/Warning", v.DefaultSeverity, v.EffectiveSeverity)
			}
		}
	}
	if !found {
		t.Fatalf("expected no-stutter violation for order.OrderService; got %+v", got)
	}
}

func TestNoStutterSkipsValidProject(t *testing.T) {
	if got := naming.NewNoStutter().Check(validContext(t)); len(got) != 0 {
		t.Fatalf("valid project got no-stutter violations: %+v", got)
	}
}
