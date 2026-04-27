package testpolicy_test

import (
	"strings"
	"sync"
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/analyzer"
	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/rules/testpolicy"
	"golang.org/x/tools/go/packages"
)

var (
	invalidOnce sync.Once
	invalidPkgs []*packages.Package
	invalidErr  error
)

func loadInvalid(t *testing.T) []*packages.Package {
	t.Helper()
	invalidOnce.Do(func() {
		invalidPkgs, invalidErr = analyzer.Load("../../testdata/invalid", "internal/...")
	})
	if invalidErr != nil {
		t.Fatal(invalidErr)
	}
	return invalidPkgs
}

func invalidContext(t *testing.T, exclude []string) *core.Context {
	t.Helper()
	return core.NewContext(loadInvalid(t), "github.com/kimtaeyun/testproject-dc-invalid", "../../testdata/invalid", core.Architecture{}, exclude)
}

func TestNoHandMockSpec(t *testing.T) {
	spec := testpolicy.NewNoHandMock(testpolicy.WithSeverity(core.Error)).Spec()

	if spec.ID != "testpolicy.no-handmock" {
		t.Fatalf("ID = %q, want testpolicy.no-handmock", spec.ID)
	}
	if spec.DefaultSeverity != core.Error {
		t.Fatalf("DefaultSeverity = %v, want Error", spec.DefaultSeverity)
	}
}

func TestNoHandMockFlagsMockStructsWithMethods(t *testing.T) {
	got := testpolicy.NewNoHandMock().Check(invalidContext(t, nil))

	want := map[string]bool{"mockOrderRepo": false, "fakeNotifier": false}
	for _, v := range got {
		for name := range want {
			if v.Rule == "testpolicy.no-handmock" && strings.Contains(v.Message, name) {
				want[name] = true
			}
		}
	}
	for name, found := range want {
		if !found {
			t.Fatalf("expected testpolicy.no-handmock violation for %s; got %+v", name, got)
		}
	}
}

func TestNoHandMockRespectsExclude(t *testing.T) {
	got := testpolicy.NewNoHandMock().Check(invalidContext(t, []string{"internal/domain/order/app/..."}))
	for _, v := range got {
		if v.Rule == "testpolicy.no-handmock" {
			t.Fatalf("excluded handmock violation returned: %+v", v)
		}
	}
}
