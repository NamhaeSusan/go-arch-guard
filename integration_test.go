package goarchguard_test

import (
	"testing"

	"github.com/kimtaeyun/go-arch-guard/analyzer"
	"github.com/kimtaeyun/go-arch-guard/report"
	"github.com/kimtaeyun/go-arch-guard/rules"
)

func TestIntegration_Valid(t *testing.T) {
	pkgs, err := analyzer.Load("testdata/valid", "internal/...")
	if err != nil {
		t.Fatal(err)
	}

	t.Run("domain isolation", func(t *testing.T) {
		report.AssertNoViolations(t, rules.CheckDomainIsolation(pkgs, "github.com/kimtaeyun/testproject-dc", "testdata/valid"))
	})
	t.Run("layer direction", func(t *testing.T) {
		report.AssertNoViolations(t, rules.CheckLayerDirection(pkgs, "github.com/kimtaeyun/testproject-dc", "testdata/valid"))
	})
	t.Run("naming", func(t *testing.T) {
		report.AssertNoViolations(t, rules.CheckNaming(pkgs))
	})
}

func TestIntegration_Invalid(t *testing.T) {
	pkgs, err := analyzer.Load("testdata/invalid", "internal/...")
	if err != nil {
		t.Fatal(err)
	}

	t.Run("domain isolation violations found", func(t *testing.T) {
		violations := rules.CheckDomainIsolation(pkgs, "github.com/kimtaeyun/testproject-dc-invalid", "testdata/invalid")
		if len(violations) == 0 {
			t.Error("expected domain isolation violations")
		}
	})
	t.Run("layer direction violations found", func(t *testing.T) {
		violations := rules.CheckLayerDirection(pkgs, "github.com/kimtaeyun/testproject-dc-invalid", "testdata/invalid")
		if len(violations) == 0 {
			t.Error("expected layer direction violations")
		}
	})
}

func TestIntegration_WarningMode(t *testing.T) {
	pkgs, err := analyzer.Load("testdata/invalid", "internal/...")
	if err != nil {
		t.Fatal(err)
	}

	violations := rules.CheckDomainIsolation(pkgs, "github.com/kimtaeyun/testproject-dc-invalid", "testdata/invalid",
		rules.WithSeverity(rules.Warning))
	if len(violations) == 0 {
		t.Error("expected violations even in warning mode")
	}
	for _, v := range violations {
		if v.Severity != rules.Warning {
			t.Errorf("expected Warning severity, got %v", v.Severity)
		}
	}
	report.AssertNoViolations(t, violations)
}
