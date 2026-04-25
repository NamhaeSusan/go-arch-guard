package dependency_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/rules/dependency"
)

func TestBlastRadiusSpec(t *testing.T) {
	rule := dependency.NewBlastRadius(dependency.WithSeverity(core.Error))

	spec := rule.Spec()
	if spec.ID != "dependency.blast-radius" {
		t.Fatalf("Spec().ID = %q, want dependency.blast-radius", spec.ID)
	}
	if spec.DefaultSeverity != core.Error {
		t.Fatalf("Spec().DefaultSeverity = %v, want %v", spec.DefaultSeverity, core.Error)
	}
	if !slices.Contains(spec.ViolationIDs(), "blast.high-coupling") {
		t.Fatalf("Spec().ViolationIDs() missing blast.high-coupling")
	}
}

func TestBlastRadiusValidProject(t *testing.T) {
	ctx := loadContext(t, "../../testdata/valid", "github.com/kimtaeyun/testproject-dc", dddArchitecture(), "internal/...")

	violations := dependency.NewBlastRadius().Check(ctx)
	// A valid project must produce ZERO blast.high-coupling violations.
	// Asserting only "no Error" lets the IQR threshold silently break: a
	// mutated coefficient (e.g. 1.5 → 0.0) would mark every package as an
	// outlier and emit Warnings against the clean fixture, but
	// EffectiveSeverity stays Warning so the looser assertion would miss it.
	for _, v := range violations {
		if v.Rule == "blast.high-coupling" {
			t.Fatalf("valid project should have no blast.high-coupling violations; got %s", v.String())
		}
		if v.EffectiveSeverity == core.Error {
			t.Fatalf("unexpected error violation: %s", v.String())
		}
	}
}

func TestBlastRadiusDetectsOutlier(t *testing.T) {
	ctx := loadContext(t, "../../testdata/blast", "github.com/kimtaeyun/testproject-blast", dddArchitecture(), "internal/...")

	violations := dependency.NewBlastRadius().Check(ctx)
	if len(violations) == 0 {
		t.Fatal("expected at least one blast.high-coupling violation")
	}

	foundPkg := false
	for _, v := range violations {
		if v.Rule != "blast.high-coupling" {
			t.Fatalf("unexpected rule: %s", v.Rule)
		}
		if v.EffectiveSeverity != core.Warning {
			t.Fatalf("expected Warning severity, got %v", v.EffectiveSeverity)
		}
		if strings.Contains(v.File, "internal/pkg") {
			foundPkg = true
		}
	}
	if !foundPkg {
		t.Fatalf("expected internal/pkg to be flagged, got %v", ruleIDs(violations))
	}
}

func TestBlastRadiusFlatLayoutEmitsMetaWarning(t *testing.T) {
	ctx := loadFlatLayoutContext(t)
	violations := dependency.NewBlastRadius().Check(ctx)
	assertExactlyOneMetaLayoutNotSupported(t, violations, "dependency.blast-radius")
}

func TestBlastRadiusExclude(t *testing.T) {
	ctx := loadContextWithExclude(t,
		"../../testdata/blast",
		"github.com/kimtaeyun/testproject-blast",
		dddArchitecture(),
		[]string{"internal/pkg/..."},
		"internal/...",
	)

	violations := dependency.NewBlastRadius().Check(ctx)
	for _, v := range violations {
		if strings.Contains(v.File, "internal/pkg") {
			t.Fatalf("excluded package should not appear in violations: %s", v.String())
		}
	}
}
