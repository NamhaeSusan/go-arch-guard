package analysisutil

import (
	"path/filepath"
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/analyzer"
	"golang.org/x/tools/go/packages"
)

func TestRelPathFromRoot(t *testing.T) {
	root := filepath.Join(string(filepath.Separator), "tmp", "project")
	file := filepath.Join(root, "internal", "domain", "order", "app", "service.go")

	got := RelPathFromRoot(root, file)
	want := "internal/domain/order/app/service.go"
	if got != want {
		t.Fatalf("RelPathFromRoot() = %q, want %q", got, want)
	}
}

func TestResolveModuleAndRoot(t *testing.T) {
	pkgs, err := analyzer.Load("../../testdata/valid", "internal/...")
	if err != nil {
		t.Fatal(err)
	}

	if got := ResolveModule(pkgs, ""); got != "github.com/kimtaeyun/testproject-dc" {
		t.Fatalf("ResolveModule() = %q", got)
	}
	if got := ResolveModule(pkgs, "custom/module"); got != "custom/module" {
		t.Fatalf("ResolveModule(explicit) = %q", got)
	}

	if got := ResolveRoot(pkgs, ""); got == "" {
		t.Fatal("ResolveRoot() = empty")
	}
	if got := ResolveRoot(pkgs, "/custom/root"); got != "/custom/root" {
		t.Fatalf("ResolveRoot(explicit) = %q", got)
	}
}

func TestFindImportPosition(t *testing.T) {
	pkgs, err := analyzer.Load("../../testdata/invalid", "internal/...")
	if err != nil {
		t.Fatal(err)
	}

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
	file, line := FindImportPosition(appPkg, importPath, "../../testdata/invalid")
	if file == "" {
		t.Fatal("FindImportPosition() returned empty file")
	}
	if line == 0 {
		t.Fatal("FindImportPosition() returned line 0")
	}
}

func TestProjectRelativePackagePath(t *testing.T) {
	tests := []struct {
		name       string
		pkgPath    string
		modulePath string
		want       string
	}{
		{name: "module root", pkgPath: "example.com/app", modulePath: "example.com/app", want: "."},
		{name: "descendant", pkgPath: "example.com/app/internal/order", modulePath: "example.com/app", want: "internal/order"},
		{name: "outside", pkgPath: "example.com/other", modulePath: "example.com/app", want: ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := ProjectRelativePackagePath(tc.pkgPath, tc.modulePath); got != tc.want {
				t.Fatalf("ProjectRelativePackagePath() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestNormalizeMatchPath(t *testing.T) {
	got := NormalizeMatchPath("./internal\\domain/order")
	want := "internal/domain/order"
	if got != want {
		t.Fatalf("NormalizeMatchPath() = %q, want %q", got, want)
	}
}
