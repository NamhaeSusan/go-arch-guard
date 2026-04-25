package structural_test

import (
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/rules/structural"
)

func TestInternalTopLevel(t *testing.T) {
	t.Run("valid fixture has no internal top-level violations", func(t *testing.T) {
		violations := runRule(t, "../../testdata/valid", structural.NewInternalTopLevel())
		assertNoRulePrefix(t, violations, "structure.internal-top-level")
	})

	t.Run("detects invalid fixture unexpected internal top-level package", func(t *testing.T) {
		violations := runRule(t, "../../testdata/invalid", structural.NewInternalTopLevel())
		assertViolation(t, violations, "structure.internal-top-level", "internal/config/")
	})
}
