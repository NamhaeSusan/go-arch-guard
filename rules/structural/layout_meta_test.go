package structural_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/analyzer"
	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/rules/structural"
)

func TestStructuralRulesEmitMetaOnFlatLayout(t *testing.T) {
	cases := []struct {
		ruleID string
		rule   core.Rule
	}{
		{"structural.alias", structural.NewAlias()},
		{"structural.layer-placement", structural.NewLayerPlacement()},
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
		structural.NewLayerPlacement(),
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

func TestStructuralRuleUsesRootDerivedByContext(t *testing.T) {
	root := t.TempDir()
	module := "example.com/derive-root"
	writeLayoutMetaFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.25.0\n")
	writeLayoutMetaFile(t, filepath.Join(root, "internal", "domain", "order", "alias.go"), "package order\n")
	writeLayoutMetaFile(t, filepath.Join(root, "internal", "domain", "order", "core", "model", "order.go"), "package model\n")
	writeLayoutMetaFile(t, filepath.Join(root, "internal", "config", "config.go"), "package config\n")

	pkgs, err := analyzer.Load(root, "internal/...")
	if err != nil {
		t.Fatal(err)
	}

	ctx := core.NewContext(pkgs, "", "", dddArch(), nil)
	violations := core.Run(ctx, core.NewRuleSet(structural.NewInternalTopLevel()))

	for _, v := range violations {
		if v.Rule == "meta.layout-not-supported" {
			t.Fatalf("expected derived root to enable structural checks, got meta violation: %s", v.String())
		}
	}
	assertViolation(t, violations, "structural.internal-top-level", "internal/config/")
}

func writeLayoutMetaFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
