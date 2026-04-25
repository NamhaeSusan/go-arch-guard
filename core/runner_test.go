package core

import (
	"strings"
	"testing"
)

func TestRunReturnsViolationsInDeterministicOrder(t *testing.T) {
	r := &fakeRule{
		spec: RuleSpec{
			ID: "fake.demo",
			Violations: []ViolationSpec{
				{ID: "fake.demo", DefaultSeverity: Error},
			},
		},
		violations: []Violation{
			{File: "b.go", Line: 1, Rule: "fake.demo", Message: "second"},
			{File: "a.go", Line: 10, Rule: "fake.demo", Message: "first"},
			{File: "a.go", Line: 2, Rule: "fake.demo", Message: "earlier"},
		},
	}
	ctx := NewContext(nil, "", "", validArchitecture(), nil)
	got := Run(ctx, RuleSet{}.With(r))
	if len(got) != 3 {
		t.Fatalf("len = %d, want 3", len(got))
	}
	if got[0].File != "a.go" || got[0].Line != 2 {
		t.Errorf("[0] = %s:%d, want a.go:2", got[0].File, got[0].Line)
	}
	if got[1].File != "a.go" || got[1].Line != 10 {
		t.Errorf("[1] = %s:%d, want a.go:10", got[1].File, got[1].Line)
	}
	if got[2].File != "b.go" {
		t.Errorf("[2].File = %s, want b.go", got[2].File)
	}
}

func TestRunSkipsViolationsFilteredByWithout(t *testing.T) {
	r := &fakeRule{
		spec: RuleSpec{
			ID: "fake.demo",
			Violations: []ViolationSpec{
				{ID: "fake.demo", DefaultSeverity: Error},
				{ID: "fake.demo.sub", DefaultSeverity: Error},
			},
		},
		violations: []Violation{
			{File: "a.go", Line: 1, Rule: "fake.demo", Message: "keep"},
			{File: "a.go", Line: 2, Rule: "fake.demo.sub", Message: "drop"},
		},
	}
	ctx := NewContext(nil, "", "", validArchitecture(), nil)
	got := Run(ctx, RuleSet{}.With(r).Without("fake.demo.sub"))
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	if got[0].Rule != "fake.demo" {
		t.Errorf("got %q, want fake.demo", got[0].Rule)
	}
}

func TestRunAppliesSeverityPrecedence(t *testing.T) {
	// rule-level Error → violation-spec Warning → runtime Error.
	r := &fakeRule{
		spec: RuleSpec{
			ID:              "fake.demo",
			DefaultSeverity: Error,
			Violations: []ViolationSpec{
				{ID: "fake.demo", DefaultSeverity: Warning},
			},
		},
		violations: []Violation{
			{File: "a.go", Line: 1, Rule: "fake.demo", Message: "x"},
		},
	}
	ctx := NewContext(nil, "", "", validArchitecture(), nil)

	// No override: ViolationSpec.DefaultSeverity (Warning) wins over RuleSpec.
	noOverride := Run(ctx, RuleSet{}.With(r))
	if noOverride[0].DefaultSeverity != Warning {
		t.Errorf("DefaultSeverity = %v, want Warning", noOverride[0].DefaultSeverity)
	}
	if noOverride[0].EffectiveSeverity != Warning {
		t.Errorf("EffectiveSeverity = %v, want Warning", noOverride[0].EffectiveSeverity)
	}

	// Runtime override wins.
	overridden := Run(ctx, RuleSet{}.With(r), WithSeverityOverride("fake.demo", Error))
	if overridden[0].DefaultSeverity != Warning {
		t.Errorf("DefaultSeverity = %v, want Warning (rule-declared, unchanged)", overridden[0].DefaultSeverity)
	}
	if overridden[0].EffectiveSeverity != Error {
		t.Errorf("EffectiveSeverity = %v, want Error (overridden)", overridden[0].EffectiveSeverity)
	}
}

func TestRunFallsBackToRuleSpecDefaultSeverity(t *testing.T) {
	// No ViolationSpec entry for this ID → fall back to RuleSpec.DefaultSeverity.
	// The emitted ID matches RuleSpec.ID itself (single-ID rule emits its own ID).
	r := &fakeRule{
		spec: RuleSpec{
			ID:              "fake.demo",
			DefaultSeverity: Warning,
		},
		violations: []Violation{
			{File: "a.go", Line: 1, Rule: "fake.demo", Message: "x"},
		},
	}
	ctx := NewContext(nil, "", "", validArchitecture(), nil)
	got := Run(ctx, RuleSet{}.With(r))
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	if got[0].EffectiveSeverity != Warning {
		t.Errorf("EffectiveSeverity = %v, want Warning (RuleSpec fallback)", got[0].EffectiveSeverity)
	}
}

func TestRunAllowsUndeclaredMetaIDs(t *testing.T) {
	// Rules may emit meta.* violations (e.g. meta.no-matching-packages)
	// without declaring them in Spec().Violations. The runner must let
	// these through unchanged, not rewrite them as meta.unknown-violation-id.
	r := &fakeRule{
		spec: RuleSpec{
			ID:         "fake.demo",
			Violations: []ViolationSpec{{ID: "fake.demo.declared", DefaultSeverity: Error}},
		},
		violations: []Violation{
			{File: ".", Rule: "meta.no-matching-packages", Message: "no packages"},
		},
	}
	ctx := NewContext(nil, "", "", validArchitecture(), nil)
	got := Run(ctx, RuleSet{}.With(r))
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	if got[0].Rule != "meta.no-matching-packages" {
		t.Errorf("Rule = %q, want meta.no-matching-packages (must NOT be rewritten)", got[0].Rule)
	}
}

