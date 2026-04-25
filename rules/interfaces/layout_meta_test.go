package interfaces_test

import (
	"strings"
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/analyzer"
	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/rules/interfaces"
)

func loadFlatContext(t *testing.T) *core.Context {
	t.Helper()
	pkgs, err := analyzer.Load("../../testdata/flat", "...")
	if err != nil {
		t.Fatalf("load packages: %v", err)
	}
	// domainArchitecture() is used deliberately: hasInternalPackages
	// short-circuits before DomainDir is read, so the arch config is
	// irrelevant for the meta path. Using a non-empty DomainDir also
	// proves the meta guard fires before the DomainDir-empty guard.
	return core.NewContext(pkgs, "github.com/kimtaeyun/testproject-flat", "../../testdata/flat", domainArchitecture(), nil)
}

func assertExactlyOneMeta(t *testing.T, violations []core.Violation, ruleID string) {
	t.Helper()
	var count int
	for _, v := range violations {
		if v.Rule == "meta.layout-not-supported" {
			count++
			if !strings.Contains(v.Message, ruleID) {
				t.Fatalf("meta message should mention %q, got %q", ruleID, v.Message)
			}
		}
	}
	if count != 1 {
		t.Fatalf("expected exactly 1 meta.layout-not-supported, got %d: %+v", count, violations)
	}
}

func TestPatternFlatLayoutEmitsMetaWarning(t *testing.T) {
	violations := interfaces.NewPattern().Check(loadFlatContext(t))
	assertExactlyOneMeta(t, violations, "interfaces.pattern")
}

func TestCrossDomainAnonymousFlatLayoutEmitsMetaWarning(t *testing.T) {
	violations := interfaces.NewCrossDomainAnonymous().Check(loadFlatContext(t))
	assertExactlyOneMeta(t, violations, "interfaces.cross-domain-anonymous")
}

func TestContainerFlatLayoutDoesNotEmitMeta(t *testing.T) {
	// Container is layout-agnostic and must not emit meta warnings.
	violations := interfaces.NewContainer().Check(loadFlatContext(t))
	for _, v := range violations {
		if v.Rule == "meta.layout-not-supported" {
			t.Fatalf("Container must not emit meta.layout-not-supported: %+v", v)
		}
	}
}
