package report

import (
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/core"
)

// AssertNoViolations prints all violations.
// Fails the test only if any ERROR-level violations exist.
func AssertNoViolations(t testing.TB, violations []core.Violation) {
	t.Helper()
	hasErrors := false
	for _, v := range violations {
		t.Log(v.String())
		if v.EffectiveSeverity == core.Error {
			hasErrors = true
		}
	}
	if hasErrors {
		t.Errorf("found %d architecture violation(s)", countErrors(violations))
	}
}

func countErrors(violations []core.Violation) int {
	n := 0
	for _, v := range violations {
		if v.EffectiveSeverity == core.Error {
			n++
		}
	}
	return n
}
