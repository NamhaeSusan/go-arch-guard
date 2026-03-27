package report

import (
	"fmt"
	"os"
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/rules"
)

// AssertNoViolations prints all violations.
// Fails the test only if any ERROR-level violations exist.
// Warnings are printed to stderr so they are always visible without -v.
func AssertNoViolations(t testing.TB, violations []rules.Violation) {
	t.Helper()
	hasErrors := false
	for _, v := range violations {
		if v.Severity == rules.Warning {
			fmt.Fprintln(os.Stderr, v.String())
		} else {
			t.Log(v.String())
			hasErrors = true
		}
	}
	if hasErrors {
		t.Errorf("found %d architecture violation(s)", countErrors(violations))
	}
}

func countErrors(violations []rules.Violation) int {
	n := 0
	for _, v := range violations {
		if v.Severity == rules.Error {
			n++
		}
	}
	return n
}
