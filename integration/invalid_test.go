package integration_test

import (
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/analyzer"
	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/presets"
	"github.com/NamhaeSusan/go-arch-guard/report"
)

func TestIntegration_Invalid(t *testing.T) {
	pkgs, err := analyzer.Load(fixturePath("testdata/invalid"), "internal/...", "cmd/...")
	if err != nil {
		t.Fatal(err)
	}

	t.Run("domain isolation violations found", func(t *testing.T) {
		violations := runDDD(pkgs, "github.com/kimtaeyun/testproject-dc-invalid", fixturePath("testdata/invalid"))
		assertHasRule(t, violations, "isolation.cross-domain")
	})
	t.Run("layer direction violations found", func(t *testing.T) {
		violations := runDDD(pkgs, "github.com/kimtaeyun/testproject-dc-invalid", fixturePath("testdata/invalid"))
		assertHasRule(t, violations, "layer.direction")
	})
	t.Run("naming violations found", func(t *testing.T) {
		violations := runDDD(pkgs, "github.com/kimtaeyun/testproject-dc-invalid", fixturePath("testdata/invalid"))
		assertHasRule(t, violations, "naming.no-layer-suffix")
	})
	t.Run("structure violations found", func(t *testing.T) {
		violations := runDDD(pkgs, "github.com/kimtaeyun/testproject-dc-invalid", fixturePath("testdata/invalid"))
		assertHasRule(t, violations, "structure.banned-package")
	})

	t.Run("new rule ids are surfaced", func(t *testing.T) {
		isolationViolations := runDDD(pkgs, "github.com/kimtaeyun/testproject-dc-invalid", fixturePath("testdata/invalid"))
		layerViolations := runDDD(pkgs, "github.com/kimtaeyun/testproject-dc-invalid", fixturePath("testdata/invalid"))
		structureViolations := runDDD(pkgs, "github.com/kimtaeyun/testproject-dc-invalid", fixturePath("testdata/invalid"))

		assertHasRule(t, isolationViolations, "isolation.domain-imports-orchestration")
		assertHasRule(t, isolationViolations, "isolation.stray-imports-orchestration")
		assertHasRule(t, isolationViolations, "isolation.pkg-imports-domain")
		assertHasRule(t, layerViolations, "layer.unknown-sublayer")
		assertHasRule(t, layerViolations, "layer.inner-imports-pkg")
		assertHasRule(t, structureViolations, "structure.internal-top-level")
		assertHasRule(t, structureViolations, "structure.domain-alias-exists")
		assertHasRule(t, structureViolations, "structure.domain-model-required")
		assertHasRule(t, structureViolations, "structure.dto-placement")
		assertHasRule(t, structureViolations, "structure.misplaced-layer")
	})
}

func TestIntegration_WarningMode(t *testing.T) {
	pkgs, err := analyzer.Load(fixturePath("testdata/invalid"), "internal/...", "cmd/...")
	if err != nil {
		t.Fatal(err)
	}

	violations := runArchitectureAsWarnings(pkgs, "github.com/kimtaeyun/testproject-dc-invalid", fixturePath("testdata/invalid"), presets.DDD(), presets.RecommendedDDD())
	if len(violations) == 0 {
		t.Error("expected violations even in warning mode")
	}
	for _, v := range violations {
		if v.EffectiveSeverity != core.Warning {
			t.Errorf("expected Warning severity, got %v", v.EffectiveSeverity)
		}
	}
	report.AssertNoViolations(t, violations)
}
