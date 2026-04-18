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

	t.Run("skips root package whose dependency has a syntax error", func(t *testing.T) {
		pkgs, err := analyzer.Load("../testdata/load_broken_dep", "internal/...")
		if err == nil {
			t.Fatal("expected error describing skipped packages with broken dependencies")
		}
		// The root package imports a broken dep; it must not appear in results.
		for _, pkg := range pkgs {
			if pkg.PkgPath == "github.com/kimtaeyun/testproject-load-broken-dep/internal/root" {
				t.Error("root package with broken dependency should have been skipped")
			}
		}
	})

	t.Run("skips root package whose transitive dependency has a syntax error", func(t *testing.T) {
		pkgs, err := analyzer.Load("../testdata/load_broken_dep_transitive", "internal/...")
		if err == nil {
			t.Fatal("expected error describing skipped packages with broken transitive dependencies")
		}
		// root -> middle -> broken: both root and middle must be skipped;
		// the unrelated clean package must still be returned.
		var sawClean bool
		for _, pkg := range pkgs {
			switch pkg.PkgPath {
			case "github.com/kimtaeyun/testproject-load-broken-dep-transitive/internal/root":
				t.Error("root (transitive broken dep) should have been skipped")
			case "github.com/kimtaeyun/testproject-load-broken-dep-transitive/internal/middle":
				t.Error("middle (direct broken dep) should have been skipped")
			case "github.com/kimtaeyun/testproject-load-broken-dep-transitive/internal/clean":
				sawClean = true
			}
		}
		if !sawClean {
			t.Error("expected unrelated clean package to still be loaded")
		}
	})

	t.Run("skips root package whose dependency has a list error", func(t *testing.T) {
		pkgs, err := analyzer.Load("../testdata/load_missing_dep", "internal/...")
		if err == nil {
			t.Fatal("expected error describing skipped packages with missing dependencies")
		}
		var sawClean bool
		for _, pkg := range pkgs {
			switch pkg.PkgPath {
			case "github.com/kimtaeyun/testproject-load-missing-dep/internal/root":
				t.Error("root (missing dep, ListError) should have been skipped")
			case "github.com/kimtaeyun/testproject-load-missing-dep/internal/clean":
				sawClean = true
			}
		}
		if !sawClean {
			t.Error("expected unrelated clean package to still be loaded")
		}
	})
}
