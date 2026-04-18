package rules_test

import (
	"path/filepath"
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
