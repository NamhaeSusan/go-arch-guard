package core

import "testing"

func TestNewRuleSetSeedsRules(t *testing.T) {
	r1 := &fakeRule{spec: RuleSpec{ID: "a"}}
	r2 := &fakeRule{spec: RuleSpec{ID: "b"}}
	rs := NewRuleSet(r1, r2)
	if got := len(rs.Rules()); got != 2 {
		t.Fatalf("len(Rules()) = %d, want 2", got)
	}
	if rs.Rules()[0].Spec().ID != "a" {
		t.Errorf("Rules()[0].ID = %q", rs.Rules()[0].Spec().ID)
	}
}

func TestRuleSetWithAppendsRules(t *testing.T) {
	r1 := &fakeRule{spec: RuleSpec{ID: "a"}}
	r2 := &fakeRule{spec: RuleSpec{ID: "b"}}
	rs := RuleSet{}.With(r1).With(r2)
	if got := len(rs.Rules()); got != 2 {
		t.Fatalf("len(Rules()) = %d, want 2", got)
	}
	if rs.Rules()[0].Spec().ID != "a" {
		t.Errorf("Rules()[0].ID = %q", rs.Rules()[0].Spec().ID)
	}
	if rs.Rules()[1].Spec().ID != "b" {
		t.Errorf("Rules()[1].ID = %q", rs.Rules()[1].Spec().ID)
	}
}

func TestRuleSetWithDropsNilRules(t *testing.T) {
	r := &fakeRule{spec: RuleSpec{ID: "a"}}
	rs := RuleSet{}.With(nil, r, nil)
	if got := len(rs.Rules()); got != 1 {
		t.Fatalf("len(Rules()) = %d, want 1 (nils must be dropped)", got)
	}
	if rs.Rules()[0].Spec().ID != "a" {
		t.Errorf("Rules()[0].ID = %q", rs.Rules()[0].Spec().ID)
	}
}

func TestRunDoesNotPanicOnNilOnlyRuleSet(t *testing.T) {
	rs := RuleSet{}.With(nil)
	ctx := NewContext(nil, "", "", validArchitecture(), nil)
	got := Run(ctx, rs)
	if len(got) != 0 {
		t.Fatalf("got %d violations from nil-only RuleSet, want 0: %+v", len(got), got)
	}
}

func TestRuleSetWithoutAccumulatesIDs(t *testing.T) {
	rs := RuleSet{}.Without("isolation.cross-domain").Without("blast.high-coupling")
	if !rs.IsViolationSkipped("isolation.cross-domain") {
		t.Errorf("isolation.cross-domain should be skipped")
	}
	if !rs.IsViolationSkipped("blast.high-coupling") {
		t.Errorf("blast.high-coupling should be skipped")
	}
	if rs.IsViolationSkipped("naming.no-stutter") {
		t.Errorf("naming.no-stutter should NOT be skipped")
	}
}

func TestRuleSetWithDoesNotMutateOriginal(t *testing.T) {
	r1 := &fakeRule{spec: RuleSpec{ID: "a"}}
	r2 := &fakeRule{spec: RuleSpec{ID: "b"}}
	base := RuleSet{}.With(r1)
	extended := base.With(r2)
	if len(base.Rules()) != 1 {
		t.Errorf("base mutated: len = %d", len(base.Rules()))
	}
	if len(extended.Rules()) != 2 {
		t.Errorf("extended len = %d", len(extended.Rules()))
	}
}

func TestRuleSetWithoutDoesNotMutateOriginal(t *testing.T) {
	base := RuleSet{}.Without("a.b")
	extended := base.Without("c.d")
	if base.IsViolationSkipped("c.d") {
		t.Errorf("base mutated: c.d should not be skipped on base")
	}
	if !extended.IsViolationSkipped("c.d") {
		t.Errorf("extended.c.d should be skipped")
	}
}
