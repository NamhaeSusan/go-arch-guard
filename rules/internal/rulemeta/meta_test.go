package rulemeta_test

import (
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/rules/internal/rulemeta"
)

func TestRuleDisabledByConfig(t *testing.T) {
	got := rulemeta.RuleDisabledByConfig("rules.example", "missing config", "set config")
	if got.Rule != "meta.rule-disabled-by-config" {
		t.Fatalf("Rule = %q", got.Rule)
	}
	if got.Message != "rules.example: missing config" {
		t.Fatalf("Message = %q", got.Message)
	}
	if got.Fix != "set config" {
		t.Fatalf("Fix = %q", got.Fix)
	}
	if got.DefaultSeverity != core.Warning || got.EffectiveSeverity != core.Warning {
		t.Fatalf("severity = %s/%s, want warning/warning", got.DefaultSeverity, got.EffectiveSeverity)
	}
}
