package rules_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/analyzer"
	"github.com/NamhaeSusan/go-arch-guard/rules"
)

func TestCheckNaming(t *testing.T) {
	t.Run("valid project has no violations", func(t *testing.T) {
		pkgs, err := analyzer.Load("../testdata/valid", "internal/...")
		if err != nil {
			t.Fatal(err)
		}
		violations := rules.CheckNaming(pkgs)
		if len(violations) > 0 {
			for _, v := range violations {
				t.Log(v.String())
			}
			t.Errorf("expected no violations, got %d", len(violations))
		}
	})

	t.Run("detects handler exported interface", func(t *testing.T) {
		pkgs, err := analyzer.Load("../testdata/invalid", "internal/...")
		if err != nil {
			t.Fatal(err)
		}
		violations := rules.CheckNaming(pkgs)
		found := false
		for _, v := range violations {
			if v.Rule == "naming.handler-no-exported-interface" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected handler-no-exported-interface violation")
		}
	})

	t.Run("reports relative file paths", func(t *testing.T) {
		pkgs, err := analyzer.Load("../testdata/invalid", "internal/...")
		if err != nil {
			t.Fatal(err)
		}
		violations := rules.CheckNaming(pkgs)
		if len(violations) == 0 {
			t.Fatal("expected naming violations")
		}
		for _, v := range violations {
			if filepath.IsAbs(v.File) {
				t.Fatalf("expected relative path, got %q", v.File)
			}
		}
	})

	t.Run("detects non-snake-case filename and suggests fix", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, filepath.Join(root, "internal", "domain", "order", "app", "createOrder.go"),
			"package app\n\ntype Request struct{}\n")
		writeFile(t, filepath.Join(root, "go.mod"), "module example.com/snaketest\n\ngo 1.25.0\n")

		pkgs, err := analyzer.Load(root, "internal/...")
		if err != nil {
			t.Fatal(err)
		}
		violations := rules.CheckNaming(pkgs)
		found := false
		for _, v := range violations {
			if v.Rule == "naming.snake-case-file" {
				found = true
				if !strings.Contains(v.Fix, "create_order.go") {
					t.Errorf("expected fix to suggest snake_case name, got %q", v.Fix)
				}
				break
			}
		}
		if !found {
			t.Error("expected snake-case-file violation for createOrder.go")
		}
	})

	t.Run("project-relative exclude skips matching files", func(t *testing.T) {
		pkgs, err := analyzer.Load("../testdata/invalid", "internal/...")
		if err != nil {
			t.Fatal(err)
		}
		violations := rules.CheckNaming(pkgs, rules.WithExclude("internal/domain/order/handler/..."))
		for _, v := range violations {
			if v.File == "internal/domain/order/handler/http/bad_handler.go" {
				t.Fatalf("expected order handler naming violations to be excluded, got %s", v.String())
			}
		}
	})
}
