package dependency_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/rules/dependency"
)

func TestLayerDirectionSpec(t *testing.T) {
	rule := dependency.NewLayerDirection(dependency.WithSeverity(core.Warning))

	spec := rule.Spec()
	if spec.ID != "dependency.layer-direction" {
		t.Fatalf("Spec().ID = %q, want dependency.layer-direction", spec.ID)
	}
	if spec.DefaultSeverity != core.Warning {
		t.Fatalf("Spec().DefaultSeverity = %v, want %v", spec.DefaultSeverity, core.Warning)
	}
	for _, id := range []string{"layer.direction", "layer.inner-imports-pkg", "layer.unknown-sublayer"} {
		if !slices.Contains(spec.ViolationIDs(), id) {
			t.Fatalf("Spec().ViolationIDs() missing %q", id)
		}
	}
}

func TestLayerDirectionValidProject(t *testing.T) {
	ctx := loadContext(t, "../../testdata/valid", "github.com/kimtaeyun/testproject-dc", dddArchitecture(), "internal/...")

	violations := dependency.NewLayerDirection().Check(ctx)
	if len(violations) > 0 {
		for _, v := range violations {
			t.Log(v.String())
		}
		t.Fatalf("expected no violations, got %d", len(violations))
	}
}

func TestLayerDirectionInvalidProject(t *testing.T) {
	ctx := loadContext(t, "../../testdata/invalid", "github.com/kimtaeyun/testproject-dc-invalid", dddArchitecture(), "internal/...")

	violations := dependency.NewLayerDirection().Check(ctx)

	assertHasRule(t, violations, "layer.direction")
	assertHasRule(t, violations, "layer.inner-imports-pkg")
	assertHasRule(t, violations, "layer.unknown-sublayer")

	foundHandlerEvent := false
	for _, v := range violations {
		if v.Rule == "layer.direction" &&
			strings.Contains(v.Message, `"handler"`) &&
			strings.Contains(v.Message, `"event"`) {
			foundHandlerEvent = true
			break
		}
	}
	if !foundHandlerEvent {
		t.Fatalf("expected handler -> event layer.direction violation, got %v", ruleIDs(violations))
	}
}

func TestLayerDirectionExclude(t *testing.T) {
	ctx := loadContextWithExclude(t,
		"../../testdata/invalid",
		"github.com/kimtaeyun/testproject-dc-invalid",
		dddArchitecture(),
		[]string{"internal/domain/payment/core/model/..."},
		"internal/...",
	)

	violations := dependency.NewLayerDirection().Check(ctx)
	for _, v := range violations {
		if v.File == "internal/domain/payment/core/model/pkg_leak.go" {
			t.Fatalf("expected payment model package to be excluded, got %s", v.String())
		}
	}
}
