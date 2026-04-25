package naming_test

import (
	"strings"
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/rules/naming"
)

func TestImplSuffixSpec(t *testing.T) {
	spec := naming.NewImplSuffix(naming.WithSeverity(core.Error)).Spec()

	if spec.ID != "naming.no-impl-suffix" {
		t.Fatalf("ID = %q, want naming.no-impl-suffix", spec.ID)
	}
	if spec.DefaultSeverity != core.Error {
		t.Fatalf("DefaultSeverity = %v, want Error", spec.DefaultSeverity)
	}
}

func TestImplSuffixFlagsExportedImplTypes(t *testing.T) {
	ctx := tempContext(t, map[string]string{
		"internal/domain/order/app/service.go": "package app\n\ntype ServiceImpl struct{}\ntype localImpl struct{}\n",
	}, dddArch())

	got := naming.NewImplSuffix().Check(ctx)
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1: %+v", len(got), got)
	}
	if got[0].Rule != "naming.no-impl-suffix" || !strings.Contains(got[0].Message, "ServiceImpl") {
		t.Fatalf("violation = %+v, want ServiceImpl impl-suffix", got[0])
	}
	if got[0].DefaultSeverity != core.Warning || got[0].EffectiveSeverity != core.Warning {
		t.Fatalf("severity = default %v effective %v, want Warning/Warning", got[0].DefaultSeverity, got[0].EffectiveSeverity)
	}
}
