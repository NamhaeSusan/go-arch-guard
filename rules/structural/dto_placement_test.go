package structural_test

import (
	"path/filepath"
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/rules/structural"
)

func TestDTOPlacement(t *testing.T) {
	t.Run("valid fixture has no dto-placement violations", func(t *testing.T) {
		violations := runRule(t, "../../testdata/valid", structural.NewDTOPlacement())
		assertNoRulePrefix(t, violations, "structural.dto-placement")
	})

	t.Run("detects invalid fixture DTO outside DTOAllowedLayers", func(t *testing.T) {
		violations := runRule(t, "../../testdata/invalid", structural.NewDTOPlacement())
		assertViolation(t, violations, "structural.dto-placement", "internal/domain/user/core/model/user_dto.go")
	})

	t.Run("DTOAllowedLayers accepts nested sublayer", func(t *testing.T) {
		root := t.TempDir()
		writeTestFile(t, filepath.Join(root, "internal", "domain", "order", "alias.go"), "package order\n")
		writeTestFile(t, filepath.Join(root, "internal", "domain", "order", "core", "model", "order.go"), "package model\n")
		writeTestFile(t, filepath.Join(root, "internal", "domain", "order", "core", "repo", "order_dto.go"), "package repo\n")

		arch := dddArch()
		arch.Structure.DTOAllowedLayers = []string{"core/repo"}
		ctx := core.NewContext(nil, "github.com/example/app", root, arch, nil)
		violations := core.Run(ctx, core.NewRuleSet(structural.NewDTOPlacement()))

		for _, v := range violations {
			if v.Rule == "structural.dto-placement" {
				t.Fatalf("DTO under core/repo must be allowed when DTOAllowedLayers contains core/repo, got %s", v.String())
			}
		}
	})

	t.Run("DTOAllowedLayers still rejects unlisted nested sublayer", func(t *testing.T) {
		root := t.TempDir()
		writeTestFile(t, filepath.Join(root, "internal", "domain", "order", "alias.go"), "package order\n")
		writeTestFile(t, filepath.Join(root, "internal", "domain", "order", "core", "model", "order.go"), "package model\n")
		writeTestFile(t, filepath.Join(root, "internal", "domain", "order", "core", "svc", "order_dto.go"), "package svc\n")

		arch := dddArch()
		arch.Structure.DTOAllowedLayers = []string{"core/repo"}
		ctx := core.NewContext(nil, "github.com/example/app", root, arch, nil)
		violations := core.Run(ctx, core.NewRuleSet(structural.NewDTOPlacement()))

		assertViolation(t, violations, "structural.dto-placement", "internal/domain/order/core/svc/order_dto.go")
	})

	t.Run("WithDTOFilenameSuffixes overrides the suffix list", func(t *testing.T) {
		root := t.TempDir()
		writeTestFile(t, filepath.Join(root, "internal", "domain", "order", "alias.go"), "package order\n")
		writeTestFile(t, filepath.Join(root, "internal", "domain", "order", "core", "model", "order.go"), "package model\n")
		// "order_payload.go" must be flagged when suffix list is "_payload.go"; existing "_dto.go" pattern should NOT fire.
		writeTestFile(t, filepath.Join(root, "internal", "domain", "order", "core", "svc", "order_payload.go"), "package svc\n")
		writeTestFile(t, filepath.Join(root, "internal", "domain", "order", "core", "svc", "order_dto.go"), "package svc\n")

		ctx := core.NewContext(nil, "github.com/example/app", root, dddArch(), nil)
		got := core.Run(ctx, core.NewRuleSet(structural.NewDTOPlacement(structural.WithDTOFilenameSuffixes("_payload.go"))))

		var sawPayload, sawDTO bool
		for _, v := range got {
			if v.Rule != "structural.dto-placement" {
				continue
			}
			switch v.File {
			case "internal/domain/order/core/svc/order_payload.go":
				sawPayload = true
			case "internal/domain/order/core/svc/order_dto.go":
				sawDTO = true
			}
		}
		if !sawPayload {
			t.Fatalf("expected order_payload.go to be flagged with WithDTOFilenameSuffixes; got %+v", got)
		}
		if sawDTO {
			t.Fatalf("order_dto.go must not be flagged when suffix list is overridden away from _dto.go; got %+v", got)
		}
	})
}
