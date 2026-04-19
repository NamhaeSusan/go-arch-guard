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
	// cmd/ is skipped by default (EnforceCmdRoot=false), so no cmd exclude needed.
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
	// cmd/ is skipped by default (EnforceCmdRoot=false), so internal-only
	// AllowedLayers is enough to produce zero violations.
	got := rules.CheckTxBoundary(pkgs,
		"github.com/kimtaeyun/testproject-txboundary",
		"../testdata/txboundary",
		rules.WithTxBoundary(rules.TxBoundaryConfig{
			StartSymbols:  []string{"database/sql.(*DB).BeginTx"},
			Types:         []string{"database/sql.Tx"},
			AllowedLayers: []string{"app", "core/repo", "core/svc"},
		}),
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
	// Default behavior skips unclassified internal packages and cmd/, so
	// this test is naturally scoped to the core/repo case.
	got := rules.CheckTxBoundary(pkgs,
		"github.com/kimtaeyun/testproject-txboundary",
		"../testdata/txboundary",
		rules.WithTxBoundary(rules.TxBoundaryConfig{
			StartSymbols:  []string{"database/sql.(*DB).BeginTx"},
			AllowedLayers: []string{"app"},
		}),
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

// TestCheckTxBoundary_UnclassifiedInternalSkippedByDefault verifies the
// default behavior: internal packages that don't map to any known sublayer
// (e.g. internal/testutil) are silently skipped so ad-hoc helper packages
// don't produce noise.
func TestCheckTxBoundary_UnclassifiedInternalSkippedByDefault(t *testing.T) {
	pkgs := loadTxBoundary(t)
	got := rules.CheckTxBoundary(pkgs,
		"github.com/kimtaeyun/testproject-txboundary",
		"../testdata/txboundary",
		rules.WithTxBoundary(rules.TxBoundaryConfig{
			StartSymbols:  []string{"database/sql.(*DB).BeginTx"},
			AllowedLayers: []string{"app"},
		}),
	)
	for _, v := range got {
		if strings.Contains(v.File, "internal/testutil") {
			t.Errorf("unexpected violation from unclassified internal/testutil by default: %+v", v)
		}
	}
}

// TestCheckTxBoundary_UnclassifiedInternalFlaggedWhenEnforced verifies that
// turning on EnforceUnclassified catches forbidden calls in internal packages
// that don't map to any known sublayer.
func TestCheckTxBoundary_UnclassifiedInternalFlaggedWhenEnforced(t *testing.T) {
	pkgs := loadTxBoundary(t)
	got := rules.CheckTxBoundary(pkgs,
		"github.com/kimtaeyun/testproject-txboundary",
		"../testdata/txboundary",
		rules.WithTxBoundary(rules.TxBoundaryConfig{
			StartSymbols:        []string{"database/sql.(*DB).BeginTx"},
			AllowedLayers:       []string{"app"},
			EnforceUnclassified: true,
		}),
	)
	var hit bool
	for _, v := range got {
		if v.Rule == "tx.start-outside-allowed-layer" && strings.Contains(v.File, "internal/testutil") {
			hit = true
		}
	}
	if !hit {
		t.Error("expected tx.start-outside-allowed-layer violation from internal/testutil when EnforceUnclassified is true")
	}
}

// TestCheckTxBoundary_DefaultIgnoresCmdRoot verifies the backward-compat
// default: composition-root packages under <module>/cmd/... are skipped
// when EnforceCmdRoot is left at its zero value (false).
func TestCheckTxBoundary_DefaultIgnoresCmdRoot(t *testing.T) {
	pkgs := loadTxBoundary(t)
	got := rules.CheckTxBoundary(pkgs,
		"github.com/kimtaeyun/testproject-txboundary",
		"../testdata/txboundary",
		rules.WithTxBoundary(rules.TxBoundaryConfig{
			StartSymbols:  []string{"database/sql.(*DB).BeginTx"},
			AllowedLayers: []string{"app"},
			// EnforceCmdRoot left false — cmd/ must be skipped.
		}),
	)
	for _, v := range got {
		if strings.Contains(v.File, "cmd/") {
			t.Errorf("unexpected violation in cmd/ when EnforceCmdRoot is false: %+v", v)
		}
	}
}

// TestCheckTxBoundary_EnforceCmdRoot_FlagsCmdCalls verifies that turning on
// EnforceCmdRoot makes tx starts under <module>/cmd/... produce violations.
func TestCheckTxBoundary_EnforceCmdRoot_FlagsCmdCalls(t *testing.T) {
	pkgs := loadTxBoundary(t)
	got := rules.CheckTxBoundary(pkgs,
		"github.com/kimtaeyun/testproject-txboundary",
		"../testdata/txboundary",
		rules.WithTxBoundary(rules.TxBoundaryConfig{
			StartSymbols:   []string{"database/sql.(*DB).BeginTx"},
			AllowedLayers:  []string{"app"},
			EnforceCmdRoot: true,
		}),
	)
	var hit bool
	for _, v := range got {
		if v.Rule == "tx.start-outside-allowed-layer" && strings.Contains(v.File, "cmd/app/main.go") {
			hit = true
		}
	}
	if !hit {
		t.Error("expected tx.start-outside-allowed-layer from cmd/app/main.go when EnforceCmdRoot is true")
	}
}

// TestCheckTxBoundary_CmdRootIndependentOfUserSublayerNamedCmd is the
// regression test for the collision bug: a custom model with a legitimate
// sublayer literally named "cmd" must not accidentally exempt the
// composition-root packages under <module>/cmd/.... The composition-root
// policy is controlled exclusively by EnforceCmdRoot.
func TestCheckTxBoundary_CmdRootIndependentOfUserSublayerNamedCmd(t *testing.T) {
	pkgs := loadTxBoundary(t)
	// Custom flat-layout model where "cmd" is a real internal sublayer.
	// AllowedLayers lists "cmd" — in the old design this could have
	// accidentally exempted <module>/cmd/... by matching a synthetic token.
	customModel := rules.NewModel(
		rules.WithDomainDir(""),
		rules.WithSublayers([]string{"app", "cmd", "core/repo", "core/svc"}),
	)
	got := rules.CheckTxBoundary(pkgs,
		"github.com/kimtaeyun/testproject-txboundary",
		"../testdata/txboundary",
		rules.WithModel(customModel),
		rules.WithTxBoundary(rules.TxBoundaryConfig{
			StartSymbols:   []string{"database/sql.(*DB).BeginTx"},
			AllowedLayers:  []string{"app", "cmd"},
			EnforceCmdRoot: true, // strict mode — composition root must be flagged
		}),
	)
	var hit bool
	for _, v := range got {
		if v.Rule == "tx.start-outside-allowed-layer" && strings.Contains(v.File, "cmd/app/main.go") {
			hit = true
		}
	}
	if !hit {
		t.Errorf(
			"expected tx.start-outside-allowed-layer in cmd/app/main.go even when 'cmd' is a real user sublayer in AllowedLayers; got %+v",
			got,
		)
	}
}

// TestCheckTxBoundary_GenericCallFlagged verifies that a forbidden symbol
// called via generic instantiation syntax (Fun is *ast.IndexExpr) is not
// silently skipped by resolveCalleeID. The forbidden symbol here is a
// generic function in core/repo, invoked as BeginGeneric[string](...).
func TestCheckTxBoundary_GenericCallFlagged(t *testing.T) {
	pkgs := loadTxBoundary(t)
	const genericSymbol = "github.com/kimtaeyun/testproject-txboundary/internal/domain/order/core/repo.BeginGeneric"
	got := rules.CheckTxBoundary(pkgs,
		"github.com/kimtaeyun/testproject-txboundary",
		"../testdata/txboundary",
		rules.WithTxBoundary(rules.TxBoundaryConfig{
			StartSymbols:  []string{genericSymbol},
			AllowedLayers: []string{"app"},
		}),
	)
	var hit bool
	for _, v := range got {
		if v.Rule == "tx.start-outside-allowed-layer" && strings.Contains(v.File, "core/repo/generic.go") {
			hit = true
		}
	}
	if !hit {
		t.Errorf("expected violation for BeginGeneric[string](...) call site in core/repo/generic.go; got %+v", got)
	}
}
