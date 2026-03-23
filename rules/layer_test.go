package rules_test

import (
	"strings"
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/analyzer"
	"github.com/NamhaeSusan/go-arch-guard/rules"
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
