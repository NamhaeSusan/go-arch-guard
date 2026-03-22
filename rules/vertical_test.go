package rules_test

import (
	"testing"

	"github.com/kimtaeyun/go-arch-guard/analyzer"
	"github.com/kimtaeyun/go-arch-guard/rules"
)

func TestCheckVerticalSlice(t *testing.T) {
	t.Run("valid project has no violations", func(t *testing.T) {
		pkgs, err := analyzer.Load("../testdata/vertical-valid", "internal/...")
		if err != nil {
			t.Fatal(err)
		}
		violations := rules.CheckVerticalSlice(pkgs, "github.com/kimtaeyun/testproject-vertical", "../testdata/vertical-valid")
		if len(violations) > 0 {
			for _, v := range violations {
				t.Log(v.String())
			}
			t.Errorf("expected no violations, got %d", len(violations))
		}
	})

	t.Run("detects cross-domain violation", func(t *testing.T) {
		pkgs, err := analyzer.Load("../testdata/vertical-invalid", "internal/...")
		if err != nil {
			t.Fatal(err)
		}
		violations := rules.CheckVerticalSlice(pkgs, "github.com/kimtaeyun/testproject-vertical-invalid", "../testdata/vertical-invalid")
		found := findViolation(violations, "vertical.cross-domain-isolation")
		if found == nil {
			t.Error("expected cross-domain-isolation violation")
		}
	})

	t.Run("usecase cross-domain import is allowed", func(t *testing.T) {
		pkgs, err := analyzer.Load("../testdata/vertical-valid", "internal/...")
		if err != nil {
			t.Fatal(err)
		}
		violations := rules.CheckVerticalSlice(pkgs, "github.com/kimtaeyun/testproject-vertical", "../testdata/vertical-valid")
		for _, v := range violations {
			if v.Rule == "vertical.cross-domain-isolation" {
				t.Errorf("usecase cross-domain should be allowed, got: %s", v.String())
			}
		}
	})
}
