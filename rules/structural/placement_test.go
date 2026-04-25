package structural_test

import (
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/rules/structural"
)

func TestPlacement(t *testing.T) {
	t.Run("valid fixture has no placement violations", func(t *testing.T) {
		violations := runRule(t, "../../testdata/valid", structural.NewPlacement())
		assertNoRulePrefix(t, violations, "structure.")
	})

	t.Run("detects invalid fixture placement violations", func(t *testing.T) {
		violations := runRule(t, "../../testdata/invalid", structural.NewPlacement())

		assertViolation(t, violations, "structure.misplaced-layer", "internal/platform/handler/")
		assertViolation(t, violations, "structure.middleware-placement", "internal/handler/middleware/")
		assertViolation(t, violations, "structure.dto-placement", "internal/domain/user/core/model/user_dto.go")
	})
}
