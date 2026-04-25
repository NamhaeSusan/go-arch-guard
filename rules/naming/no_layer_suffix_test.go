package naming_test

import (
	"strings"
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/rules/naming"
)

func TestNoLayerSuffixSpec(t *testing.T) {
	spec := naming.NewNoLayerSuffix(naming.WithSeverity(core.Error)).Spec()

	if spec.ID != "naming.no-layer-suffix" {
		t.Fatalf("ID = %q, want naming.no-layer-suffix", spec.ID)
	}
	if spec.DefaultSeverity != core.Error {
		t.Fatalf("DefaultSeverity = %v, want Error", spec.DefaultSeverity)
	}
}

func TestNoLayerSuffixUsesLayerDirNames(t *testing.T) {
	ctx := tempContext(t, map[string]string{
		"internal/domain/order/usecase/order_usecase.go": "package usecase\n\ntype Order struct{}\n",
	}, core.Architecture{
		Layers: core.LayerModel{
			Sublayers:     []string{"usecase"},
			LayerDirNames: map[string]bool{"usecase": true},
		},
	})

	got := naming.NewNoLayerSuffix().Check(ctx)
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1: %+v", len(got), got)
	}
	if got[0].Rule != "naming.no-layer-suffix" || !strings.Contains(got[0].Fix, "order.go") {
		t.Fatalf("violation = %+v, want no-layer-suffix rename to order.go", got[0])
	}
}

func TestNoLayerSuffixSkipsNonLayerDirs(t *testing.T) {
	ctx := tempContext(t, map[string]string{
		"internal/domain/order/usecase/order_usecase.go": "package usecase\n\ntype Order struct{}\n",
	}, core.Architecture{Layers: core.LayerModel{Sublayers: []string{"usecase"}}})

	if got := naming.NewNoLayerSuffix().Check(ctx); len(got) != 0 {
		t.Fatalf("without LayerDirNames, got violations: %+v", got)
	}
}
