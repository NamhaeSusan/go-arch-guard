package rules_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/analyzer"
	"github.com/NamhaeSusan/go-arch-guard/rules"
)

func TestCheckDomainIsolation(t *testing.T) {
	t.Run("valid project has no violations", func(t *testing.T) {
		pkgs := loadValid(t)
		violations := rules.CheckDomainIsolation(pkgs, "github.com/kimtaeyun/testproject-dc", "../testdata/valid")
		if len(violations) > 0 {
			for _, v := range violations {
				t.Log(v.String())
			}
			t.Errorf("expected no violations, got %d", len(violations))
		}
	})

	t.Run("detects cross-domain violation", func(t *testing.T) {
		pkgs := loadInvalid(t)
		violations := rules.CheckDomainIsolation(pkgs, "github.com/kimtaeyun/testproject-dc-invalid", "../testdata/invalid")
		found := false
		for _, v := range violations {
			if v.Rule == "isolation.cross-domain" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected cross-domain violation")
		}
	})

	t.Run("detects unauthorized internal package importing domain root", func(t *testing.T) {
		pkgs := loadInvalid(t)
		violations := rules.CheckDomainIsolation(pkgs, "github.com/kimtaeyun/testproject-dc-invalid", "../testdata/invalid")
		found := false
		for _, v := range violations {
			if v.Rule == "isolation.stray-imports-domain" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected internal-imports-domain violation")
		}
	})

	t.Run("detects orchestration deep import", func(t *testing.T) {
		pkgs := loadInvalid(t)
		violations := rules.CheckDomainIsolation(pkgs, "github.com/kimtaeyun/testproject-dc-invalid", "../testdata/invalid")
		found := false
		for _, v := range violations {
			if v.Rule == "isolation.orchestration-deep-import" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected orchestration-deep-import violation")
		}
	})

	t.Run("detects cmd deep import", func(t *testing.T) {
		pkgs := loadInvalid(t)
		violations := rules.CheckDomainIsolation(pkgs, "github.com/kimtaeyun/testproject-dc-invalid", "../testdata/invalid")
		found := false
		for _, v := range violations {
			if v.Rule == "isolation.cmd-deep-import" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected cmd-deep-import violation")
		}
	})

	t.Run("detects pkg importing orchestration", func(t *testing.T) {
		pkgs := loadInvalid(t)
		violations := rules.CheckDomainIsolation(pkgs, "github.com/kimtaeyun/testproject-dc-invalid", "../testdata/invalid")
		found := false
		for _, v := range violations {
			if v.Rule == "isolation.pkg-imports-orchestration" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected pkg-imports-orchestration violation")
		}
	})

	t.Run("detects unauthorized internal package importing orchestration", func(t *testing.T) {
		pkgs := loadInvalid(t)
		violations := rules.CheckDomainIsolation(pkgs, "github.com/kimtaeyun/testproject-dc-invalid", "../testdata/invalid")
		found := false
		for _, v := range violations {
			if v.Rule == "isolation.stray-imports-orchestration" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected internal-imports-orchestration violation")
		}
	})

	t.Run("detects pkg importing domain", func(t *testing.T) {
		pkgs := loadInvalid(t)
		violations := rules.CheckDomainIsolation(pkgs, "github.com/kimtaeyun/testproject-dc-invalid", "../testdata/invalid")
		found := false
		for _, v := range violations {
			if v.Rule == "isolation.pkg-imports-domain" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected pkg-imports-domain violation")
		}
	})

	t.Run("detects unauthorized internal package importing domain sub-package", func(t *testing.T) {
		pkgs := loadInvalid(t)
		violations := rules.CheckDomainIsolation(pkgs, "github.com/kimtaeyun/testproject-dc-invalid", "../testdata/invalid")
		found := false
		for _, v := range violations {
			if v.Rule == "isolation.stray-imports-domain" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected internal-imports-domain violation")
		}
	})

	t.Run("detects domain importing orchestration", func(t *testing.T) {
		pkgs := loadInvalid(t)
		violations := rules.CheckDomainIsolation(pkgs, "github.com/kimtaeyun/testproject-dc-invalid", "../testdata/invalid")
		found := false
		for _, v := range violations {
			if v.Rule == "isolation.domain-imports-orchestration" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected domain-imports-orchestration violation")
		}
	})

	t.Run("project-relative exclude skips matching package", func(t *testing.T) {
		pkgs := loadInvalid(t)
		violations := rules.CheckDomainIsolation(pkgs, "github.com/kimtaeyun/testproject-dc-invalid", "../testdata/invalid",
			rules.WithExclude("internal/config/..."))
		for _, v := range violations {
			if v.File == "internal/config/config.go" || v.File == "internal/config/domain_alias.go" || v.File == "internal/config/orchestration.go" {
				t.Fatalf("expected config package to be excluded, got %s", v.String())
			}
		}
	})

	t.Run("module-qualified exclude does not skip matching package", func(t *testing.T) {
		pkgs := loadInvalid(t)
		violations := rules.CheckDomainIsolation(pkgs, "github.com/kimtaeyun/testproject-dc-invalid", "../testdata/invalid",
			rules.WithExclude("github.com/kimtaeyun/testproject-dc-invalid/internal/config/..."))
		found := false
		for _, v := range violations {
			if v.File == "internal/config/config.go" || v.File == "internal/config/domain_alias.go" || v.File == "internal/config/orchestration.go" {
				found = true
				break
			}
		}
		if !found {
			t.Fatal("expected module-qualified exclude to be ignored")
		}
	})
	t.Run("warns when module path matches no packages", func(t *testing.T) {
		pkgs, err := analyzer.Load("../testdata/valid", "internal/...")
		if err != nil {
			t.Fatal(err)
		}
		violations := rules.CheckDomainIsolation(pkgs, "github.com/wrong/module", "../testdata/valid")
		found := false
		for _, v := range violations {
			if v.Rule == "meta.no-matching-packages" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected meta.no-matching-packages warning for wrong module path")
		}
	})

	t.Run("flat layout skipped", func(t *testing.T) {
		root := t.TempDir()
		module := "example.com/workeriso"
		m := rules.ConsumerWorker()

		writeTestFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.21\n")
		writeTestFile(t, filepath.Join(root, "internal", "worker", "w.go"),
			"package worker\n\nimport _ \""+module+"/internal/service\"\n")
		writeTestFile(t, filepath.Join(root, "internal", "service", "s.go"),
			"package service\n")

		pkgs := loadTestPackages(t, root)
		violations := rules.CheckDomainIsolation(pkgs, module, root, rules.WithModel(m))
		if len(violations) > 0 {
			for _, v := range violations {
				t.Log(v.String())
			}
			t.Error("expected 0 domain isolation violations for flat layout")
		}
	})

	t.Run("auto-extracts module when empty string passed", func(t *testing.T) {
		pkgs := loadValid(t)
		violations := rules.CheckDomainIsolation(pkgs, "", "")
		if len(violations) > 0 {
			for _, v := range violations {
				t.Log(v.String())
			}
			t.Errorf("expected no violations with auto-extracted module, got %d", len(violations))
		}
	})

}

