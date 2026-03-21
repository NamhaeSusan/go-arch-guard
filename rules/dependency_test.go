package rules_test

import (
	"testing"

	"github.com/kimtaeyun/go-arch-guard/analyzer"
	"github.com/kimtaeyun/go-arch-guard/rules"
)

func TestCheckDependency(t *testing.T) {
	t.Run("valid project has no violations", func(t *testing.T) {
		pkgs, err := analyzer.Load("../testdata/valid", "internal/...")
		if err != nil {
			t.Fatal(err)
		}
		violations := rules.CheckDependency(pkgs, "github.com/kimtaeyun/testproject", "../testdata/valid")
		if len(violations) > 0 {
			t.Errorf("expected no violations, got %d: %v", len(violations), violations)
		}
	})

	t.Run("detects handler importing infra", func(t *testing.T) {
		pkgs, err := analyzer.Load("../testdata/invalid", "internal/...")
		if err != nil {
			t.Fatal(err)
		}
		violations := rules.CheckDependency(pkgs, "github.com/kimtaeyun/testproject-invalid", "../testdata/invalid")
		found := findViolation(violations, "dependency.layer-direction")
		if found == nil {
			t.Error("expected layer-direction violation")
		}
	})

	t.Run("detects domain importing app", func(t *testing.T) {
		pkgs, err := analyzer.Load("../testdata/invalid", "internal/...")
		if err != nil {
			t.Fatal(err)
		}
		violations := rules.CheckDependency(pkgs, "github.com/kimtaeyun/testproject-invalid", "../testdata/invalid")
		found := findViolation(violations, "dependency.domain-purity")
		if found == nil {
			t.Error("expected domain-purity violation")
		}
	})

	t.Run("detects cross-domain import", func(t *testing.T) {
		pkgs, err := analyzer.Load("../testdata/invalid", "internal/...")
		if err != nil {
			t.Fatal(err)
		}
		violations := rules.CheckDependency(pkgs, "github.com/kimtaeyun/testproject-invalid", "../testdata/invalid")
		found := findViolation(violations, "dependency.domain-isolation")
		if found == nil {
			t.Error("expected domain-isolation violation")
		}
	})
}
