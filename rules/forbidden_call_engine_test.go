package rules

import (
	"strings"
	"sync"
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/analyzer"
	"golang.org/x/tools/go/packages"
)

var (
	txbInternalOnce sync.Once
	txbInternalPkgs []*packages.Package
	txbInternalErr  error
)

func loadTxBoundaryInternal(t *testing.T) []*packages.Package {
	t.Helper()
	txbInternalOnce.Do(func() {
		txbInternalPkgs, txbInternalErr = analyzer.Load("../testdata/txboundary", "internal/...")
	})
	if txbInternalErr != nil {
		t.Fatal(txbInternalErr)
	}
	return txbInternalPkgs
}

func TestForbiddenCallEngine_SymbolMatch_OutsideAllowedLayer(t *testing.T) {
	pkgs := loadTxBoundaryInternal(t)
	rules := []forbiddenCallRule{{
		Symbols:       []string{"database/sql.(*DB).BeginTx"},
		AllowedLayers: []string{"app"},
		RuleName:      "tx.start-outside-allowed-layer",
		Message:       "transaction must start in allowed layers: %v",
		Fix:           "move tx start into one of: %v",
	}}
	cfg := NewConfig()
	m := DDD()

	got := checkForbiddenCallsByLayer(pkgs,
		"github.com/kimtaeyun/testproject-txboundary",
		"../testdata/txboundary",
		m, cfg, rules)

	if len(got) != 1 {
		t.Fatalf("expected 1 violation, got %d: %+v", len(got), got)
	}
	v := got[0]
	if v.Rule != "tx.start-outside-allowed-layer" {
		t.Errorf("unexpected rule: %s", v.Rule)
	}
	if !strings.Contains(v.File, "core/repo/repository.go") {
		t.Errorf("unexpected file: %s", v.File)
	}
}

func TestForbiddenCallEngine_SymbolMatch_InsideAllowedLayer(t *testing.T) {
	pkgs := loadTxBoundaryInternal(t)
	rules := []forbiddenCallRule{{
		Symbols:       []string{"database/sql.(*DB).BeginTx"},
		AllowedLayers: []string{"app", "core/repo"},
		RuleName:      "tx.start-outside-allowed-layer",
		Message:       "tx must start in: %v",
		Fix:           "move to: %v",
	}}
	got := checkForbiddenCallsByLayer(pkgs,
		"github.com/kimtaeyun/testproject-txboundary",
		"../testdata/txboundary",
		DDD(), NewConfig(), rules)
	if len(got) != 0 {
		t.Fatalf("expected 0 violations, got %d: %+v", len(got), got)
	}
}

func TestForbiddenCallEngine_NoSymbolMatch(t *testing.T) {
	pkgs := loadTxBoundaryInternal(t)
	rules := []forbiddenCallRule{{
		Symbols:       []string{"some/unknown/pkg.Nothing"},
		AllowedLayers: []string{"app"},
		RuleName:      "x.test",
		Message:       "%v",
		Fix:           "%v",
	}}
	got := checkForbiddenCallsByLayer(pkgs,
		"github.com/kimtaeyun/testproject-txboundary",
		"../testdata/txboundary",
		DDD(), NewConfig(), rules)
	if len(got) != 0 {
		t.Fatalf("expected 0 violations, got %d", len(got))
	}
}

func TestForbiddenCallEngine_RespectsExclude(t *testing.T) {
	pkgs := loadTxBoundaryInternal(t)
	rules := []forbiddenCallRule{{
		Symbols:       []string{"database/sql.(*DB).BeginTx"},
		AllowedLayers: []string{"app"},
		RuleName:      "tx.start-outside-allowed-layer",
		Message:       "%v",
		Fix:           "%v",
	}}
	cfg := NewConfig(WithExclude("internal/domain/order/core/repo/..."))
	got := checkForbiddenCallsByLayer(pkgs,
		"github.com/kimtaeyun/testproject-txboundary",
		"../testdata/txboundary",
		DDD(), cfg, rules)
	if len(got) != 0 {
		t.Fatalf("expected 0 violations after exclude, got %d", len(got))
	}
}

func TestForbiddenCallEngine_RespectsSeverity(t *testing.T) {
	pkgs := loadTxBoundaryInternal(t)
	rules := []forbiddenCallRule{{
		Symbols:       []string{"database/sql.(*DB).BeginTx"},
		AllowedLayers: []string{"app"},
		RuleName:      "tx.start-outside-allowed-layer",
		Message:       "%v",
		Fix:           "%v",
	}}
	cfg := NewConfig(WithSeverity(Warning))
	got := checkForbiddenCallsByLayer(pkgs,
		"github.com/kimtaeyun/testproject-txboundary",
		"../testdata/txboundary",
		DDD(), cfg, rules)
	if len(got) == 0 {
		t.Fatal("expected at least one violation")
	}
	for _, v := range got {
		if v.Severity != Warning {
			t.Errorf("expected Warning severity, got %v", v.Severity)
		}
	}
}
