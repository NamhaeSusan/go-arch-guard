package tui_test

import (
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/analyzer"
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

	tree := tui.BuildTree(pkgs, module)
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
	tree := tui.BuildTree(nil, "example.com/test")
	root := tree.GetRoot()
	if root == nil {
		t.Fatal("tree root is nil for empty packages")
	}
	if len(root.GetChildren()) != 0 {
		t.Error("expected no children for empty packages")
	}
}
