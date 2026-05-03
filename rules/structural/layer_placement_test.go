package structural_test

import (
	"path/filepath"
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/core"
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

	t.Run("detects custom layer names using configured locations", func(t *testing.T) {
		root := t.TempDir()
		writeTestFile(t, filepath.Join(root, "internal", "domain", "order", "controller", "controller.go"), "package controller\n")
		writeTestFile(t, filepath.Join(root, "internal", "platform", "controller", "controller.go"), "package controller\n")

		arch := dddArch()
		arch.Layers.LayerDirNames["controller"] = true
		arch.Layers.LayerLocations = map[string][]string{
			"controller": {"{InternalRoot}/{DomainDir}/*/controller"},
		}
		ctx := core.NewContext(nil, "github.com/example/app", root, arch, nil)

		violations := core.Run(ctx, core.NewRuleSet(structural.NewLayerPlacement()))
		assertViolation(t, violations, "structural.misplaced-layer", "internal/platform/controller/")
		assertNoViolationAt(t, violations, "structural.misplaced-layer", "internal/domain/order/controller/")
	})

	t.Run("keeps built-in fallback checks when custom locations are configured", func(t *testing.T) {
		root := t.TempDir()
		writeTestFile(t, filepath.Join(root, "internal", "platform", "handler", "handler.go"), "package handler\n")
		writeTestFile(t, filepath.Join(root, "internal", "domain", "order", "controller", "controller.go"), "package controller\n")

		arch := dddArch()
		arch.Layers.LayerDirNames["controller"] = true
		arch.Layers.LayerLocations = map[string][]string{
			"controller": {"{InternalRoot}/{DomainDir}/*/controller"},
		}
		ctx := core.NewContext(nil, "github.com/example/app", root, arch, nil)

		violations := core.Run(ctx, core.NewRuleSet(structural.NewLayerPlacement()))
		assertViolation(t, violations, "structural.misplaced-layer", "internal/platform/handler/")
	})

	t.Run("treats LayerLocations keys as recognized layer names", func(t *testing.T) {
		root := t.TempDir()
		writeTestFile(t, filepath.Join(root, "internal", "platform", "usecase", "usecase.go"), "package usecase\n")

		arch := dddArch()
		delete(arch.Layers.LayerDirNames, "usecase")
		arch.Layers.LayerLocations = map[string][]string{
			"usecase": {"{InternalRoot}/{DomainDir}/*/usecase"},
		}
		ctx := core.NewContext(nil, "github.com/example/app", root, arch, nil)

		violations := core.Run(ctx, core.NewRuleSet(structural.NewLayerPlacement()))
		assertViolation(t, violations, "structural.misplaced-layer", "internal/platform/usecase/")
	})
}

func assertNoViolationAt(t *testing.T, violations []core.Violation, rule, file string) {
	t.Helper()
	for _, v := range violations {
		if v.Rule == rule && v.File == file {
			t.Fatalf("unexpected %s at %s in %#v", rule, file, violations)
		}
	}
}
