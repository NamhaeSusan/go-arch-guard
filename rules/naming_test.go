package rules_test

import (
	"testing"

	"github.com/kimtaeyun/go-arch-guard/analyzer"
	"github.com/kimtaeyun/go-arch-guard/rules"
)

func TestCheckNaming(t *testing.T) {
	t.Run("valid project has no violations", func(t *testing.T) {
		pkgs, err := analyzer.Load("../testdata/valid", "internal/...")
		if err != nil {
			t.Fatal(err)
		}
		violations := rules.CheckNaming(pkgs)
		if len(violations) > 0 {
			t.Errorf("expected no violations, got %d: %v", len(violations), violations)
		}
	})

	t.Run("detects package stutter", func(t *testing.T) {
		pkgs, err := analyzer.Load("../testdata/invalid", "internal/...")
		if err != nil {
			t.Fatal(err)
		}
		violations := rules.CheckNaming(pkgs)
		found := findViolation(violations, "naming.no-stutter")
		if found == nil {
			t.Error("expected no-stutter violation for UserService")
		}
	})

	t.Run("detects Impl suffix", func(t *testing.T) {
		pkgs, err := analyzer.Load("../testdata/invalid", "internal/...")
		if err != nil {
			t.Fatal(err)
		}
		violations := rules.CheckNaming(pkgs)
		found := findViolation(violations, "naming.no-impl-suffix")
		if found == nil {
			t.Error("expected no-impl-suffix violation for RepoImpl")
		}
	})

	t.Run("detects non-snake-case filenames", func(t *testing.T) {
		pkgs, err := analyzer.Load("../testdata/invalid", "internal/...")
		if err != nil {
			t.Fatal(err)
		}
		violations := rules.CheckNaming(pkgs)
		found := findViolation(violations, "naming.snake-case-file")
		if found == nil {
			t.Error("expected snake-case-file violation for userService.go")
		}
	})

	t.Run("detects repo file without matching interface", func(t *testing.T) {
		pkgs, err := analyzer.Load("../testdata/invalid", "internal/...")
		if err != nil {
			t.Fatal(err)
		}
		violations := rules.CheckNaming(pkgs)
		found := findViolation(violations, "naming.repo-file-interface")
		if found == nil {
			t.Error("expected repo-file-interface violation for repo/user.go")
		}
	})

	t.Run("detects layer suffix in filename", func(t *testing.T) {
		pkgs, err := analyzer.Load("../testdata/invalid", "internal/...")
		if err != nil {
			t.Fatal(err)
		}
		violations := rules.CheckNaming(pkgs)
		found := findViolation(violations, "naming.no-layer-suffix")
		if found == nil {
			t.Error("expected no-layer-suffix violation for order_svc.go")
		}
	})
}
