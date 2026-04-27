package structural_test

import (
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/rules/structural"
)

func TestLayerPlacement(t *testing.T) {
	t.Run("valid fixture has no layer-placement violations", func(t *testing.T) {
		violations := runRule(t, "../../testdata/valid", structural.NewLayerPlacement())
		assertNoRulePrefix(t, violations, "structural.misplaced-layer")
	})

	t.Run("detects invalid fixture misplaced layer dirs", func(t *testing.T) {
		violations := runRule(t, "../../testdata/invalid", structural.NewLayerPlacement())
		assertViolation(t, violations, "structural.misplaced-layer", "internal/platform/handler/")
	})
}