func TestCheckDomainIsolation_TransportCannotImportDomain(t *testing.T) {
	root := t.TempDir()
	module := "example.com/dddapp"

	writeTestFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.21\n")

	// domain/order/core/model
	writeTestFile(t, filepath.Join(root, "internal", "domain", "order", "alias.go"), "package order\n")
	writeTestFile(t, filepath.Join(root, "internal", "domain", "order", "core", "model", "model.go"), "package model\ntype Order struct{}\n")

	// internal/app/container.go — legal: imports domain (kindApp can import anything)
	writeTestFile(t, filepath.Join(root, "internal", "app", "container.go"),
		"package app\n\nimport _ \""+module+"/internal/domain/order/core/model\"\n")

	// internal/server/http/server.go — legal: imports internal/app only
	writeTestFile(t, filepath.Join(root, "internal", "server", "http", "server.go"),
		"package http\n\nimport _ \""+module+"/internal/app\"\n")

	// internal/server/http/bad.go — ILLEGAL: transport imports domain directly
	writeTestFile(t, filepath.Join(root, "internal", "server", "http", "bad.go"),
		"package http\n\nimport _ \""+module+"/internal/domain/order/core/model\"\n")

	pkgs := loadTestPackages(t, root)
	violations := rules.CheckDomainIsolation(pkgs, module, root, rules.WithModel(rules.DDD()))

	found := false
	for _, v := range violations {
		if v.Rule == "isolation.transport-imports-domain" {
			found = true
			break
		}
	}
	if !found {
		for _, v := range violations {
			t.Logf("violation: %s", v.String())
		}
		t.Error("expected isolation.transport-imports-domain violation")
	}
}

