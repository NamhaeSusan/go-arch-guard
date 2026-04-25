package tui_test

import (
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/analyzer"
	"github.com/NamhaeSusan/go-arch-guard/presets"
	"github.com/NamhaeSusan/go-arch-guard/tui"
)

func TestBuildTree_ValidProject(t *testing.T) {
	pkgs, err := analyzer.Load("../testdata/valid", "internal/...", "cmd/...")
	if err != nil {
		t.Logf("partial load: %v", err)
	}
	if len(pkgs) == 0 {
		t.Fatal("no packages loaded")
	}

	module := ""
	for _, pkg := range pkgs {
		if pkg.Module != nil {
			module = pkg.Module.Path
			break
		}
	}

	tree := tui.BuildTree(pkgs, module, nil)
	root := tree.GetRoot()
	if root == nil {
		t.Fatal("tree root is nil")
	}
	if len(root.GetChildren()) == 0 {
		t.Fatal("tree has no children")
	}
}

func TestBuildImportedByMap(t *testing.T) {
	pkgs, err := analyzer.Load("../testdata/valid", "internal/...", "cmd/...")
	if err != nil {
		t.Logf("partial load: %v", err)
	}
	if len(pkgs) == 0 {
		t.Fatal("no packages loaded")
	}

	importedBy := tui.BuildImportedByMap(pkgs)
	if len(importedBy) == 0 {
		t.Fatal("importedBy map is empty")
	}
}

func TestBuildTree_EmptyPackages(t *testing.T) {
	tree := tui.BuildTree(nil, "example.com/test", nil)
	root := tree.GetRoot()
	if root == nil {
		t.Fatal("tree root is nil for empty packages")
	}
	if len(root.GetChildren()) != 0 {
		t.Error("expected no children for empty packages")
	}
}

func TestBuildViolationIndex(t *testing.T) {
	pkgs, err := analyzer.Load("../testdata/invalid", "internal/...", "cmd/...")
	if err != nil {
		t.Logf("partial load: %v", err)
	}
	if len(pkgs) == 0 {
		t.Skip("no packages loaded from testdata/invalid")
	}

	module := ""
	root := ""
	for _, pkg := range pkgs {
		if pkg.Module != nil {
			module = pkg.Module.Path
			root = pkg.Module.Dir
			break
		}
	}

	violations := tui.BuildViolationIndex(pkgs, module, root, presets.DDD(), presets.RecommendedDDD())
	if len(violations) == 0 {
		t.Error("expected violations from invalid testdata")
	}
}

func TestTUIDoesNotImportV1Rules(t *testing.T) {
	v1RulesImport := "github.com/NamhaeSusan/go-arch-guard/" + "rules"
	files := []string{
		"app.go",
		"detail.go",
		"violations.go",
		"../cmd/tui/main.go",
	}

	for _, file := range files {
		t.Run(file, func(t *testing.T) {
			src, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			parsed, err := parser.ParseFile(token.NewFileSet(), file, src, parser.ImportsOnly)
			if err != nil {
				t.Fatal(err)
			}
			for _, imp := range parsed.Imports {
				if strings.Trim(imp.Path.Value, `"`) == v1RulesImport {
					t.Fatal("tui must use core API instead of v1 rules package")
				}
			}
		})
	}
}

func TestBuildMetricsIndex(t *testing.T) {
	pkgs, err := analyzer.Load("../testdata/valid", "internal/...", "cmd/...")
	if err != nil {
		t.Logf("partial load: %v", err)
	}
	if len(pkgs) == 0 {
		t.Fatal("no packages loaded")
	}

	module := ""
	for _, pkg := range pkgs {
		if pkg.Module != nil {
			module = pkg.Module.Path
			break
		}
	}

	metrics := tui.BuildMetricsIndex(pkgs, module)
	if len(metrics) == 0 {
		t.Error("expected metrics for internal packages")
	}

	// Verify all metrics have non-negative values.
	for path, m := range metrics {
		if m.Ca < 0 || m.Ce < 0 {
			t.Errorf("negative coupling for %s: Ca=%d Ce=%d", path, m.Ca, m.Ce)
		}
		if m.Instability < 0 || m.Instability > 1 {
			t.Errorf("instability out of range for %s: %.2f", path, m.Instability)
		}
	}
}
