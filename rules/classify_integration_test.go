package rules_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/rules"
)

// These tests prove that CheckLayerDirection and CheckDomainIsolation
// route through the unified classifier. Each scenario exercises a path
// the classifier must handle (shared, orchestration, domain, domain-root,
// unclassified).

func TestCheckLayerDirection_Classifier_SharedSkipped(t *testing.T) {
	root := t.TempDir()
	mod := "example.com/cls-shared"

	writeTestFile(t, filepath.Join(root, "go.mod"), "module "+mod+"\n\ngo 1.21\n")
	// pkg/ is classified as kindShared: source of pkg/ is ignored by layer rule.
	writeTestFile(t, filepath.Join(root, "internal", "pkg", "errors", "e.go"),
		"package errors\n")
	// Domain file that imports pkg/ — core is PkgRestricted, so inner-imports-pkg
	// must still fire. This proves import side classification still reaches
	// kindShared through the classifier.
	writeTestFile(t, filepath.Join(root, "internal", "domain", "order", "alias.go"),
		"package order\n")
	writeTestFile(t, filepath.Join(root, "internal", "domain", "order", "core", "model", "m.go"),
		"package model\n\nimport _ \""+mod+"/internal/pkg/errors\"\n")

	pkgs := loadTestPackages(t, root)
	violations := rules.CheckLayerDirection(pkgs, mod, root)

	found := false
	for _, v := range violations {
		if v.Rule == "layer.inner-imports-pkg" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected layer.inner-imports-pkg for core/model importing pkg/errors")
		for _, v := range violations {
			t.Log(v.String())
		}
	}
}

func TestCheckLayerDirection_Classifier_UnclassifiedSrcIsSkipped(t *testing.T) {
	root := t.TempDir()
	mod := "example.com/cls-unclass"

	writeTestFile(t, filepath.Join(root, "go.mod"), "module "+mod+"\n\ngo 1.21\n")
	// internal/config is neither domain/, orchestration/, nor pkg/ — kindUnclassified.
	// It imports a domain package. layer.go must skip unclassified sources.
	writeTestFile(t, filepath.Join(root, "internal", "domain", "order", "alias.go"),
		"package order\n")
	writeTestFile(t, filepath.Join(root, "internal", "domain", "order", "core", "model", "m.go"),
		"package model\n")
	writeTestFile(t, filepath.Join(root, "internal", "config", "c.go"),
		"package config\n\nimport _ \""+mod+"/internal/domain/order/core/model\"\n")

	pkgs := loadTestPackages(t, root)
	violations := rules.CheckLayerDirection(pkgs, mod, root)

	for _, v := range violations {
		if v.File == "internal/config/c.go" {
			t.Errorf("unclassified source must be skipped by layer rule, got %s", v.String())
		}
	}
}

func TestCheckDomainIsolation_Classifier_StrayImportsDomain(t *testing.T) {
	root := t.TempDir()
	mod := "example.com/cls-iso-stray"

	writeTestFile(t, filepath.Join(root, "go.mod"), "module "+mod+"\n\ngo 1.21\n")
	writeTestFile(t, filepath.Join(root, "internal", "domain", "order", "alias.go"),
		"package order\n")
	writeTestFile(t, filepath.Join(root, "internal", "domain", "order", "core", "model", "m.go"),
		"package model\n")
	// internal/config is kindUnclassified; importing a domain must trip
	// stray-imports-domain via the classifier.
	writeTestFile(t, filepath.Join(root, "internal", "config", "c.go"),
		"package config\n\nimport _ \""+mod+"/internal/domain/order/core/model\"\n")

	pkgs := loadTestPackages(t, root)
	violations := rules.CheckDomainIsolation(pkgs, mod, root)

	found := false
	for _, v := range violations {
		if v.Rule == "isolation.stray-imports-domain" &&
			strings.Contains(v.Message, "order") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected isolation.stray-imports-domain for config importing domain")
		for _, v := range violations {
			t.Log(v.String())
		}
	}
}

func TestCheckDomainIsolation_Classifier_OrchestrationDeepImport(t *testing.T) {
	root := t.TempDir()
	mod := "example.com/cls-iso-orch"

	writeTestFile(t, filepath.Join(root, "go.mod"), "module "+mod+"\n\ngo 1.21\n")
	writeTestFile(t, filepath.Join(root, "internal", "domain", "order", "alias.go"),
		"package order\n")
	writeTestFile(t, filepath.Join(root, "internal", "domain", "order", "core", "model", "m.go"),
		"package model\n")
	// orchestration/ importing a domain sub-package (not the alias) must fire
	// orchestration-deep-import when RequireAlias is true (DDD default).
	writeTestFile(t, filepath.Join(root, "internal", "orchestration", "o.go"),
		"package orchestration\n\nimport _ \""+mod+"/internal/domain/order/core/model\"\n")

	pkgs := loadTestPackages(t, root)
	violations := rules.CheckDomainIsolation(pkgs, mod, root)

	found := false
	for _, v := range violations {
		if v.Rule == "isolation.orchestration-deep-import" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected isolation.orchestration-deep-import via classifier")
		for _, v := range violations {
			t.Log(v.String())
		}
	}
}

func TestCheckDomainIsolation_Classifier_DomainRootAllowed(t *testing.T) {
	root := t.TempDir()
	mod := "example.com/cls-iso-root"

	writeTestFile(t, filepath.Join(root, "go.mod"), "module "+mod+"\n\ngo 1.21\n")
	// Domain root (alias.go) importing its own sublayer must be allowed —
	// the classifier identifies it as kindDomainRoot with matching Domain.
	writeTestFile(t, filepath.Join(root, "internal", "domain", "order", "alias.go"),
		"package order\n\nimport _ \""+mod+"/internal/domain/order/core/model\"\n")
	writeTestFile(t, filepath.Join(root, "internal", "domain", "order", "core", "model", "m.go"),
		"package model\n")

	pkgs := loadTestPackages(t, root)
	violations := rules.CheckDomainIsolation(pkgs, mod, root)
	for _, v := range violations {
		if v.File == "internal/domain/order/alias.go" {
			t.Errorf("domain root importing own sublayer must be allowed, got %s", v.String())
		}
	}
}
