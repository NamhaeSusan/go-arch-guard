package analysisutil

import (
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/core"
)

func TestPortAndContractSublayers(t *testing.T) {
	layers := core.LayerModel{
		Sublayers:      []string{"store", "svc", "model"},
		PortLayers:     []string{"store"},
		ContractLayers: []string{"store", "svc"},
	}

	if !IsPortSublayer(layers, "store") {
		t.Fatal("IsPortSublayer(store) = false")
	}
	if IsPortSublayer(layers, "model") {
		t.Fatal("IsPortSublayer(model) = true")
	}
	if !IsContractSublayer(layers, "svc") {
		t.Fatal("IsContractSublayer(svc) = false")
	}
	if IsContractSublayer(layers, "model") {
		t.Fatal("IsContractSublayer(model) = true")
	}
}

func TestFallbackSublayerMatching(t *testing.T) {
	layers := core.LayerModel{
		Sublayers: []string{"core/repo", "core/svc", "core/model"},
	}

	if !IsPortSublayer(layers, "core/repo") {
		t.Fatal("fallback port layer not detected")
	}
	if !IsContractSublayer(layers, "core/svc") {
		t.Fatal("fallback contract layer not detected")
	}
	if got := MatchPortSublayer(layers, "example.com/app/internal/domain/order/core/repo"); got != "core/repo" {
		t.Fatalf("MatchPortSublayer() = %q", got)
	}
	if got := MatchContractSublayer(layers, "example.com/app/internal/domain/order/core/svc"); got != "core/svc" {
		t.Fatalf("MatchContractSublayer() = %q", got)
	}
	if got := PortSublayerName(layers); got != "core/repo" {
		t.Fatalf("PortSublayerName() = %q", got)
	}
}

func TestHasPortAndKnownSublayer(t *testing.T) {
	layers := core.LayerModel{
		Sublayers:      []string{"handler", "store"},
		PortLayers:     []string{"store"},
		ContractLayers: []string{"store"},
	}

	if !HasPortSublayer(layers) {
		t.Fatal("HasPortSublayer() = false")
	}
	if !IsKnownSublayer(layers, "handler") {
		t.Fatal("IsKnownSublayer(handler) = false")
	}
	if IsKnownSublayer(layers, "missing") {
		t.Fatal("IsKnownSublayer(missing) = true")
	}
}
