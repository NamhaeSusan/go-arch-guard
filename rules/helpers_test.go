package rules

import (
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/analyzer"
	"golang.org/x/tools/go/packages"
)

func TestFindImportPosition(t *testing.T) {
	pkgs, err := analyzer.Load("../testdata/invalid", "internal/...")
	if err != nil {
		t.Fatal(err)
	}

	// Find the order/app package which imports core/model
	var appPkg *packages.Package
	for _, pkg := range pkgs {
		if pkg.PkgPath == "github.com/kimtaeyun/testproject-dc-invalid/internal/domain/order/app" {
			appPkg = pkg
			break
		}
	}
	if appPkg == nil {
		t.Fatal("order/app package not found")
	}

	importPath := "github.com/kimtaeyun/testproject-dc-invalid/internal/domain/order/core/model"
	file, line := findImportPosition(appPkg, importPath, "../testdata/invalid")

	if file == "" {
		t.Error("expected non-empty file")
	}
	if line == 0 {
		t.Error("expected non-zero line")
	}
}

func TestResolveModule(t *testing.T) {
	pkgs, err := analyzer.Load("../testdata/valid", "internal/...")
	if err != nil {
		t.Fatal(err)
	}
	mod := resolveModule(pkgs, "")
	if mod != "github.com/kimtaeyun/testproject-dc" {
		t.Errorf("got %q, want github.com/kimtaeyun/testproject-dc", mod)
	}

	// explicit value passes through
	mod = resolveModule(pkgs, "custom/module")
	if mod != "custom/module" {
		t.Errorf("got %q, want custom/module", mod)
	}
}

func TestResolveRoot(t *testing.T) {
	pkgs, err := analyzer.Load("../testdata/valid", "internal/...")
	if err != nil {
		t.Fatal(err)
	}
	root := resolveRoot(pkgs, "")
	if root == "" {
		t.Error("expected non-empty root from auto-extraction")
	}

	// explicit value passes through
	root = resolveRoot(pkgs, "/custom/root")
	if root != "/custom/root" {
		t.Errorf("got %q, want /custom/root", root)
	}
}
