package analyzer_test

import (
	"strings"
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
		// Regression guard for issue #28: this test must prove that the
		// dep-walk path (firstFatalErrorInDeps) actually rejected the root,
		// not that the root happened to fail for an unrelated reason. The
		// dep-walk path produces error messages prefixed with "dependency ";
		// root-level handling does not. If a future refactor moves
		// classification entirely to the root path, this assertion fails
		// and forces a deliberate update.
		if !strings.Contains(err.Error(), "dependency ") {
			t.Errorf("expected error to surface dep-walk path with %q prefix, got %q", "dependency ", err.Error())
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

	t.Run("error message is bounded when many packages fail", func(t *testing.T) {
		// Regression guard: summarizeLoadErrs caps the joined error string at
		// the first few entries plus a count. We can't easily produce 6+
		// failing packages from fixtures, but we can at least assert the
		// existing fixture's message reads naturally and includes the count
		// prefix.
		_, err := analyzer.Load("../testdata/load_syntax_error", "internal/...")
		if err == nil {
			t.Fatal("expected error from load_syntax_error")
		}
		if !strings.Contains(err.Error(), "packages with errors were skipped") {
			t.Errorf("error message must keep its leading description, got %q", err.Error())
		}
	})

	t.Run("absolute module-path patterns are passed through unchanged", func(t *testing.T) {
		// Regression guard: prior version unconditionally prefixed every
		// pattern with "./", which broke patterns that were already module
		// paths (`github.com/foo/bar/...` -> `./github.com/foo/bar/...`).
		// The fixture is unimportant; we just need Load to not blow up on
		// the prefix transform.
		_, err := analyzer.Load("../testdata/load_type_error", "github.com/kimtaeyun/testproject-load-type-error/internal/...")
		if err != nil {
			// the fixture has only type errors, so a non-nil err here would
			// be from pattern handling, not classification
			if strings.Contains(err.Error(), "./github.com/") {
				t.Fatalf("loader double-prefixed an absolute module pattern: %v", err)
			}
		}
	})
}
