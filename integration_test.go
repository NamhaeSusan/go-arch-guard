package goarchguard_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/analyzer"
	"github.com/NamhaeSusan/go-arch-guard/report"
	"github.com/NamhaeSusan/go-arch-guard/rules"
	"golang.org/x/tools/go/packages"
)

func TestIntegration_Valid(t *testing.T) {
	pkgs, err := analyzer.Load("testdata/valid", "internal/...", "cmd/...")
	if err != nil {
		t.Fatal(err)
	}

	t.Run("domain isolation", func(t *testing.T) {
		report.AssertNoViolations(t, rules.CheckDomainIsolation(pkgs, "github.com/kimtaeyun/testproject-dc", "testdata/valid"))
	})
	t.Run("layer direction", func(t *testing.T) {
		report.AssertNoViolations(t, rules.CheckLayerDirection(pkgs, "github.com/kimtaeyun/testproject-dc", "testdata/valid"))
	})
	t.Run("naming", func(t *testing.T) {
		report.AssertNoViolations(t, rules.CheckNaming(pkgs))
	})
	t.Run("structure", func(t *testing.T) {
		report.AssertNoViolations(t, rules.CheckStructure("testdata/valid"))
	})
	t.Run("blast radius", func(t *testing.T) {
		report.AssertNoViolations(t, rules.AnalyzeBlastRadius(pkgs, "github.com/kimtaeyun/testproject-dc", "testdata/valid"))
	})
}

func TestIntegration_BlastRadius(t *testing.T) {
	pkgs, err := analyzer.Load("testdata/blast", "internal/...")
	if err != nil {
		t.Fatal(err)
	}

	violations := rules.AnalyzeBlastRadius(pkgs, "github.com/kimtaeyun/testproject-blast", "testdata/blast")
	if len(violations) == 0 {
		t.Error("expected blast radius violations for hub package")
	}
	assertHasRule(t, violations, "blast-radius.high-coupling")
}

func TestIntegration_Invalid(t *testing.T) {
	pkgs, err := analyzer.Load("testdata/invalid", "internal/...", "cmd/...")
	if err != nil {
		t.Fatal(err)
	}

	t.Run("domain isolation violations found", func(t *testing.T) {
		violations := rules.CheckDomainIsolation(pkgs, "github.com/kimtaeyun/testproject-dc-invalid", "testdata/invalid")
		if len(violations) == 0 {
			t.Error("expected domain isolation violations")
		}
	})
	t.Run("layer direction violations found", func(t *testing.T) {
		violations := rules.CheckLayerDirection(pkgs, "github.com/kimtaeyun/testproject-dc-invalid", "testdata/invalid")
		if len(violations) == 0 {
			t.Error("expected layer direction violations")
		}
	})
	t.Run("naming violations found", func(t *testing.T) {
		violations := rules.CheckNaming(pkgs)
		if len(violations) == 0 {
			t.Error("expected naming violations")
		}
	})
	t.Run("structure violations found", func(t *testing.T) {
		violations := rules.CheckStructure("testdata/invalid")
		if len(violations) == 0 {
			t.Error("expected structure violations")
		}
	})

	t.Run("new rule ids are surfaced", func(t *testing.T) {
		isolationViolations := rules.CheckDomainIsolation(pkgs, "github.com/kimtaeyun/testproject-dc-invalid", "testdata/invalid")
		layerViolations := rules.CheckLayerDirection(pkgs, "github.com/kimtaeyun/testproject-dc-invalid", "testdata/invalid")
		structureViolations := rules.CheckStructure("testdata/invalid")

		assertHasRule(t, isolationViolations, "isolation.domain-imports-orchestration")
		assertHasRule(t, isolationViolations, "isolation.internal-imports-orchestration")
		assertHasRule(t, isolationViolations, "isolation.pkg-imports-domain")
		assertHasRule(t, layerViolations, "layer.unknown-sublayer")
		assertHasRule(t, layerViolations, "layer.inner-imports-pkg")
		assertHasRule(t, structureViolations, "structure.internal-top-level")
		assertHasRule(t, structureViolations, "structure.domain-root-alias-required")
		assertHasRule(t, structureViolations, "structure.domain-model-required")
		assertHasRule(t, structureViolations, "structure.dto-placement")
	})
}

func TestIntegration_WarningMode(t *testing.T) {
	pkgs, err := analyzer.Load("testdata/invalid", "internal/...", "cmd/...")
	if err != nil {
		t.Fatal(err)
	}

	violations := rules.CheckDomainIsolation(pkgs, "github.com/kimtaeyun/testproject-dc-invalid", "testdata/invalid",
		rules.WithSeverity(rules.Warning))
	if len(violations) == 0 {
		t.Error("expected violations even in warning mode")
	}
	for _, v := range violations {
		if v.Severity != rules.Warning {
			t.Errorf("expected Warning severity, got %v", v.Severity)
		}
	}
	report.AssertNoViolations(t, violations)
}

func TestIntegration_RejectsUnexpectedInternalTopLevelPackages(t *testing.T) {
	root := t.TempDir()
	module := "example.com/supportzones"

	writeIntegrationFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.25.0\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "order", "alias.go"), "package order\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "order", "core", "model", "order.go"), "package model\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "config", "config.go"), "package config\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "platform", "platform.go"), "package platform\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "system", "system.go"), "package system\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "foundation", "foundation.go"), "package foundation\n")

	pkgs, err := analyzer.Load(root, "internal/...")
	if err != nil {
		t.Fatal(err)
	}

	assertHasPackage(t, pkgs, module+"/internal/config")
	assertHasPackage(t, pkgs, module+"/internal/platform")
	assertHasPackage(t, pkgs, module+"/internal/system")
	assertHasPackage(t, pkgs, module+"/internal/foundation")

	t.Run("domain isolation", func(t *testing.T) {
		report.AssertNoViolations(t, rules.CheckDomainIsolation(pkgs, module, root))
	})
	t.Run("layer direction", func(t *testing.T) {
		report.AssertNoViolations(t, rules.CheckLayerDirection(pkgs, module, root))
	})
	t.Run("naming", func(t *testing.T) {
		report.AssertNoViolations(t, rules.CheckNaming(pkgs))
	})
	t.Run("structure", func(t *testing.T) {
		violations := rules.CheckStructure(root)
		assertHasRule(t, violations, "structure.internal-top-level")
	})
}

func assertHasRule(t *testing.T, violations []rules.Violation, rule string) {
	t.Helper()
	for _, v := range violations {
		if v.Rule == rule {
			return
		}
	}
	t.Fatalf("expected rule %q", rule)
}

func assertHasPackage(t *testing.T, pkgs []*packages.Package, pkgPath string) {
	t.Helper()
	for _, pkg := range pkgs {
		if pkg.PkgPath == pkgPath {
			return
		}
	}
	t.Fatalf("expected package %q to be loaded", pkgPath)
}

func writeIntegrationFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
