package core

import "testing"

func TestWithSeverityOverrideAccumulates(t *testing.T) {
	opts := newRunOpts(
		WithSeverityOverride("dependency.cross-domain", Warning),
		WithSeverityOverride("dependency.high-coupling", Warning),
	)
	if got, ok := opts.severityFor("dependency.cross-domain"); !ok || got != Warning {
		t.Errorf("severityFor(dependency.cross-domain) = (%v, %v), want (Warning, true)", got, ok)
	}
	if got, ok := opts.severityFor("dependency.high-coupling"); !ok || got != Warning {
		t.Errorf("severityFor(dependency.high-coupling) = (%v, %v)", got, ok)
	}
	if _, ok := opts.severityFor("naming.no-stutter"); ok {
		t.Errorf("severityFor(naming.no-stutter) should be (_, false)")
	}
}

func TestWithSeverityOverrideLastWins(t *testing.T) {
	opts := newRunOpts(
		WithSeverityOverride("dependency.cross-domain", Warning),
		WithSeverityOverride("dependency.cross-domain", Error),
	)
	if got, _ := opts.severityFor("dependency.cross-domain"); got != Error {
		t.Errorf("last-wins failed: got %v, want Error", got)
	}
}
