package integration_test

import (
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/analyzer"
	"github.com/NamhaeSusan/go-arch-guard/report"
)

func TestIntegration_Valid(t *testing.T) {
	pkgs, err := analyzer.Load(fixturePath("testdata/valid"), "internal/...", "cmd/...")
	if err != nil {
		t.Fatal(err)
	}

	t.Run("domain isolation", func(t *testing.T) {
		report.AssertNoViolations(t, runDDD(pkgs, "github.com/kimtaeyun/testproject-dc", fixturePath("testdata/valid")))
	})
	t.Run("layer direction", func(t *testing.T) {
		report.AssertNoViolations(t, runDDD(pkgs, "github.com/kimtaeyun/testproject-dc", fixturePath("testdata/valid")))
	})
	t.Run("naming", func(t *testing.T) {
		report.AssertNoViolations(t, runDDD(pkgs, "github.com/kimtaeyun/testproject-dc", fixturePath("testdata/valid")))
	})
	t.Run("structure", func(t *testing.T) {
		report.AssertNoViolations(t, runDDD(pkgs, "github.com/kimtaeyun/testproject-dc", fixturePath("testdata/valid")))
	})
	t.Run("blast radius", func(t *testing.T) {
		report.AssertNoViolations(t, runDDD(pkgs, "github.com/kimtaeyun/testproject-dc", fixturePath("testdata/valid")))
	})
}