func TestRunReplacesUnknownEmittedIDWithMeta(t *testing.T) {
	// Rule emits a violation whose ID isn't in its own Spec().Violations
	// or RuleSpec.ID. This is a rule-author bug; runner replaces with
	// meta.unknown-violation-id rather than panicking.
	r := &fakeRule{
		spec: RuleSpec{
			ID:         "fake.demo",
			Violations: []ViolationSpec{{ID: "fake.demo.declared", DefaultSeverity: Error}},
		},
		violations: []Violation{
			{File: "a.go", Line: 1, Rule: "fake.demo.typo", Message: "boom"},
		},
	}
	ctx := NewContext(nil, "", "", validArchitecture(), nil)
	got := Run(ctx, RuleSet{}.With(r))
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	if got[0].Rule != "meta.unknown-violation-id" {
		t.Errorf("Rule = %q, want meta.unknown-violation-id", got[0].Rule)
	}
	if !strings.Contains(got[0].Message, "fake.demo.typo") {
		t.Errorf("Message = %q, want mention of fake.demo.typo", got[0].Message)
	}
}

func TestRunAcceptsRuleSpecIDAsValidEmittedID(t *testing.T) {
	// Single-ID rules whose Spec().Violations is empty must be allowed to
	// emit their RuleSpec.ID as Violation.Rule.
	r := &fakeRule{
		spec: RuleSpec{ID: "naming.no-stutter", DefaultSeverity: Warning},
		violations: []Violation{
			{File: "a.go", Line: 1, Rule: "naming.no-stutter", Message: "stutters"},
		},
	}
	ctx := NewContext(nil, "", "", validArchitecture(), nil)
	got := Run(ctx, RuleSet{}.With(r))
	if len(got) != 1 || got[0].Rule != "naming.no-stutter" {
		t.Fatalf("got %+v, want single naming.no-stutter violation", got)
	}
}

func TestRunDedupesMetaViolations(t *testing.T) {
	r1 := &fakeRule{
		spec: RuleSpec{ID: "fake.a"},
		violations: []Violation{
			{File: ".", Rule: "meta.no-matching-packages", Message: "boom"},
		},
	}
	r2 := &fakeRule{
		spec: RuleSpec{ID: "fake.b"},
		violations: []Violation{
			{File: ".", Rule: "meta.no-matching-packages", Message: "boom"},
		},
	}
	ctx := NewContext(nil, "", "", validArchitecture(), nil)
	got := Run(ctx, RuleSet{}.With(r1).With(r2))
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1 (deduped)", len(got))
	}
}

func TestRunRejectsUnknownWithoutID(t *testing.T) {
	r := &fakeRule{
		spec: RuleSpec{
			ID:         "fake.demo",
			Violations: []ViolationSpec{{ID: "fake.demo", DefaultSeverity: Error}},
		},
	}
	ctx := NewContext(nil, "", "", validArchitecture(), nil)
	defer func() {
		rec := recover()
		if rec == nil {
			t.Fatal("expected panic for unknown Without ID")
		}
		msg, ok := rec.(string)
		if !ok {
			t.Fatalf("expected string panic, got %T: %v", rec, rec)
		}
		if !strings.Contains(msg, "fake.ghost") {
			t.Errorf("panic message = %q, want mention of fake.ghost", msg)
		}
	}()
	Run(ctx, RuleSet{}.With(r).Without("fake.ghost"))
}

func TestRunRejectsUnknownSeverityOverrideID(t *testing.T) {
	r := &fakeRule{
		spec: RuleSpec{
			ID:         "fake.demo",
			Violations: []ViolationSpec{{ID: "fake.demo", DefaultSeverity: Error}},
		},
	}
	ctx := NewContext(nil, "", "", validArchitecture(), nil)
	defer func() {
		rec := recover()
		if rec == nil {
			t.Fatal("expected panic for unknown WithSeverityOverride ID")
		}
		msg, ok := rec.(string)
		if !ok {
			t.Fatalf("expected string panic, got %T: %v", rec, rec)
		}
		if !strings.Contains(msg, "fake.ghost") {
			t.Errorf("panic message = %q, want mention of fake.ghost", msg)
		}
	}()
	Run(ctx, RuleSet{}.With(r), WithSeverityOverride("fake.ghost", Warning))
}

func TestRunAllowsMetaIDInWithoutAndOverride(t *testing.T) {
	// meta.* IDs are not in any rule's catalog, but callers may legitimately
	// filter or downgrade them — Without/WithSeverityOverride must not panic.
	r := &fakeRule{
		spec: RuleSpec{
			ID:         "fake.demo",
			Violations: []ViolationSpec{{ID: "fake.demo", DefaultSeverity: Error}},
		},
	}
	ctx := NewContext(nil, "", "", validArchitecture(), nil)

	// Both should run without panic.
	Run(ctx, RuleSet{}.With(r).Without("meta.no-matching-packages"))
	Run(ctx, RuleSet{}.With(r), WithSeverityOverride("meta.no-matching-packages", Warning))
}

func TestRunPanicsOnInvalidArchitecture(t *testing.T) {
	bad := validArchitecture()
	delete(bad.Layers.Direction, "core/svc")
	ctx := NewContext(nil, "", "", bad, nil)
	defer func() {
		rec := recover()
		if rec == nil {
			t.Fatal("expected panic for invalid Architecture")
		}
	}()
	Run(ctx, RuleSet{})
}
