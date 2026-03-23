package rules_test

import (
	"strings"
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/analyzer"
	"github.com/NamhaeSusan/go-arch-guard/rules"
)

func TestCheckLayerDirection(t *testing.T) {
	t.Run("valid project has no violations", func(t *testing.T) {
		pkgs, err := analyzer.Load("../testdata/valid", "internal/...")
		if err != nil {
			t.Fatal(err)
		}
		violations := rules.CheckLayerDirection(pkgs, "github.com/kimtaeyun/testproject-dc", "../testdata/valid")
		if len(violations) > 0 {
			for _, v := range violations {
				t.Log(v.String())
			}
			t.Errorf("expected no violations, got %d", len(violations))
		}
	})

	t.Run("detects core importing app (reverse dependency)", func(t *testing.T) {
		pkgs, err := analyzer.Load("../testdata/invalid", "internal/...")
		if err != nil {
			t.Fatal(err)
		}
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
		pkgs, err := analyzer.Load("../testdata/invalid", "internal/...")
		if err != nil {
			t.Fatal(err)
		}
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
}
