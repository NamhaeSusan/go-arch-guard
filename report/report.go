package report

import (
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/core"
)

// AssertNoViolations prints all violations.
// Fails the test only if any ERROR-level violations exist.
func AssertNoViolations(t testing.TB, violations []core.Violation) {
	t.Helper()
	errorCount := 0
	for _, v := range violations {
		t.Log(v.String())
		if v.EffectiveSeverity == core.Error {
			errorCount++
		}
	}
	if errorCount > 0 {
		t.Errorf("found %d architecture violation(s)", errorCount)
	}
}
