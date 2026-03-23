package goarchguard_test

import (
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/analyzer"
	"github.com/NamhaeSusan/go-arch-guard/report"
	"github.com/NamhaeSusan/go-arch-guard/rules"
)

func TestIntegration_Valid(t *testing.T) {
	pkgs, err := analyzer.Load("testdata/valid", "internal/...", "cmd/...")
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
	t.Run("structure", func(t *testing.T) {
		report.AssertNoViolations(t, rules.CheckStructure("testdata/valid"))
	})
}

func TestIntegration_Invalid(t *testing.T) {
	pkgs, err := analyzer.Load("testdata/invalid", "internal/...", "cmd/...")
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

	t.Run("new rule ids are surfaced", func(t *testing.T) {
		isolationViolations := rules.CheckDomainIsolation(pkgs, "github.com/kimtaeyun/testproject-dc-invalid", "testdata/invalid")
		layerViolations := rules.CheckLayerDirection(pkgs, "github.com/kimtaeyun/testproject-dc-invalid", "testdata/invalid")
		structureViolations := rules.CheckStructure("testdata/invalid")

		assertHasRule(t, isolationViolations, "isolation.domain-imports-orchestration")
		assertHasRule(t, isolationViolations, "isolation.internal-imports-orchestration")
		assertHasRule(t, isolationViolations, "isolation.pkg-imports-domain")
		assertHasRule(t, layerViolations, "layer.unknown-sublayer")
		assertHasRule(t, layerViolations, "layer.inner-imports-pkg")
		assertHasRule(t, structureViolations, "structure.domain-root-alias-required")
		assertHasRule(t, structureViolations, "structure.domain-model-required")
		assertHasRule(t, structureViolations, "structure.dto-placement")
	})
}

func TestIntegration_WarningMode(t *testing.T) {
	pkgs, err := analyzer.Load("testdata/invalid", "internal/...", "cmd/...")
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

func assertHasRule(t *testing.T, violations []rules.Violation, rule string) {
	t.Helper()
	for _, v := range violations {
		if v.Rule == rule {
			return
		}
	}
	t.Fatalf("expected rule %q", rule)
}
