package rules_test

import (
	"strings"
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/rules"
)

func TestAnalyzeBlastRadius(t *testing.T) {
	t.Run("returns no violations for small project", func(t *testing.T) {
		pkgs := loadValid(t)
		violations := rules.AnalyzeBlastRadius(pkgs, "github.com/kimtaeyun/testproject-dc", "../testdata/valid")
		for _, v := range violations {
			if v.Severity == rules.Error {
				t.Errorf("unexpected error violation: %s", v.String())
			}
		}
	})
}

func TestAnalyzeBlastRadius_DetectsOutlier(t *testing.T) {
	pkgs := loadBlast(t)
	violations := rules.AnalyzeBlastRadius(pkgs, "github.com/kimtaeyun/testproject-blast", "../testdata/blast")

	if len(violations) == 0 {
		t.Fatal("expected at least one blast-radius violation for the hub package")
	}

	foundPkg := false
	for _, v := range violations {
		if v.Rule != "blast-radius.high-coupling" {
			t.Errorf("unexpected rule: %s", v.Rule)
		}
		if v.Severity != rules.Warning {
			t.Errorf("expected Warning severity, got %v", v.Severity)
		}
		if strings.Contains(v.File, "internal/pkg") {
			foundPkg = true
		}
		t.Logf("violation: %s", v.String())
	}
	if !foundPkg {
		t.Error("expected internal/pkg to be flagged as high-coupling outlier")
	}
}

func TestAnalyzeBlastRadius_SkipsTooFewPackages(t *testing.T) {
	pkgs := loadValid(t)
	violations := rules.AnalyzeBlastRadius(pkgs, "github.com/kimtaeyun/testproject-dc", "../testdata/valid")
	for _, v := range violations {
		if v.Severity == rules.Error {
			t.Errorf("unexpected error-severity violation: %s", v.String())
		}
	}
}

func TestAnalyzeBlastRadius_RespectsExclude(t *testing.T) {
	pkgs := loadBlast(t)
	violations := rules.AnalyzeBlastRadius(pkgs,
		"github.com/kimtaeyun/testproject-blast", "../testdata/blast",
		rules.WithExclude("internal/pkg/..."),
	)
	for _, v := range violations {
		if strings.Contains(v.File, "internal/pkg") {
			t.Error("excluded package should not appear in violations")
		}
	}
}

func TestAnalyzeBlastRadius_RespectsSeverity(t *testing.T) {
	pkgs := loadBlast(t)
	violations := rules.AnalyzeBlastRadius(pkgs,
		"github.com/kimtaeyun/testproject-blast", "../testdata/blast",
		rules.WithSeverity(rules.Error),
	)
	for _, v := range violations {
		if v.Severity != rules.Error {
			t.Errorf("expected Error severity, got %v", v.Severity)
		}
	}
}
