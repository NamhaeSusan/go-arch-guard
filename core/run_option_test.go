package core

import "testing"

func TestWithSeverityOverrideAccumulates(t *testing.T) {
	opts := newRunOpts(
		WithSeverityOverride("isolation.cross-domain", Warning),
		WithSeverityOverride("blast.high-coupling", Warning),
	)
	if got, ok := opts.severityFor("isolation.cross-domain"); !ok || got != Warning {
		t.Errorf("severityFor(isolation.cross-domain) = (%v, %v), want (Warning, true)", got, ok)
	}
	if got, ok := opts.severityFor("blast.high-coupling"); !ok || got != Warning {
		t.Errorf("severityFor(blast.high-coupling) = (%v, %v)", got, ok)
	}
	if _, ok := opts.severityFor("naming.no-stutter"); ok {
		t.Errorf("severityFor(naming.no-stutter) should be (_, false)")
	}
}

func TestWithSeverityOverrideLastWins(t *testing.T) {
	opts := newRunOpts(
		WithSeverityOverride("isolation.cross-domain", Warning),
		WithSeverityOverride("isolation.cross-domain", Error),
	)
	if got, _ := opts.severityFor("isolation.cross-domain"); got != Error {
		t.Errorf("last-wins failed: got %v, want Error", got)
	}
}
