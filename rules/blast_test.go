package rules_test

import (
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/rules"
)

func TestAnalyzeBlastRadius(t *testing.T) {
	t.Run("returns no violations for small project", func(t *testing.T) {
		pkgs := loadValid(t)
		violations := rules.AnalyzeBlastRadius(pkgs, "github.com/kimtaeyun/testproject-dc", "../testdata/valid")
		for _, v := range violations {
			if v.Severity == rules.Error {
				t.Errorf("unexpected error violation: %s", v.String())
			}
		}
	})
}
