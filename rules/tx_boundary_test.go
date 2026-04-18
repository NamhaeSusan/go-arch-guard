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

	var repoStartHit bool
	var leakHits int
	for _, v := range got {
		switch v.Rule {
		case "tx.start-outside-allowed-layer":
			if strings.Contains(v.File, "core/repo/repository.go") {
				repoStartHit = true
			}
		case "tx.type-in-signature":
			leakHits++
		default:
			t.Errorf("unexpected rule: %s", v.Rule)
		}
	}
	if !repoStartHit {
		t.Error("expected tx.start-outside-allowed-layer violation in core/repo/repository.go")
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
		rules.WithExclude(
			"internal/domain/order/core/repo/...",
			"internal/domain/order/core/svc/...",
			"internal/generic/...",
			"internal/testutil/...",
		),
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

func TestCheckTxBoundary_AllowedLayersIncludeOffenders(t *testing.T) {
	pkgs := loadTxBoundary(t)
	// Exclude new fixture packages that intentionally produce violations for
	// other test cases. This test is scoped to the original txboundary scenario.
	got := rules.CheckTxBoundary(pkgs,
		"github.com/kimtaeyun/testproject-txboundary",
		"../testdata/txboundary",
		rules.WithTxBoundary(rules.TxBoundaryConfig{
			StartSymbols:  []string{"database/sql.(*DB).BeginTx"},
			Types:         []string{"database/sql.Tx"},
			AllowedLayers: []string{"app", "core/repo", "core/svc"},
		}),
		rules.WithExclude("internal/generic/...", "internal/testutil/..."),
	)
	if len(got) != 0 {
		t.Fatalf("expected 0 violations when all layers allowed, got %d", len(got))
	}
}

func TestCheckTxBoundary_UnknownSymbolsNoStartViolations(t *testing.T) {
	pkgs := loadTxBoundary(t)
	got := rules.CheckTxBoundary(pkgs,
		"github.com/kimtaeyun/testproject-txboundary",
		"../testdata/txboundary",
		rules.WithTxBoundary(rules.TxBoundaryConfig{
			StartSymbols:  []string{"some/unknown/pkg.Begin"},
			Types:         []string{"database/sql.Tx"},
			AllowedLayers: []string{"app"},
		}),
	)
	for _, v := range got {
		if v.Rule == "tx.start-outside-allowed-layer" {
			t.Errorf("unexpected start violation for unknown symbol: %+v", v)
		}
	}
}

func TestCheckTxBoundary_StripsNonPointerWrappers(t *testing.T) {
	pkgs := loadTxBoundary(t)
	got := rules.CheckTxBoundary(pkgs,
		"github.com/kimtaeyun/testproject-txboundary",
		"../testdata/txboundary",
		rules.WithTxBoundary(rules.TxBoundaryConfig{
			Types:         []string{"database/sql.Tx"},
			AllowedLayers: []string{"app"},
		}),
	)
	var sliceHit bool
	for _, v := range got {
		if v.Rule != "tx.type-in-signature" {
			continue
		}
		if strings.Contains(v.File, "core/repo/repository.go") && strings.Contains(v.Message, "database/sql.Tx") {
			// BatchSave([]*sql.Tx, ...) should be flagged via slice-wrapper stripping.
			sliceHit = true
		}
	}
	if !sliceHit {
		t.Error("expected []*sql.Tx in BatchSave to be flagged via slice wrapper stripping")
	}
}

func TestCheckTxBoundary_OnlyStartSymbolsConfigured(t *testing.T) {
	pkgs := loadTxBoundary(t)
	// Exclude new fixture packages scoped to other test cases.
	got := rules.CheckTxBoundary(pkgs,
		"github.com/kimtaeyun/testproject-txboundary",
		"../testdata/txboundary",
		rules.WithTxBoundary(rules.TxBoundaryConfig{
			StartSymbols:  []string{"database/sql.(*DB).BeginTx"},
			AllowedLayers: []string{"app"},
		}),
		rules.WithExclude("internal/generic/...", "internal/testutil/..."),
	)
	for _, v := range got {
		if v.Rule == "tx.type-in-signature" {
			t.Errorf("unexpected type violation when Types unset: %+v", v)
		}
	}
	var starts int
	for _, v := range got {
		if v.Rule == "tx.start-outside-allowed-layer" {
			starts++
		}
	}
	if starts != 1 {
		t.Errorf("expected 1 start violation, got %d", starts)
	}
}

// TestCheckTxBoundary_UnclassifiedInternalPackageFlagged verifies that an
// internal package that does not belong to any known sublayer (e.g.
// internal/testutil) is treated as non-allowed and therefore produces a
// violation when it calls a forbidden symbol.
func TestCheckTxBoundary_UnclassifiedInternalPackageFlagged(t *testing.T) {
	pkgs := loadTxBoundary(t)
	got := rules.CheckTxBoundary(pkgs,
		"github.com/kimtaeyun/testproject-txboundary",
		"../testdata/txboundary",
		rules.WithTxBoundary(rules.TxBoundaryConfig{
			StartSymbols:  []string{"database/sql.(*DB).BeginTx"},
			AllowedLayers: []string{"app"},
		}),
	)
	var hit bool
	for _, v := range got {
		if v.Rule == "tx.start-outside-allowed-layer" && strings.Contains(v.File, "testutil") {
			hit = true
		}
	}
	if !hit {
		t.Error("expected tx.start-outside-allowed-layer violation from unclassified internal/testutil package")
	}
}

// TestCheckTxBoundary_GenericCallFlagged verifies that a forbidden symbol
// called via generic instantiation syntax (Fun is *ast.IndexExpr) is not
// silently skipped by resolveCalleeID.
func TestCheckTxBoundary_GenericCallFlagged(t *testing.T) {
	pkgs := loadTxBoundary(t)
	got := rules.CheckTxBoundary(pkgs,
		"github.com/kimtaeyun/testproject-txboundary",
		"../testdata/txboundary",
		rules.WithTxBoundary(rules.TxBoundaryConfig{
			StartSymbols:  []string{"database/sql.(*DB).BeginTx"},
			AllowedLayers: []string{"app"},
		}),
	)
	var hit bool
	for _, v := range got {
		if v.Rule == "tx.start-outside-allowed-layer" && strings.Contains(v.File, "generic") {
			hit = true
		}
	}
	if !hit {
		t.Error("expected tx.start-outside-allowed-layer violation from internal/generic package (generic call site)")
	}
}
