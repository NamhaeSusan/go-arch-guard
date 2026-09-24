package report

import (
	"github.com/NamhaeSusan/go-arch-guard/core"
	"testing"
)

func TestInvalidSeverityCountsAsError(t *testing.T) {
	r := BuildJSONReport([]core.Violation{{Rule: "demo", EffectiveSeverity: core.Severity(2)}})
	if r.Summary.Errors != 1 || r.Summary.Warnings != 0 || r.Violations[0].EffectiveSeverity != "error" {
		t.Fatalf("got %+v", r)
	}
}
