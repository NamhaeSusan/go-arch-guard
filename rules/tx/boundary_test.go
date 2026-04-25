package tx_test

import (
	"strings"
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/analyzer"
	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/rules/tx"
	"golang.org/x/tools/go/packages"
)

func TestBoundarySpec(t *testing.T) {
	spec := tx.New(tx.Config{}).Spec()

	if spec.ID != "tx.boundary" {
		t.Fatalf("Spec().ID = %q, want tx.boundary", spec.ID)
	}
	if spec.DefaultSeverity != core.Error {
		t.Fatalf("Spec().DefaultSeverity = %v, want Error", spec.DefaultSeverity)
	}

	got := spec.ViolationIDs()
	want := []string{"tx.start-outside-allowed-layer", "tx.type-in-signature"}
	if len(got) != len(want) {
		t.Fatalf("ViolationIDs() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("ViolationIDs() = %v, want %v", got, want)
		}
	}
}

func TestBoundaryEmptyConfigNoop(t *testing.T) {
	rule := tx.New(tx.Config{})

	if got := rule.Check(txBoundaryContext(t, nil)); len(got) != 0 {
		t.Fatalf("Check() with empty config returned %d violations: %+v", len(got), got)
	}
}

func TestBoundaryDetectsStartAndSignatureOutsideDefaultAllowedLayer(t *testing.T) {
	rule := tx.New(tx.Config{
		StartSymbols: []string{"database/sql.(*DB).BeginTx"},
		Types:        []string{"database/sql.Tx"},
	})

	got := rule.Check(txBoundaryContext(t, nil))

	var startHits, signatureHits int
	for _, v := range got {
		switch v.Rule {
		case "tx.start-outside-allowed-layer":
			startHits++
			if !strings.Contains(v.File, "core/repo/repository.go") {
				t.Fatalf("start violation file = %q, want repo file", v.File)
			}
		case "tx.type-in-signature":
			signatureHits++
		default:
			t.Fatalf("unexpected violation ID %q", v.Rule)
		}
	}
	if startHits != 1 {
		t.Fatalf("start violation count = %d, want 1; violations: %+v", startHits, got)
	}
	if signatureHits < 3 {
		t.Fatalf("signature violation count = %d, want at least 3; violations: %+v", signatureHits, got)
	}
}

func TestBoundaryAllowedLayersCanIncludeOffenders(t *testing.T) {
	rule := tx.New(tx.Config{
		StartSymbols:  []string{"database/sql.(*DB).BeginTx"},
		Types:         []string{"database/sql.Tx"},
		AllowedLayers: []string{"app", "core/repo", "core/svc"},
	})

	if got := rule.Check(txBoundaryContext(t, nil)); len(got) != 0 {
		t.Fatalf("Check() returned %d violations, want 0: %+v", len(got), got)
	}
}

func TestBoundaryRespectsSeverityOptionThroughRunner(t *testing.T) {
	rule := tx.New(tx.Config{
		StartSymbols: []string{"database/sql.(*DB).BeginTx"},
		Types:        []string{"database/sql.Tx"},
	}, tx.WithSeverity(core.Warning))

	got := core.Run(txBoundaryContext(t, nil), core.NewRuleSet(rule))
	if len(got) == 0 {
		t.Fatal("expected violations")
	}
	for _, v := range got {
		if v.DefaultSeverity != core.Warning || v.EffectiveSeverity != core.Warning {
			t.Fatalf("severity = default %v effective %v, want Warning", v.DefaultSeverity, v.EffectiveSeverity)
		}
	}
}

func txBoundaryContext(t *testing.T, exclude []string) *core.Context {
	t.Helper()
	return core.NewContext(
		loadTxBoundary(t),
		"github.com/kimtaeyun/testproject-txboundary",
		"../../testdata/txboundary",
		txBoundaryArch(),
		exclude,
	)
}

func loadTxBoundary(t *testing.T) []*packages.Package {
	t.Helper()
	pkgs, err := analyzer.Load("../../testdata/txboundary", "internal/...")
	if err != nil {
		t.Fatal(err)
	}
	return pkgs
}

func txBoundaryArch() core.Architecture {
	return core.Architecture{
		Layers: core.LayerModel{
			Sublayers: []string{"handler", "app", "core", "core/model", "core/repo", "core/svc", "event", "infra"},
			Direction: map[string][]string{
				"handler":    {"app"},
				"app":        {"core/model", "core/repo", "core/svc", "event"},
				"core":       {"core/model"},
				"core/model": {},
				"core/repo":  {"core/model"},
				"core/svc":   {"core/model"},
				"event":      {"core/model"},
				"infra":      {"core/repo", "core/model", "event"},
			},
			PortLayers:     []string{"core/repo"},
			ContractLayers: []string{"core/repo", "core/svc"},
			PkgRestricted: map[string]bool{
				"core": true, "core/model": true, "core/repo": true, "core/svc": true, "event": true,
			},
			InternalTopLevel: map[string]bool{"domain": true, "orchestration": true, "pkg": true},
			LayerDirNames: map[string]bool{
				"handler": true, "app": true, "core": true, "model": true,
				"repo": true, "svc": true, "event": true, "infra": true,
			},
		},
		Layout: core.LayoutModel{
			DomainDir:        "domain",
			OrchestrationDir: "orchestration",
			SharedDir:        "pkg",
		},
		Structure: core.StructurePolicy{
			DTOAllowedLayers:        []string{"handler", "app"},
			InterfacePatternExclude: map[string]bool{"handler": true, "app": true, "core/model": true, "core/repo": true, "event": true},
		},
	}
}
