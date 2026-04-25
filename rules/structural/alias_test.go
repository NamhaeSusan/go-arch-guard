package structural_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/rules/structural"
)

func TestAlias(t *testing.T) {
	t.Run("valid fixture has no alias violations", func(t *testing.T) {
		violations := runRule(t, "../../testdata/valid", structural.NewAlias())
		assertNoRulePrefix(t, violations, "structure.domain-alias-")
	})

	t.Run("detects invalid fixture alias violations", func(t *testing.T) {
		violations := runRule(t, "../../testdata/invalid", structural.NewAlias())

		assertViolation(t, violations, "structure.domain-alias-exists", "internal/domain/noalias/")
		assertViolation(t, violations, "structure.domain-alias-exclusive", "internal/domain/order/alias_bad.go")
		assertMessageContains(t, violations, "structure.domain-alias-no-interface", "CrossDomainOps")
	})

	t.Run("detects alias package name mismatch", func(t *testing.T) {
		root := t.TempDir()
		writeTestFile(t, filepath.Join(root, "internal", "domain", "billing", "alias.go"), "package billingapi\n")
		writeTestFile(t, filepath.Join(root, "internal", "domain", "billing", "core", "model", "billing.go"), "package model\n")

		violations := runRule(t, root, structural.NewAlias())
		assertViolation(t, violations, "structure.domain-alias-package", "internal/domain/billing/alias.go")
	})

	t.Run("detects contract re-export from alias", func(t *testing.T) {
		root := t.TempDir()
		writeTestFile(t, filepath.Join(root, "internal", "domain", "order", "alias.go"), `package order

import "example.com/app/internal/domain/order/core/svc"

type AdminOps = svc.AdminOps
`)
		writeTestFile(t, filepath.Join(root, "internal", "domain", "order", "core", "model", "order.go"), "package model\n")
		writeTestFile(t, filepath.Join(root, "internal", "domain", "order", "core", "svc", "admin.go"), "package svc\n\ntype AdminOps interface{ Do() }\n")

		violations := runRule(t, root, structural.NewAlias())
		assertMessageContains(t, violations, "structure.domain-alias-contract-reexport", "AdminOps")
	})

	t.Run("skips non-DDD architecture", func(t *testing.T) {
		arch := dddArch()
		arch.Layout.DomainDir = ""
		ctx := core.NewContext(nil, "github.com/example/app", "../../testdata/invalid", arch, nil)

		if got := structural.NewAlias().Check(ctx); len(got) != 0 {
			t.Fatalf("len = %d, want 0 for non-DDD architecture", len(got))
		}
	})
}

func dddArch() core.Architecture {
	return core.Architecture{
		Layers: core.LayerModel{
			Sublayers: []string{
				"handler", "app", "core", "core/model",
				"core/repo", "core/svc", "event", "infra",
			},
			Direction: map[string][]string{
				"handler":    {"app"},
				"app":        {"core/model", "core/repo", "core/svc", "event"},
				"core":       {"core/model"},
				"core/model": {},
				"core/repo":  {"core/model"},
				"core/svc":   {"core/model"},
				"event":      {"core/model"},
				"infra":      {"core/repo", "core/model", "event"},
			},
			ContractLayers: []string{"core/repo", "core/svc"},
			InternalTopLevel: map[string]bool{
				"domain": true, "orchestration": true, "pkg": true,
				"app": true, "server": true,
			},
			LayerDirNames: map[string]bool{
				"handler": true, "app": true, "core": true,
				"model": true, "repo": true, "svc": true,
				"event": true, "infra": true,
				"service": true, "controller": true,
				"entity": true, "store": true, "persistence": true,
				"domain": true,
			},
			PortLayers: []string{"core/repo"},
		},
		Layout: core.LayoutModel{
			DomainDir:        "domain",
			OrchestrationDir: "orchestration",
			SharedDir:        "pkg",
			AppDir:           "app",
			ServerDir:        "server",
		},
		Naming: core.NamingPolicy{
			AliasFileName:  "alias.go",
			BannedPkgNames: []string{"util", "common", "misc", "helper", "shared", "services"},
			LegacyPkgNames: []string{"router", "bootstrap"},
		},
		Structure: core.StructurePolicy{
			RequireAlias:     true,
			RequireModel:     true,
			ModelPath:        "core/model",
			DTOAllowedLayers: []string{"handler", "app"},
		},
	}
}

func runRule(t *testing.T, root string, rule core.Rule) []core.Violation {
	t.Helper()
	ctx := core.NewContext(nil, "github.com/example/app", root, dddArch(), nil)
	return core.Run(ctx, core.NewRuleSet(rule))
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func assertViolation(t *testing.T, violations []core.Violation, rule, file string) {
	t.Helper()
	for _, v := range violations {
		if v.Rule == rule && v.File == file {
			return
		}
	}
	t.Fatalf("expected %s at %s, got %#v", rule, file, violations)
}

func assertMessageContains(t *testing.T, violations []core.Violation, rule, substr string) {
	t.Helper()
	for _, v := range violations {
		if v.Rule == rule && strings.Contains(v.Message, substr) {
			return
		}
	}
	t.Fatalf("expected %s message containing %q, got %#v", rule, substr, violations)
}

func assertNoRulePrefix(t *testing.T, violations []core.Violation, prefix string) {
	t.Helper()
	for _, v := range violations {
		if strings.HasPrefix(v.Rule, prefix) {
			t.Fatalf("unexpected %s violation: %#v", prefix, v)
		}
	}
}
