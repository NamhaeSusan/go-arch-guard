package rules_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/analyzer"
	"github.com/NamhaeSusan/go-arch-guard/rules"
	"golang.org/x/tools/go/packages"
)

func TestCheckLayerDirection(t *testing.T) {
	t.Run("valid project has no violations", func(t *testing.T) {
		pkgs := loadValid(t)
		violations := rules.CheckLayerDirection(pkgs, "github.com/kimtaeyun/testproject-dc", "../testdata/valid")
		if len(violations) > 0 {
			for _, v := range violations {
				t.Log(v.String())
			}
			t.Errorf("expected no violations, got %d", len(violations))
		}
	})

	t.Run("detects core importing app (reverse dependency)", func(t *testing.T) {
		pkgs := loadInvalid(t)
		violations := rules.CheckLayerDirection(pkgs, "github.com/kimtaeyun/testproject-dc-invalid", "../testdata/invalid")
		found := false
		for _, v := range violations {
			if v.Rule == "layer.direction" && strings.Contains(v.Message, `"core"`) && strings.Contains(v.Message, `"app"`) {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected core→app reverse dependency violation")
		}
	})

	t.Run("detects core/svc importing core/repo", func(t *testing.T) {
		pkgs := loadInvalid(t)
		violations := rules.CheckLayerDirection(pkgs, "github.com/kimtaeyun/testproject-dc-invalid", "../testdata/invalid")
		found := false
		for _, v := range violations {
			if v.Rule == "layer.direction" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected layer.direction violation")
		}
	})

	t.Run("detects unknown domain sublayer", func(t *testing.T) {
		pkgs := loadInvalid(t)
		violations := rules.CheckLayerDirection(pkgs, "github.com/kimtaeyun/testproject-dc-invalid", "../testdata/invalid")
		found := false
		for _, v := range violations {
			if v.Rule == "layer.unknown-sublayer" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected layer.unknown-sublayer violation")
		}
	})

	t.Run("detects inner layer importing pkg", func(t *testing.T) {
		pkgs := loadInvalid(t)
		violations := rules.CheckLayerDirection(pkgs, "github.com/kimtaeyun/testproject-dc-invalid", "../testdata/invalid")
		found := false
		for _, v := range violations {
			if v.Rule == "layer.inner-imports-pkg" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected layer.inner-imports-pkg violation")
		}
	})

	t.Run("detects handler importing event directly", func(t *testing.T) {
		pkgs := loadInvalid(t)
		violations := rules.CheckLayerDirection(pkgs, "github.com/kimtaeyun/testproject-dc-invalid", "../testdata/invalid")
		found := false
		for _, v := range violations {
			if v.Rule == "layer.direction" && strings.Contains(v.Message, `"handler"`) && strings.Contains(v.Message, `"event"`) {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected handler->event layer.direction violation")
		}
	})

	t.Run("project-relative exclude skips matching package", func(t *testing.T) {
		pkgs := loadInvalid(t)
		violations := rules.CheckLayerDirection(pkgs, "github.com/kimtaeyun/testproject-dc-invalid", "../testdata/invalid",
			rules.WithExclude("internal/domain/payment/core/model/..."))
		for _, v := range violations {
			if v.File == "internal/domain/payment/core/model/pkg_leak.go" {
				t.Fatalf("expected model package to be excluded, got %s", v.String())
			}
		}
	})

	t.Run("module-qualified exclude does not skip matching package", func(t *testing.T) {
		pkgs := loadInvalid(t)
		violations := rules.CheckLayerDirection(pkgs, "github.com/kimtaeyun/testproject-dc-invalid", "../testdata/invalid",
			rules.WithExclude("github.com/kimtaeyun/testproject-dc-invalid/internal/domain/payment/core/model/..."))
		found := false
		for _, v := range violations {
			if v.File == "internal/domain/payment/core/model/pkg_leak.go" {
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
		violations := rules.CheckLayerDirection(pkgs, "github.com/wrong/module", "../testdata/valid")
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

func loadTestPackages(t *testing.T, root string) []*packages.Package {
	t.Helper()
	pkgs, err := analyzer.Load(root, "internal/...")
	if err != nil {
		t.Log("partial load:", err)
	}
	if len(pkgs) == 0 {
		t.Fatal("no packages loaded")
	}
	return pkgs
}

func TestCheckLayerDirection_FlatLayout_Valid(t *testing.T) {
	root := t.TempDir()
	mod := "example.com/flat-valid"

	writeTestFile(t, filepath.Join(root, "go.mod"),
		"module "+mod+"\n\ngo 1.21\n")

	// worker → service (allowed)
	writeTestFile(t, filepath.Join(root, "internal", "worker", "w.go"),
		"package worker\n\nimport _ \""+mod+"/internal/service\"\n")
	// service → store (allowed)
	writeTestFile(t, filepath.Join(root, "internal", "service", "s.go"),
		"package service\n\nimport _ \""+mod+"/internal/store\"\n")
	// store → model (allowed)
	writeTestFile(t, filepath.Join(root, "internal", "store", "st.go"),
		"package store\n\nimport _ \""+mod+"/internal/model\"\n")
	// model has no imports
	writeTestFile(t, filepath.Join(root, "internal", "model", "m.go"),
		"package model\n\ntype Event struct{}\n")

	pkgs := loadTestPackages(t, root)
	violations := rules.CheckLayerDirection(pkgs, mod, root,
		rules.WithModel(rules.ConsumerWorker()))

	if len(violations) > 0 {
		for _, v := range violations {
			t.Log(v.String())
		}
		t.Errorf("expected no violations, got %d", len(violations))
	}
}

func TestCheckLayerDirection_FlatLayout_Violation(t *testing.T) {
	root := t.TempDir()
	mod := "example.com/flat-violation"

	writeTestFile(t, filepath.Join(root, "go.mod"),
		"module "+mod+"\n\ngo 1.21\n")

	// store → worker (NOT allowed, reverse direction)
	writeTestFile(t, filepath.Join(root, "internal", "store", "st.go"),
		"package store\n\nimport _ \""+mod+"/internal/worker\"\n")
	writeTestFile(t, filepath.Join(root, "internal", "worker", "w.go"),
		"package worker\n")
	writeTestFile(t, filepath.Join(root, "internal", "model", "m.go"),
		"package model\n")

	pkgs := loadTestPackages(t, root)
	violations := rules.CheckLayerDirection(pkgs, mod, root,
		rules.WithModel(rules.ConsumerWorker()))

	found := false
	for _, v := range violations {
		if v.Rule == "layer.direction" &&
			strings.Contains(v.Message, `"store"`) &&
			strings.Contains(v.Message, `"worker"`) {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected layer.direction violation for store→worker")
		for _, v := range violations {
			t.Log(v.String())
		}
	}
}

func TestCheckLayerDirection_FlatLayout_PkgImport(t *testing.T) {
	root := t.TempDir()
	mod := "example.com/flat-pkg"

	writeTestFile(t, filepath.Join(root, "go.mod"),
		"module "+mod+"\n\ngo 1.21\n")

	// worker → pkg (allowed, worker is not PkgRestricted)
	writeTestFile(t, filepath.Join(root, "internal", "worker", "w.go"),
		"package worker\n\nimport _ \""+mod+"/internal/pkg\"\n")
	// model → pkg (NOT allowed, model is PkgRestricted)
	writeTestFile(t, filepath.Join(root, "internal", "model", "m.go"),
		"package model\n\nimport _ \""+mod+"/internal/pkg\"\n")
	writeTestFile(t, filepath.Join(root, "internal", "pkg", "p.go"),
		"package pkg\n\nvar X = 1\n")

	pkgs := loadTestPackages(t, root)
	violations := rules.CheckLayerDirection(pkgs, mod, root,
		rules.WithModel(rules.ConsumerWorker()))

	foundPkgViolation := false
	for _, v := range violations {
		if v.Rule == "layer.inner-imports-pkg" &&
			strings.Contains(v.Message, `"model"`) {
			foundPkgViolation = true
		}
	}
	if !foundPkgViolation {
		t.Error("expected layer.inner-imports-pkg violation for model→pkg")
		for _, v := range violations {
			t.Log(v.String())
		}
	}

	// Ensure worker→pkg did NOT produce a violation
	for _, v := range violations {
		if v.Rule == "layer.inner-imports-pkg" &&
			strings.Contains(v.Message, `"worker"`) {
			t.Error("unexpected layer.inner-imports-pkg violation for worker→pkg")
		}
	}
}
