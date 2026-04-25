package integration_test

import (
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/analyzer"
)

func TestIntegration_BlastRadius(t *testing.T) {
	pkgs, err := analyzer.Load(fixturePath("testdata/blast"), "internal/...")
	if err != nil {
		t.Fatal(err)
	}

	violations := runDDD(pkgs, "github.com/kimtaeyun/testproject-blast", fixturePath("testdata/blast"))
	if len(violations) == 0 {
		t.Error("expected blast radius violations for hub package")
	}
	assertHasRule(t, violations, "blast.high-coupling")
}
