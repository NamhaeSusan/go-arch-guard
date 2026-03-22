package goarchguard_test

import (
	"testing"

	"github.com/kimtaeyun/go-arch-guard/analyzer"
	"github.com/kimtaeyun/go-arch-guard/report"
	"github.com/kimtaeyun/go-arch-guard/rules"
)

func TestIntegration_ValidProject(t *testing.T) {
	pkgs, err := analyzer.Load("testdata/valid", "internal/...")
	if err != nil {
		t.Fatal(err)
	}

	t.Run("dependency", func(t *testing.T) {
		report.AssertNoViolations(t, rules.CheckDependency(pkgs, "github.com/kimtaeyun/testproject", "testdata/valid"))
	})
	t.Run("naming", func(t *testing.T) {
		report.AssertNoViolations(t, rules.CheckNaming(pkgs))
	})
	t.Run("structure", func(t *testing.T) {
		report.AssertNoViolations(t, rules.CheckStructure("testdata/valid"))
	})
}

func TestIntegration_InvalidProject(t *testing.T) {
	pkgs, err := analyzer.Load("testdata/invalid", "internal/...")
	if err != nil {
		t.Fatal(err)
	}

	t.Run("dependency violations found", func(t *testing.T) {
		violations := rules.CheckDependency(pkgs, "github.com/kimtaeyun/testproject-invalid", "testdata/invalid")
		if len(violations) == 0 {
			t.Error("expected dependency violations")
		}
	})
	t.Run("naming violations found", func(t *testing.T) {
		violations := rules.CheckNaming(pkgs)
		if len(violations) == 0 {
			t.Error("expected naming violations")
		}
	})
	t.Run("structure violations found", func(t *testing.T) {
		violations := rules.CheckStructure("testdata/invalid")
		if len(violations) == 0 {
			t.Error("expected structure violations")
		}
	})
}

func TestIntegration_WarningMode(t *testing.T) {
	pkgs, err := analyzer.Load("testdata/invalid", "internal/...")
	if err != nil {
		t.Fatal(err)
	}

	// Warning mode: violations exist but test passes
	violations := rules.CheckDependency(pkgs, "github.com/kimtaeyun/testproject-invalid", "testdata/invalid",
		rules.WithSeverity(rules.Warning))
	if len(violations) == 0 {
		t.Error("expected violations even in warning mode")
	}
	for _, v := range violations {
		if v.Severity != rules.Warning {
			t.Errorf("expected Warning severity, got %v", v.Severity)
		}
	}
	// AssertNoViolations should pass with warnings only
	report.AssertNoViolations(t, violations)
}

func TestIntegration_VerticalValid(t *testing.T) {
	pkgs, err := analyzer.Load("testdata/vertical-valid", "internal/...")
	if err != nil {
		t.Fatal(err)
	}

	t.Run("cross-domain", func(t *testing.T) {
		report.AssertNoViolations(t, rules.CheckVerticalSlice(pkgs, "github.com/kimtaeyun/testproject-vertical", "testdata/vertical-valid"))
	})
	t.Run("internal-direction", func(t *testing.T) {
		report.AssertNoViolations(t, rules.CheckVerticalSliceInternal(pkgs, "github.com/kimtaeyun/testproject-vertical", "testdata/vertical-valid"))
	})
}

func TestIntegration_VerticalInvalid(t *testing.T) {
	pkgs, err := analyzer.Load("testdata/vertical-invalid", "internal/...")
	if err != nil {
		t.Fatal(err)
	}

	t.Run("cross-domain violations found", func(t *testing.T) {
		violations := rules.CheckVerticalSlice(pkgs, "github.com/kimtaeyun/testproject-vertical-invalid", "testdata/vertical-invalid")
		if len(violations) == 0 {
			t.Error("expected cross-domain violations")
		}
	})
	t.Run("internal-direction violations found", func(t *testing.T) {
		violations := rules.CheckVerticalSliceInternal(pkgs, "github.com/kimtaeyun/testproject-vertical-invalid", "testdata/vertical-invalid")
		if len(violations) == 0 {
			t.Error("expected internal-direction violations")
		}
	})
}
