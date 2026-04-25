package structural_test

import (
	"strings"
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

	t.Run("allowedNames hint is deterministic and comma-joined", func(t *testing.T) {
		var prev string
		for i := range 5 {
			violations := runRule(t, "../../testdata/invalid", structural.NewInternalTopLevel())
			var msg string
			for _, v := range violations {
				if v.Rule == "structure.internal-top-level" {
					msg = v.Fix
					break
				}
			}
			if msg == "" {
				t.Fatalf("no internal-top-level violation found")
			}
			if strings.Contains(msg, "[") || strings.Contains(msg, "]") {
				t.Fatalf("Fix should use comma-joined names, got %q", msg)
			}
			if i > 0 && msg != prev {
				t.Fatalf("Fix non-deterministic: run %d %q != run 0 %q", i, msg, prev)
			}
			prev = msg
		}
	})
}
