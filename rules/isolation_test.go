package rules_test

import (
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/analyzer"
	"github.com/NamhaeSusan/go-arch-guard/rules"
)

func TestCheckDomainIsolation(t *testing.T) {
	t.Run("valid project has no violations", func(t *testing.T) {
		pkgs, err := analyzer.Load("../testdata/valid", "internal/...", "cmd/...")
		if err != nil {
			t.Fatal(err)
		}
		violations := rules.CheckDomainIsolation(pkgs, "github.com/kimtaeyun/testproject-dc", "../testdata/valid")
		if len(violations) > 0 {
			for _, v := range violations {
				t.Log(v.String())
			}
			t.Errorf("expected no violations, got %d", len(violations))
		}
	})

	t.Run("detects cross-domain violation", func(t *testing.T) {
		pkgs, err := analyzer.Load("../testdata/invalid", "internal/...")
		if err != nil {
			t.Fatal(err)
		}
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
		pkgs, err := analyzer.Load("../testdata/invalid", "internal/...")
		if err != nil {
			t.Fatal(err)
		}
		violations := rules.CheckDomainIsolation(pkgs, "github.com/kimtaeyun/testproject-dc-invalid", "../testdata/invalid")
		found := false
		for _, v := range violations {
			if v.Rule == "isolation.internal-imports-domain" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected internal-imports-domain violation")
		}
	})

	t.Run("detects orchestration deep import", func(t *testing.T) {
		pkgs, err := analyzer.Load("../testdata/invalid", "internal/...")
		if err != nil {
			t.Fatal(err)
		}
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
		pkgs, err := analyzer.Load("../testdata/invalid", "internal/...", "cmd/...")
		if err != nil {
			t.Fatal(err)
		}
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
		pkgs, err := analyzer.Load("../testdata/invalid", "internal/...")
		if err != nil {
			t.Fatal(err)
		}
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

	t.Run("detects unauthorized internal package importing domain sub-package", func(t *testing.T) {
		pkgs, err := analyzer.Load("../testdata/invalid", "internal/...")
		if err != nil {
			t.Fatal(err)
		}
		violations := rules.CheckDomainIsolation(pkgs, "github.com/kimtaeyun/testproject-dc-invalid", "../testdata/invalid")
		found := false
		for _, v := range violations {
			if v.Rule == "isolation.internal-imports-domain" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected internal-imports-domain violation")
		}
	})
}
