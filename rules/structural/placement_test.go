package structural_test

import (
	"path/filepath"
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/core"
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

func TestPlacementDTOAllowedLayersAcceptsNestedSublayer(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "internal", "domain", "order", "alias.go"), "package order\n")
	writeTestFile(t, filepath.Join(root, "internal", "domain", "order", "core", "model", "order.go"), "package model\n")
	writeTestFile(t, filepath.Join(root, "internal", "domain", "order", "core", "repo", "order_dto.go"), "package repo\n")

	arch := dddArch()
	arch.Structure.DTOAllowedLayers = []string{"core/repo"}
	ctx := core.NewContext(nil, "github.com/example/app", root, arch, nil)
	violations := core.Run(ctx, core.NewRuleSet(structural.NewPlacement()))

	for _, v := range violations {
		if v.Rule == "structure.dto-placement" {
			t.Fatalf("DTO under core/repo must be allowed when DTOAllowedLayers contains core/repo, got %s", v.String())
		}
	}
}

func TestPlacementDTOAllowedLayersStillRejectsUnlistedNestedSublayer(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "internal", "domain", "order", "alias.go"), "package order\n")
	writeTestFile(t, filepath.Join(root, "internal", "domain", "order", "core", "model", "order.go"), "package model\n")
	writeTestFile(t, filepath.Join(root, "internal", "domain", "order", "core", "svc", "order_dto.go"), "package svc\n")

	arch := dddArch()
	arch.Structure.DTOAllowedLayers = []string{"core/repo"}
	ctx := core.NewContext(nil, "github.com/example/app", root, arch, nil)
	violations := core.Run(ctx, core.NewRuleSet(structural.NewPlacement()))

	assertViolation(t, violations, "structure.dto-placement", "internal/domain/order/core/svc/order_dto.go")
}
