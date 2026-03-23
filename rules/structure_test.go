package rules_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/rules"
)

func TestCheckStructure(t *testing.T) {
	t.Run("valid project has no violations", func(t *testing.T) {
		violations := rules.CheckStructure("../testdata/valid")
		if len(violations) > 0 {
			for _, v := range violations {
				t.Log(v.String())
			}
			t.Errorf("expected no violations, got %d", len(violations))
		}
	})

	t.Run("detects legacy packages", func(t *testing.T) {
		violations := rules.CheckStructure("../testdata/invalid")
		found := false
		for _, v := range violations {
			if v.Rule == "structure.legacy-package" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected legacy-package violation for internal/handler/")
		}
	})

	t.Run("detects router and bootstrap as legacy packages", func(t *testing.T) {
		violations := rules.CheckStructure("../testdata/invalid")
		foundRouter := false
		foundBootstrap := false
		for _, v := range violations {
			if v.Rule != "structure.legacy-package" {
				continue
			}
			if v.File == "internal/router/" {
				foundRouter = true
			}
			if v.File == "internal/bootstrap/" {
				foundBootstrap = true
			}
		}
		if !foundRouter || !foundBootstrap {
			t.Errorf("expected legacy-package violations for router and bootstrap, got router=%v bootstrap=%v", foundRouter, foundBootstrap)
		}
	})

	t.Run("detects middleware outside pkg", func(t *testing.T) {
		violations := rules.CheckStructure("../testdata/invalid")
		found := false
		for _, v := range violations {
			if v.Rule == "structure.middleware-placement" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected middleware-placement violation")
		}
	})

	t.Run("detects middleware nested under non-root pkg path", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, filepath.Join(root, "internal", "domain", "billing", "alias.go"), "package billing\n")
		writeFile(t, filepath.Join(root, "internal", "domain", "billing", "core", "model", "billing.go"), "package model\n")
		writeFile(t, filepath.Join(root, "internal", "domain", "billing", "pkg", "middleware", "auth.go"), "package middleware\n")

		violations := rules.CheckStructure(root)
		found := false
		for _, v := range violations {
			if v.Rule == "structure.middleware-placement" && v.File == "internal/domain/billing/pkg/middleware/" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected middleware nested outside internal/pkg to be rejected")
		}
	})

	t.Run("detects extra domain root files beyond alias.go", func(t *testing.T) {
		violations := rules.CheckStructure("../testdata/invalid")
		found := false
		for _, v := range violations {
			if v.Rule == "structure.domain-root-alias-only" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected domain-root-alias-only violation")
		}
	})

	t.Run("detects missing domain root alias", func(t *testing.T) {
		violations := rules.CheckStructure("../testdata/invalid")
		found := false
		for _, v := range violations {
			if v.Rule == "structure.domain-root-alias-required" && v.File == "internal/domain/noalias/" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected domain-root-alias-required violation")
		}
	})

	t.Run("detects missing domain model", func(t *testing.T) {
		violations := rules.CheckStructure("../testdata/invalid")
		found := false
		for _, v := range violations {
			if v.Rule == "structure.domain-model-required" && v.File == "internal/domain/ghost/" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected domain-model-required violation")
		}
	})

	t.Run("detects dto placement under domain", func(t *testing.T) {
		violations := rules.CheckStructure("../testdata/invalid")
		found := false
		for _, v := range violations {
			if v.Rule == "structure.dto-placement" && v.File == "internal/domain/user/core/model/user_dto.go" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected dto-placement violation")
		}
	})

	t.Run("detects recursive banned package name", func(t *testing.T) {
		violations := rules.CheckStructure("../testdata/invalid")
		found := false
		for _, v := range violations {
			if v.Rule == "structure.banned-package" && v.File == "internal/platform/common/" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected recursive banned-package violation")
		}
	})

	t.Run("detects misplaced handler directory", func(t *testing.T) {
		violations := rules.CheckStructure("../testdata/invalid")
		found := false
		for _, v := range violations {
			if v.Rule == "structure.legacy-package" && v.File == "internal/platform/handler/" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected misplaced handler legacy-package violation")
		}
	})

	t.Run("detects recursive bootstrap directory", func(t *testing.T) {
		violations := rules.CheckStructure("../testdata/invalid")
		found := false
		for _, v := range violations {
			if v.Rule == "structure.legacy-package" && v.File == "internal/platform/bootstrap/" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected recursive bootstrap legacy-package violation")
		}
	})

	t.Run("detects alias package name mismatch", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, filepath.Join(root, "internal", "domain", "billing", "alias.go"), "package billingapi\n")
		writeFile(t, filepath.Join(root, "internal", "domain", "billing", "core", "model", "billing.go"), "package model\n")

		violations := rules.CheckStructure(root)
		found := false
		for _, v := range violations {
			if v.Rule == "structure.domain-root-alias-package" && v.File == "internal/domain/billing/alias.go" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected domain-root-alias-package violation")
		}
	})

	t.Run("detects empty core model directory", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, filepath.Join(root, "internal", "domain", "billing", "alias.go"), "package billing\n")
		if err := os.MkdirAll(filepath.Join(root, "internal", "domain", "billing", "core", "model"), 0o755); err != nil {
			t.Fatal(err)
		}
		writeFile(t, filepath.Join(root, "internal", "domain", "billing", "core", "model", "README.md"), "# placeholder\n")

		violations := rules.CheckStructure(root)
		found := false
		for _, v := range violations {
			if v.Rule == "structure.domain-model-required" && v.File == "internal/domain/billing/" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected domain-model-required violation for empty core/model")
		}
	})

	t.Run("detects nested-only core model files", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, filepath.Join(root, "internal", "domain", "billing", "alias.go"), "package billing\n")
		writeFile(t, filepath.Join(root, "internal", "domain", "billing", "core", "model", "types", "billing.go"), "package types\n")

		violations := rules.CheckStructure(root)
		found := false
		for _, v := range violations {
			if v.Rule == "structure.domain-model-required" && v.File == "internal/domain/billing/" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected domain-model-required violation when core/model has no direct Go files")
		}
	})

	t.Run("root model file does not satisfy alias-only domain model requirement", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, filepath.Join(root, "internal", "domain", "billing", "alias.go"), "package billing\n")
		writeFile(t, filepath.Join(root, "internal", "domain", "billing", "model.go"), "package billing\n")

		violations := rules.CheckStructure(root)
		foundAliasOnly := false
		foundModelRequired := false
		for _, v := range violations {
			if v.Rule == "structure.domain-root-alias-only" && v.File == "internal/domain/billing/model.go" {
				foundAliasOnly = true
			}
			if v.Rule == "structure.domain-model-required" && v.File == "internal/domain/billing/" {
				foundModelRequired = true
			}
		}
		if !foundAliasOnly {
			t.Error("expected domain-root-alias-only violation for root model.go")
		}
		if !foundModelRequired {
			t.Error("expected domain-model-required violation when core/model is missing")
		}
	})

	t.Run("allows dto in handler sublayer", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, filepath.Join(root, "internal", "domain", "order", "alias.go"), "package order\n")
		writeFile(t, filepath.Join(root, "internal", "domain", "order", "core", "model", "order.go"), "package model\n")
		writeFile(t, filepath.Join(root, "internal", "domain", "order", "handler", "http", "dto.go"), "package http\n")

		violations := rules.CheckStructure(root)
		for _, v := range violations {
			if v.Rule == "structure.dto-placement" {
				t.Errorf("dto in handler/ should be allowed, got violation: %s", v.String())
			}
		}
	})

	t.Run("allows dto in app sublayer", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, filepath.Join(root, "internal", "domain", "order", "alias.go"), "package order\n")
		writeFile(t, filepath.Join(root, "internal", "domain", "order", "core", "model", "order.go"), "package model\n")
		writeFile(t, filepath.Join(root, "internal", "domain", "order", "app", "request_dto.go"), "package app\n")

		violations := rules.CheckStructure(root)
		for _, v := range violations {
			if v.Rule == "structure.dto-placement" {
				t.Errorf("dto in app/ should be allowed, got violation: %s", v.String())
			}
		}
	})

	t.Run("still rejects dto in core model", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, filepath.Join(root, "internal", "domain", "order", "alias.go"), "package order\n")
		writeFile(t, filepath.Join(root, "internal", "domain", "order", "core", "model", "order.go"), "package model\n")
		writeFile(t, filepath.Join(root, "internal", "domain", "order", "core", "model", "order_dto.go"), "package model\n")

		violations := rules.CheckStructure(root)
		found := false
		for _, v := range violations {
			if v.Rule == "structure.dto-placement" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected dto-placement violation for core/model/order_dto.go")
		}
	})

	t.Run("detects services as banned package", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, filepath.Join(root, "internal", "services", "order.go"), "package services\n")

		violations := rules.CheckStructure(root)
		found := false
		for _, v := range violations {
			if v.Rule == "structure.banned-package" && strings.Contains(v.Message, "services") {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected banned-package violation for internal/services/")
		}
	})

	t.Run("project-relative exclude skips matching directory tree", func(t *testing.T) {
		violations := rules.CheckStructure("../testdata/invalid", rules.WithExclude("internal/platform/..."))
		for _, v := range violations {
			if strings.HasPrefix(v.File, "internal/platform/") {
				t.Fatalf("expected platform subtree to be excluded, got %s", v.String())
			}
		}
	})
}

func writeFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
