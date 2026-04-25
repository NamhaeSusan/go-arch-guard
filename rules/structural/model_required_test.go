package structural_test

import (
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/rules/structural"
)

func TestModelRequired(t *testing.T) {
	t.Run("valid fixture has no model-required violations", func(t *testing.T) {
		violations := runRule(t, "../../testdata/valid", structural.NewModelRequired())
		assertNoRulePrefix(t, violations, "structure.domain-model-required")
	})

	t.Run("detects invalid fixture missing domain model", func(t *testing.T) {
		violations := runRule(t, "../../testdata/invalid", structural.NewModelRequired())
		assertViolation(t, violations, "structure.domain-model-required", "internal/domain/ghost/")
	})

	t.Run("skips non-DDD architecture", func(t *testing.T) {
		arch := dddArch()
		arch.Layout.DomainDir = ""
		ctx := core.NewContext(nil, "github.com/example/app", "../../testdata/invalid", arch, nil)

		if got := structural.NewModelRequired().Check(ctx); len(got) != 0 {
			t.Fatalf("len = %d, want 0 for non-DDD architecture", len(got))
		}
	})
}
