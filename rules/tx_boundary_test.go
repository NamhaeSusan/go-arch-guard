package rules_test

import (
	"strings"
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/rules"
)

func TestCheckTxBoundary_EmptyConfigNoop(t *testing.T) {
	pkgs := loadTxBoundary(t)
	got := rules.CheckTxBoundary(pkgs,
		"github.com/kimtaeyun/testproject-txboundary",
		"../testdata/txboundary",
	)
	if got != nil {
		t.Fatalf("expected nil from empty config, got %+v", got)
	}
}

func TestCheckTxBoundary_DetectsStartOutsideApp(t *testing.T) {
	pkgs := loadTxBoundary(t)
	got := rules.CheckTxBoundary(pkgs,
		"github.com/kimtaeyun/testproject-txboundary",
		"../testdata/txboundary",
		rules.WithTxBoundary(rules.TxBoundaryConfig{
			StartSymbols:  []string{"database/sql.(*DB).BeginTx"},
			Types:         []string{"database/sql.Tx"},
			AllowedLayers: []string{"app"},
		}),
	)

	var startHits, leakHits int
	for _, v := range got {
		switch v.Rule {
		case "tx.start-outside-allowed-layer":
			startHits++
			if !strings.Contains(v.File, "core/repo/repository.go") {
				t.Errorf("unexpected start-violation file: %s", v.File)
			}
		case "tx.type-in-signature":
			leakHits++
		default:
			t.Errorf("unexpected rule: %s", v.Rule)
		}
	}
	if startHits != 1 {
		t.Errorf("expected 1 start violation, got %d", startHits)
	}
	if leakHits < 2 {
		t.Errorf("expected >=2 signature violations (repo param + svc return), got %d", leakHits)
	}
}

func TestCheckTxBoundary_RespectsExclude(t *testing.T) {
	pkgs := loadTxBoundary(t)
	got := rules.CheckTxBoundary(pkgs,
		"github.com/kimtaeyun/testproject-txboundary",
		"../testdata/txboundary",
		rules.WithTxBoundary(rules.TxBoundaryConfig{
			StartSymbols:  []string{"database/sql.(*DB).BeginTx"},
			Types:         []string{"database/sql.Tx"},
			AllowedLayers: []string{"app"},
		}),
		rules.WithExclude("internal/domain/order/core/repo/...",
			"internal/domain/order/core/svc/..."),
	)
	if len(got) != 0 {
		t.Fatalf("expected 0 violations after exclude, got %d: %+v", len(got), got)
	}
}

func TestCheckTxBoundary_RespectsSeverity(t *testing.T) {
	pkgs := loadTxBoundary(t)
	got := rules.CheckTxBoundary(pkgs,
		"github.com/kimtaeyun/testproject-txboundary",
		"../testdata/txboundary",
		rules.WithTxBoundary(rules.TxBoundaryConfig{
			StartSymbols:  []string{"database/sql.(*DB).BeginTx"},
			Types:         []string{"database/sql.Tx"},
			AllowedLayers: []string{"app"},
		}),
		rules.WithSeverity(rules.Warning),
	)
	if len(got) == 0 {
		t.Fatal("expected at least one violation")
	}
	for _, v := range got {
		if v.Severity != rules.Warning {
			t.Errorf("expected Warning severity, got %v", v.Severity)
		}
	}
}
