package rules_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/rules"
)

func TestAnalyzeBlastRadius(t *testing.T) {
	t.Run("returns no violations for small project", func(t *testing.T) {
		pkgs := loadValid(t)
		violations := rules.AnalyzeBlastRadius(pkgs, "github.com/kimtaeyun/testproject-dc", "../testdata/valid")
		for _, v := range violations {
			if v.EffectiveSeverity == rules.Error {
				t.Errorf("unexpected error violation: %s", v.String())
			}
		}
	})
}

func TestAnalyzeBlastRadius_DetectsOutlier(t *testing.T) {
	pkgs := loadBlast(t)
	violations := rules.AnalyzeBlastRadius(pkgs, "github.com/kimtaeyun/testproject-blast", "../testdata/blast")

	if len(violations) == 0 {
		t.Fatal("expected at least one blast.high-coupling violation for the hub package")
	}

	foundPkg := false
	for _, v := range violations {
		if v.Rule != "blast.high-coupling" {
			t.Errorf("unexpected rule: %s", v.Rule)
		}
		if v.EffectiveSeverity != rules.Warning {
			t.Errorf("expected Warning severity, got %v", v.EffectiveSeverity)
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
		if v.EffectiveSeverity == rules.Error {
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

// TestAnalyzeBlastRadius_IQRZeroSingleOutlier is a regression test for the bug where
// iqr == 0 caused an early return nil, silently missing a single outlier package.
func TestAnalyzeBlastRadius_IQRZeroSingleOutlier(t *testing.T) {
	root := t.TempDir()
	mod := "example.com/iqr-zero"

	writeTestFile(t, filepath.Join(root, "go.mod"), "module "+mod+"\n\ngo 1.21\n")

	// hub is imported by a, b, c, d — 4 transitive dependents; leaves have 0
	writeTestFile(t, filepath.Join(root, "internal", "hub", "hub.go"),
		"package hub\n\nfunc Hub() {}\n")
	for _, leaf := range []string{"a", "b", "c", "d"} {
		writeTestFile(t, filepath.Join(root, "internal", leaf, leaf+".go"),
			"package "+leaf+"\n\nimport _ \""+mod+"/internal/hub\"\n")
	}
	// fifth leaf has no imports — leaf e is isolated
	writeTestFile(t, filepath.Join(root, "internal", "e", "e.go"),
		"package e\n\nfunc E() {}\n")

	pkgs := loadTestPackages(t, root)
	violations := rules.AnalyzeBlastRadius(pkgs, mod, root)

	found := false
	for _, v := range violations {
		if v.Rule == "blast.high-coupling" && strings.Contains(v.File, "hub") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected blast.high-coupling violation for hub (IQR=0 single-outlier case)")
		for _, v := range violations {
			t.Log(v.String())
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
		if v.EffectiveSeverity != rules.Error {
			t.Errorf("expected Error severity, got %v", v.EffectiveSeverity)
		}
	}
}
