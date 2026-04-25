package dependency_test

import (
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/rules/dependency"
)

func TestIsolationCustomInternalRootRecognizesProject(t *testing.T) {
	arch := dddArchitecture()
	arch.Layout.InternalRoot = "packages"
	ctx := loadContext(t, "../../testdata/custom_root", "github.com/kimtaeyun/testproject-customroot", arch, "...")

	violations := dependency.NewIsolation().Check(ctx)
	for _, v := range violations {
		if v.Rule == "meta.layout-not-supported" {
			t.Fatalf("custom InternalRoot=packages must not emit layout-not-supported: %s", v.String())
		}
	}
}

func TestIsolationCustomInternalRootEmitsMetaForMissingRoot(t *testing.T) {
	arch := dddArchitecture()
	arch.Layout.InternalRoot = "src" // src/ does not exist in the fixture
	ctx := loadContext(t, "../../testdata/custom_root", "github.com/kimtaeyun/testproject-customroot", arch, "...")

	violations := dependency.NewIsolation().Check(ctx)
	var meta int
	for _, v := range violations {
		if v.Rule == "meta.layout-not-supported" {
			meta++
		}
	}
	if meta != 1 {
		t.Fatalf("expected exactly 1 meta.layout-not-supported when InternalRoot points to a missing dir, got %d: %+v", meta, violations)
	}
}

