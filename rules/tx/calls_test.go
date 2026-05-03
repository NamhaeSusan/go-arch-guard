package tx_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/analyzer"
	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/rules/tx"
)

func TestForbiddenCallsDetectsWrongLayer(t *testing.T) {
	rule := tx.NewForbiddenCalls([]tx.ForbiddenCall{{
		Symbols:       []string{"database/sql.(*DB).BeginTx"},
		AllowedLayers: []string{"app"},
	}})

	got := rule.Check(txBoundaryContext(t, nil))

	if countRule(got, "tx.forbidden-call") != 1 {
		t.Fatalf("forbidden call count = %d, want 1: %+v", countRule(got, "tx.forbidden-call"), got)
	}
	if got[0].File != "internal/domain/order/core/repo/repository.go" {
		t.Fatalf("violation file = %q, want repo", got[0].File)
	}
}

func TestForbiddenCallsRespectsExcludesAndSeverity(t *testing.T) {
	rule := tx.NewForbiddenCalls([]tx.ForbiddenCall{{
		Symbols:       []string{"database/sql.(*DB).BeginTx"},
		AllowedLayers: []string{"app"},
	}}, tx.WithCallSeverity(core.Warning))

	ctx := txBoundaryContext(t, []string{"internal/domain/order/core/repo/..."})
	got := core.Run(ctx, core.NewRuleSet(rule))
	if len(got) != 0 {
		t.Fatalf("excluded repo should suppress forbidden calls, got %+v", got)
	}

	got = core.Run(txBoundaryContext(t, nil), core.NewRuleSet(rule))
	if len(got) == 0 {
		t.Fatal("expected severity-covered violation")
	}
	if got[0].DefaultSeverity != core.Warning || got[0].EffectiveSeverity != core.Warning {
		t.Fatalf("severity = %s/%s, want warning/warning", got[0].DefaultSeverity, got[0].EffectiveSeverity)
	}
}

func TestMandatoryWrapperDetectsDirectCallAndMentionsReplacement(t *testing.T) {
	rule := tx.NewMandatoryWrapper([]tx.MandatoryWrapper{{
		Symbols:       []string{"database/sql.(*DB).BeginTx"},
		AllowedLayers: []string{"app"},
		ReplaceWith:   "internal/pkg/tx.BeginTx",
	}})

	got := rule.Check(txBoundaryContext(t, nil))

	if countRule(got, "tx.mandatory-wrapper") != 1 {
		t.Fatalf("mandatory wrapper count = %d, want 1: %+v", countRule(got, "tx.mandatory-wrapper"), got)
	}
	if !strings.Contains(got[0].Fix, "internal/pkg/tx.BeginTx") {
		t.Fatalf("fix should mention replacement, got %q", got[0].Fix)
	}
}

func TestCallRulesAllowProjectRelativePackagePrefixesUnderInternalRoot(t *testing.T) {
	rule := tx.NewMandatoryWrapper([]tx.MandatoryWrapper{{
		Symbols:       []string{"database/sql.(*DB).BeginTx"},
		AllowedLayers: []string{"pkg/httpclient"},
		ReplaceWith:   "pkg/httpclient.BeginTx",
	}})

	got := rule.Check(callRuleSharedContext(t))

	if countRule(got, "tx.mandatory-wrapper") != 1 {
		t.Fatalf("mandatory wrapper count = %d, want only sibling shared package violation: %+v", countRule(got, "tx.mandatory-wrapper"), got)
	}
	if !strings.Contains(got[0].File, "internal/pkg/bad/bad.go") {
		t.Fatalf("violation file = %q, want sibling shared package", got[0].File)
	}
}

func TestCallRulesEmptyConfigEmitsMeta(t *testing.T) {
	for _, rule := range []core.Rule{
		tx.NewForbiddenCalls(nil),
		tx.NewMandatoryWrapper(nil),
	} {
		got := rule.Check(txBoundaryContext(t, nil))
		if len(got) != 1 || got[0].Rule != "meta.rule-disabled-by-config" {
			t.Fatalf("%s empty config: got %+v", rule.Spec().ID, got)
		}
	}
}

func countRule(violations []core.Violation, rule string) int {
	var count int
	for _, v := range violations {
		if v.Rule == rule {
			count++
		}
	}
	return count
}

func callRuleSharedContext(t *testing.T) *core.Context {
	t.Helper()
	root := t.TempDir()
	writeCallRuleFile(t, filepath.Join(root, "go.mod"), "module example.com/callrules\n\ngo 1.25.0\n")
	source := `package PLACEHOLDER

import (
	"context"
	"database/sql"
)

func Begin(ctx context.Context, db *sql.DB) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	return tx.Commit()
}
`
	writeCallRuleFile(t, filepath.Join(root, "internal", "pkg", "httpclient", "httpclient.go"), strings.ReplaceAll(source, "PLACEHOLDER", "httpclient"))
	writeCallRuleFile(t, filepath.Join(root, "internal", "pkg", "bad", "bad.go"), strings.ReplaceAll(source, "PLACEHOLDER", "bad"))

	pkgs, err := analyzer.Load(root, "internal/...")
	if err != nil {
		t.Fatal(err)
	}
	arch := txBoundaryArch()
	return core.NewContext(pkgs, "example.com/callrules", root, arch, nil)
}

func writeCallRuleFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
