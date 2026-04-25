package types_test

import (
	"strings"
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/core"
	types "github.com/NamhaeSusan/go-arch-guard/rules/types"
)

func typePatternArch() core.Architecture {
	return core.Architecture{
		Structure: core.StructurePolicy{
			TypePatterns: []core.TypePattern{
				{
					Dir:           "worker",
					FilePrefix:    "worker",
					TypeSuffix:    "Worker",
					RequireMethod: "Process",
				},
			},
		},
	}
}

func TestTypePatternSpec(t *testing.T) {
	spec := types.NewTypePattern(types.WithSeverity(core.Warning)).Spec()

	if spec.ID != "types.type-pattern" {
		t.Fatalf("ID = %q, want types.type-pattern", spec.ID)
	}
	if spec.DefaultSeverity != core.Warning {
		t.Fatalf("DefaultSeverity = %v, want Warning", spec.DefaultSeverity)
	}
	if len(spec.Violations) != 2 {
		t.Fatalf("expected two violation specs, got %+v", spec.Violations)
	}
	if spec.Violations[0].ID != "naming.type-pattern-mismatch" {
		t.Fatalf("first violation ID = %q", spec.Violations[0].ID)
	}
	if spec.Violations[1].ID != "naming.type-pattern-missing-method" {
		t.Fatalf("second violation ID = %q", spec.Violations[1].ID)
	}
}

func TestTypePatternFlagsMissingTypeAndMethod(t *testing.T) {
	ctx := newFixtureContext(t, typePatternArch(), nil)
	got := types.NewTypePattern().Check(ctx)

	var mismatch, missingMethod int
	for _, v := range got {
		switch v.Rule {
		case "naming.type-pattern-mismatch":
			mismatch++
			if !strings.Contains(v.File, "worker_payment.go") {
				t.Fatalf("mismatch file = %q, want worker_payment.go", v.File)
			}
		case "naming.type-pattern-missing-method":
			missingMethod++
			if !strings.Contains(v.File, "worker_invoice.go") {
				t.Fatalf("missing-method file = %q, want worker_invoice.go", v.File)
			}
		}
		if v.DefaultSeverity != core.Error || v.EffectiveSeverity != core.Error {
			t.Fatalf("type-pattern severity = default %v effective %v, want Error/Error", v.DefaultSeverity, v.EffectiveSeverity)
		}
	}

	if mismatch != 1 || missingMethod != 1 {
		t.Fatalf("mismatch=%d missingMethod=%d, want 1/1; all violations: %+v", mismatch, missingMethod, got)
	}
}

func TestTypePatternSkipsValidAndExcludedFiles(t *testing.T) {
	ctx := newFixtureContext(t, typePatternArch(), []string{"internal/worker/worker_payment.go"})
	got := types.NewTypePattern().Check(ctx)

	for _, v := range got {
		if strings.Contains(v.File, "worker_order.go") {
			t.Fatalf("valid worker should not be flagged: %+v", v)
		}
		if strings.Contains(v.File, "worker_payment.go") {
			t.Fatalf("excluded worker should not be flagged: %+v", v)
		}
	}
}
