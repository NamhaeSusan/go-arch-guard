package structural_test

import (
	"strings"
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/rules/structural"
)

func TestModelRequired(t *testing.T) {
	t.Run("valid fixture has no model-required violations", func(t *testing.T) {
		violations := runRule(t, "../../testdata/valid", structural.NewModelRequired())
		assertNoRulePrefix(t, violations, "structural.domain-model-required")
	})

	t.Run("detects invalid fixture missing domain model", func(t *testing.T) {
		violations := runRule(t, "../../testdata/invalid", structural.NewModelRequired())
		assertViolation(t, violations, "structural.domain-model-required", "internal/domain/ghost/")
	})

	t.Run("emits meta.rule-disabled-by-config when DomainDir is empty (flat layout)", func(t *testing.T) {
		arch := dddArch()
		arch.Layout.DomainDir = ""
		ctx := core.NewContext(nil, "github.com/example/app", "../../testdata/invalid", arch, nil)

		got := structural.NewModelRequired().Check(ctx)
		if len(got) != 1 || got[0].Rule != "meta.rule-disabled-by-config" {
			t.Fatalf("expected exactly 1 meta.rule-disabled-by-config violation, got %+v", got)
		}
		if !strings.Contains(got[0].Message, "Layout.DomainDir is empty") {
			t.Fatalf("meta message should mention DomainDir, got %q", got[0].Message)
		}
	})

	t.Run("emits meta.rule-disabled-by-config when RequireModel is false", func(t *testing.T) {
		arch := dddArch()
		arch.Structure.RequireModel = false
		ctx := core.NewContext(nil, "github.com/example/app", "../../testdata/invalid", arch, nil)

		got := structural.NewModelRequired().Check(ctx)
		if len(got) != 1 || got[0].Rule != "meta.rule-disabled-by-config" {
			t.Fatalf("expected exactly 1 meta.rule-disabled-by-config violation, got %+v", got)
		}
		if !strings.Contains(got[0].Message, "RequireModel is false") {
			t.Fatalf("meta message should mention RequireModel, got %q", got[0].Message)
		}
	})
}
