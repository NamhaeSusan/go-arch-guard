package core

import "testing"

func TestRuleSpecViolationIDs(t *testing.T) {
	spec := RuleSpec{
		ID:              "dependency.isolation",
		Description:     "domain boundaries",
		DefaultSeverity: Error,
		Violations: []ViolationSpec{
			{ID: "isolation.cross-domain", Description: "no sibling import", DefaultSeverity: Error},
			{ID: "isolation.pkg-imports-domain", Description: "pkg cannot import domain", DefaultSeverity: Error},
		},
	}
	got := spec.ViolationIDs()
	want := []string{"isolation.cross-domain", "isolation.pkg-imports-domain"}
	if len(got) != len(want) {
		t.Fatalf("ViolationIDs() len = %d, want %d", len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("ViolationIDs()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestRuleSpecViolationIDsEmpty(t *testing.T) {
	spec := RuleSpec{ID: "naming.no-stutter"}
	if got := spec.ViolationIDs(); len(got) != 0 {
		t.Errorf("ViolationIDs() = %v, want empty", got)
	}
}
