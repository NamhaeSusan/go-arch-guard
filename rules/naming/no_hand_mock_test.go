package naming_test

import (
	"strings"
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/rules/naming"
)

func TestNoHandMockSpec(t *testing.T) {
	spec := naming.NewNoHandMock(naming.WithSeverity(core.Error)).Spec()

	if spec.ID != "testing.no-handmock" {
		t.Fatalf("ID = %q, want testing.no-handmock", spec.ID)
	}
	if spec.DefaultSeverity != core.Error {
		t.Fatalf("DefaultSeverity = %v, want Error", spec.DefaultSeverity)
	}
}

func TestNoHandMockFlagsMockStructsWithMethods(t *testing.T) {
	got := naming.NewNoHandMock().Check(invalidContext(t, nil))

	want := map[string]bool{"mockOrderRepo": false, "fakeNotifier": false}
	for _, v := range got {
		for name := range want {
			if v.Rule == "testing.no-handmock" && strings.Contains(v.Message, name) {
				want[name] = true
			}
		}
	}
	for name, found := range want {
		if !found {
			t.Fatalf("expected testing.no-handmock violation for %s; got %+v", name, got)
		}
	}
}

func TestNoHandMockRespectsExclude(t *testing.T) {
	got := naming.NewNoHandMock().Check(invalidContext(t, []string{"internal/domain/order/app/..."}))
	for _, v := range got {
		if v.Rule == "testing.no-handmock" {
			t.Fatalf("excluded handmock violation returned: %+v", v)
		}
	}
}
