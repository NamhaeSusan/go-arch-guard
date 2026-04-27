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

	t.Run("defaults to error for both banned and legacy violations", func(t *testing.T) {
		violations := runRule(t, "../../testdata/invalid", structural.NewBannedPackage())
		var sawBanned, sawLegacy bool
		for _, v := range violations {
			switch v.Rule {
			case "structure.banned-package":
				sawBanned = true
				if v.DefaultSeverity != core.Error {
					t.Fatalf("banned DefaultSeverity = %v, want Error", v.DefaultSeverity)
				}
			case "structure.legacy-package":
				sawLegacy = true
				if v.DefaultSeverity != core.Error {
					t.Fatalf("legacy DefaultSeverity = %v, want Error", v.DefaultSeverity)
				}
			}
		}
		if !sawBanned || !sawLegacy {
			t.Fatalf("expected both banned and legacy violations in invalid fixture, got banned=%v legacy=%v", sawBanned, sawLegacy)
		}
	})
}
