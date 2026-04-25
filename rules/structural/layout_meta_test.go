package structural_test

import (
	"strings"
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/rules/structural"
)

func TestStructuralRulesEmitMetaOnFlatLayout(t *testing.T) {
	cases := []struct {
		ruleID string
		rule   core.Rule
	}{
		{"structural.alias", structural.NewAlias()},
		{"structural.placement", structural.NewPlacement()},
		{"structural.banned-package", structural.NewBannedPackage()},
		{"structural.model-required", structural.NewModelRequired()},
		{"structural.internal-top-level", structural.NewInternalTopLevel()},
	}
	for _, tc := range cases {
		t.Run(tc.ruleID, func(t *testing.T) {
			violations := runRule(t, "../../testdata/flat", tc.rule)
			var count int
			for _, v := range violations {
				if v.Rule == "meta.layout-not-supported" {
					count++
					if !strings.Contains(v.Message, tc.ruleID) {
						t.Fatalf("meta message should mention %q, got %q", tc.ruleID, v.Message)
					}
				}
			}
			if count != 1 {
				t.Fatalf("expected exactly 1 meta.layout-not-supported, got %d: %+v", count, violations)
			}
		})
	}
}

func TestStructuralRulesNoMetaOnDDDLayout(t *testing.T) {
	cases := []core.Rule{
		structural.NewAlias(),
		structural.NewPlacement(),
		structural.NewBannedPackage(),
		structural.NewModelRequired(),
		structural.NewInternalTopLevel(),
	}
	for _, rule := range cases {
		violations := runRule(t, "../../testdata/valid", rule)
		for _, v := range violations {
			if v.Rule == "meta.layout-not-supported" {
				t.Fatalf("internal/-based project must not emit meta.layout-not-supported: %s", v.String())
			}
		}
	}
}
