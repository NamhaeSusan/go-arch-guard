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
		Sublayers: []string{"core/repo", "core/svc", "core/model", "port"},
	}

	if !IsPortSublayer(layers, "core/repo") {
		t.Fatal("fallback port layer not detected")
	}
	if IsPortSublayer(layers, "port") {
		t.Fatal("fallback port layer must not treat a literal port/ sublayer as a port unless PortLayers opts in")
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

func TestSublayerMatchingRejectsPathSubstringsOutsideDomainLayer(t *testing.T) {
	layers := core.LayerModel{
		Sublayers: []string{"core/repo", "core/svc", "core/model"},
	}

	if got := MatchPortSublayer(layers, "example.com/app/internal/pkg/core/repo/cache"); got != "" {
		t.Fatalf("MatchPortSublayer() = %q, want no match outside domain layer", got)
	}
	if got := MatchPortSublayer(layers, "example.com/app/internal/pkg/core/repo"); got != "" {
		t.Fatalf("MatchPortSublayer() = %q, want no exact suffix match outside domain layer", got)
	}
	if got := MatchPortSublayer(layers, "example.com/app/internal/pkg/domain/order/core/repo"); got != "" {
		t.Fatalf("MatchPortSublayer() = %q, want no internal/pkg/domain false match", got)
	}
	if got := MatchPortSublayer(layers, "example.com/domain/app/internal/domain/order/core/repo"); got != "core/repo" {
		t.Fatalf("MatchPortSublayer() = %q, want core/repo with module prefix containing domain", got)
	}
	if got := MatchPortSublayer(layers, "example.com/app/internal/domain/order/core/repo/internal/domain/cache"); got != "core/repo" {
		t.Fatalf("MatchPortSublayer() = %q, want core/repo with nested internal/domain below layer", got)
	}
	if got := MatchContractSublayer(layers, "example.com/app/internal/pkg/core/svc/cache"); got != "" {
		t.Fatalf("MatchContractSublayer() = %q, want no match outside domain layer", got)
	}
}

func TestSublayerMatchingUsesConfiguredLayout(t *testing.T) {
	layers := core.LayerModel{
		Sublayers: []string{"core/repo", "core/svc", "core/model"},
	}

	layout := core.LayoutModel{InternalRoot: "packages", DomainDir: "domain"}
	if got := MatchPortSublayerInLayout(layers, layout, "example.com/app/packages/domain/order/core/repo"); got != "core/repo" {
		t.Fatalf("MatchPortSublayerInLayout() = %q, want core/repo for custom InternalRoot", got)
	}
	if got := MatchPortSublayerInLayout(layers, layout, "example.com/app/internal/domain/order/core/repo"); got != "" {
		t.Fatalf("MatchPortSublayerInLayout() = %q, want no match for wrong InternalRoot", got)
	}

	layout = core.LayoutModel{InternalRoot: "internal", DomainDir: "module"}
	if got := MatchContractSublayerInLayout(layers, layout, "example.com/app/internal/module/order/core/svc"); got != "core/svc" {
		t.Fatalf("MatchContractSublayerInLayout() = %q, want core/svc for custom DomainDir", got)
	}
}

func TestExplicitPortSublayerMatching(t *testing.T) {
	layers := core.LayerModel{
		Sublayers:      []string{"handler", "usecase", "port", "domain", "adapter"},
		PortLayers:     []string{"port"},
		ContractLayers: []string{"port"},
	}

	if !IsPortSublayer(layers, "port") {
		t.Fatal("explicit PortLayers entry should mark port as a port sublayer")
	}
	if got := MatchPortSublayer(layers, "example.com/app/internal/domain/order/port"); got != "port" {
		t.Fatalf("MatchPortSublayer() = %q, want port", got)
	}
	if got := PortSublayerName(layers); got != "port" {
		t.Fatalf("PortSublayerName() = %q, want port", got)
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