func TestCheckDomainIsolation_AppCanImportAnything(t *testing.T) {
	root := t.TempDir()
	module := "example.com/dddapp2"

	writeTestFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.21\n")

	// domain/order/core/model
	writeTestFile(t, filepath.Join(root, "internal", "domain", "order", "alias.go"), "package order\n")
	writeTestFile(t, filepath.Join(root, "internal", "domain", "order", "core", "model", "model.go"), "package model\ntype Order struct{}\n")

	// domain/user/core/model
	writeTestFile(t, filepath.Join(root, "internal", "domain", "user", "alias.go"), "package user\n")
	writeTestFile(t, filepath.Join(root, "internal", "domain", "user", "core", "model", "model.go"), "package model\ntype User struct{}\n")

	// internal/orchestration — cross-domain coordination
	writeTestFile(t, filepath.Join(root, "internal", "orchestration", "orch.go"), "package orchestration\n")

	// internal/app/container.go — imports multiple domains (legal)
	writeTestFile(t, filepath.Join(root, "internal", "app", "container.go"),
		"package app\n\nimport (\n\t_ \""+module+"/internal/domain/order/core/model\"\n\t_ \""+module+"/internal/domain/user/core/model\"\n\t_ \""+module+"/internal/orchestration\"\n)\n")

	pkgs := loadTestPackages(t, root)
	violations := rules.CheckDomainIsolation(pkgs, module, root, rules.WithModel(rules.DDD()))

	for _, v := range violations {
		t.Errorf("expected no violations for kindApp, got: %s", v.String())
	}
}

func TestCheckDomainIsolation_TransportCannotImportOrchestration(t *testing.T) {
	root := t.TempDir()
	module := "example.com/dddapp3"

	writeTestFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.21\n")

	writeTestFile(t, filepath.Join(root, "internal", "domain", "order", "alias.go"), "package order\n")
	writeTestFile(t, filepath.Join(root, "internal", "domain", "order", "core", "model", "model.go"), "package model\ntype Order struct{}\n")
	writeTestFile(t, filepath.Join(root, "internal", "orchestration", "orch.go"), "package orchestration\n")

	// transport imports orchestration — ILLEGAL
	writeTestFile(t, filepath.Join(root, "internal", "server", "http", "bad.go"),
		"package http\n\nimport _ \""+module+"/internal/orchestration\"\n")

	pkgs := loadTestPackages(t, root)
	violations := rules.CheckDomainIsolation(pkgs, module, root, rules.WithModel(rules.DDD()))

	found := false
	for _, v := range violations {
		if v.Rule == "isolation.transport-imports-orchestration" {
			found = true
			break
		}
	}
	if !found {
		for _, v := range violations {
			t.Logf("violation: %s", v.String())
		}
		t.Error("expected isolation.transport-imports-orchestration violation")
	}
}

func TestCheckDomainIsolation_TransportCannotImportUnclassified(t *testing.T) {
	root := t.TempDir()
	module := "example.com/dddapp4"

	writeTestFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.21\n")

	writeTestFile(t, filepath.Join(root, "internal", "domain", "order", "alias.go"), "package order\n")
	writeTestFile(t, filepath.Join(root, "internal", "domain", "order", "core", "model", "model.go"), "package model\ntype Order struct{}\n")

	// internal/bootstrap is unclassified (not domain/orchestration/shared/app/server).
	writeTestFile(t, filepath.Join(root, "internal", "bootstrap", "xyz", "boot.go"), "package xyz\n")

	// transport imports unclassified package — ILLEGAL under tightened policy.
	writeTestFile(t, filepath.Join(root, "internal", "server", "http", "bad.go"),
		"package http\n\nimport _ \""+module+"/internal/bootstrap/xyz\"\n")

	pkgs := loadTestPackages(t, root)
	violations := rules.CheckDomainIsolation(pkgs, module, root, rules.WithModel(rules.DDD()))

	var transportUnclassified []rules.Violation
	for _, v := range violations {
		if v.Rule == "isolation.transport-imports-unclassified" {
			transportUnclassified = append(transportUnclassified, v)
		}
	}
	if len(transportUnclassified) != 1 {
		for _, v := range violations {
			t.Logf("violation: %s", v.String())
		}
		t.Fatalf("expected exactly 1 isolation.transport-imports-unclassified violation, got %d", len(transportUnclassified))
	}
}

func TestCheckDomainIsolation_TransportCanImportSharedPkg(t *testing.T) {
	root := t.TempDir()
	module := "example.com/dddapp5"

	writeTestFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.21\n")

	writeTestFile(t, filepath.Join(root, "internal", "domain", "order", "alias.go"), "package order\n")
	writeTestFile(t, filepath.Join(root, "internal", "domain", "order", "core", "model", "model.go"), "package model\ntype Order struct{}\n")

	// internal/pkg/logger — shared utility
	writeTestFile(t, filepath.Join(root, "internal", "pkg", "logger", "logger.go"), "package logger\n")

	// transport imports internal/pkg/... — LEGAL.
	writeTestFile(t, filepath.Join(root, "internal", "server", "http", "server.go"),
		"package http\n\nimport _ \""+module+"/internal/pkg/logger\"\n")

	pkgs := loadTestPackages(t, root)
	violations := rules.CheckDomainIsolation(pkgs, module, root, rules.WithModel(rules.DDD()))

	for _, v := range violations {
		if strings.HasPrefix(v.Rule, "isolation.transport-") {
			t.Errorf("expected no transport-* violations, got: %s", v.String())
		}
	}
}
