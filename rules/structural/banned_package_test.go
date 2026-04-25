package structural_test

import (
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/rules/structural"
)

func TestBannedPackage(t *testing.T) {
	t.Run("valid fixture has no banned package violations", func(t *testing.T) {
		violations := runRule(t, "../../testdata/valid", structural.NewBannedPackage())
		assertNoRulePrefix(t, violations, "structure.")
	})

	t.Run("detects invalid fixture banned and legacy packages", func(t *testing.T) {
		violations := runRule(t, "../../testdata/invalid", structural.NewBannedPackage())

		assertViolation(t, violations, "structure.banned-package", "internal/platform/common/")
		assertViolation(t, violations, "structure.legacy-package", "internal/router/")
	})

	t.Run("defaults to warning", func(t *testing.T) {
		violations := runRule(t, "../../testdata/invalid", structural.NewBannedPackage())
		for _, v := range violations {
			if v.Rule == "structure.banned-package" && v.DefaultSeverity != core.Warning {
				t.Fatalf("DefaultSeverity = %v, want Warning", v.DefaultSeverity)
			}
		}
	})
}
