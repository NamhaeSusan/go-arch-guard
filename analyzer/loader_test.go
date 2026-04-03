package analyzer_test

import (
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/analyzer"
)

func TestLoad(t *testing.T) {
	t.Run("returns error for nonexistent directory", func(t *testing.T) {
		_, err := analyzer.Load("/nonexistent", "internal/...")
		if err == nil {
			t.Error("expected error for nonexistent directory")
		}
	})

	t.Run("loads packages without requiring successful type checking", func(t *testing.T) {
		pkgs, err := analyzer.Load("../testdata/load_type_error", "internal/...")
		if err != nil {
			t.Fatalf("expected type-check errors to be ignored, got %v", err)
		}
		if len(pkgs) == 0 {
			t.Fatal("expected packages to be loaded")
		}
	})

	t.Run("skips packages with syntax errors and returns the rest", func(t *testing.T) {
		pkgs, err := analyzer.Load("../testdata/load_syntax_error", "internal/...")
		if err == nil {
			t.Fatal("expected error describing skipped packages")
		}
		if len(pkgs) == 0 {
			t.Fatal("expected valid packages to be returned alongside the error")
		}
		// The valid package (user/app) should be loaded; the broken one (order/app) skipped.
		for _, pkg := range pkgs {
			if pkg.PkgPath == "github.com/kimtaeyun/testproject-load-syntax-error/internal/domain/order/app" {
				t.Error("broken package should have been skipped")
			}
		}
	})
}
