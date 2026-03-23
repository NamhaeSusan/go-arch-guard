package rules_test

import (
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
}
