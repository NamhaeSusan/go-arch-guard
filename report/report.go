package report

import (
	"fmt"
	"testing"

	"github.com/kimtaeyun/go-arch-guard/rules"
)

// AssertNoViolations prints all violations.
// Fails the test only if any ERROR-level violations exist.
func AssertNoViolations(t testing.TB, violations []rules.Violation) {
	t.Helper()
	hasErrors := false
	for _, v := range violations {
		fmt.Println(v.String())
		if v.Severity == rules.Error {
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
