package naming_test

import (
	"strings"
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/rules/naming"
)

func TestRepoFileInterfaceSpec(t *testing.T) {
	spec := naming.NewRepoFileInterface(naming.WithSeverity(core.Warning)).Spec()

	if spec.ID != "naming.repo-file-interface" {
		t.Fatalf("ID = %q, want naming.repo-file-interface", spec.ID)
	}
	if spec.DefaultSeverity != core.Warning {
		t.Fatalf("DefaultSeverity = %v, want Warning", spec.DefaultSeverity)
	}
	want := []string{"structure.repo-file-interface", "structure.repo-file-extra-interface", "structure.interface-placement"}
	got := spec.ViolationIDs()
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("ViolationIDs = %v, want %v", got, want)
		}
	}
}

func TestRepoFileInterfaceFlagsRepoFilenameContract(t *testing.T) {
	ctx := tempContext(t, map[string]string{
		"internal/domain/order/core/repo/order.go": "package repo\n\nfunc FindOrder() {}\n",
	}, dddArch())

	got := naming.NewRepoFileInterface().Check(ctx)
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1: %+v", len(got), got)
	}
	if got[0].Rule != "structure.repo-file-interface" || got[0].DefaultSeverity != core.Error {
		t.Fatalf("violation = %+v, want repo-file-interface Error", got[0])
	}
}

func TestRepoFileInterfaceFlagsExtraInterfaces(t *testing.T) {
	ctx := tempContext(t, map[string]string{
		"internal/domain/order/core/repo/review.go": "package repo\n\ntype Review interface { Find() }\ntype Helper interface { Assist() }\n",
	}, dddArch())

	got := naming.NewRepoFileInterface().Check(ctx)
	if len(got) != 1 || got[0].Rule != "structure.repo-file-extra-interface" {
		t.Fatalf("got %+v, want one extra-interface violation", got)
	}
}

func TestRepoFileInterfaceFlagsRepositoryPortOutsidePortLayer(t *testing.T) {
	got := naming.NewRepoFileInterface().Check(invalidContext(t, nil))

	var direct, alias bool
	for _, v := range got {
		if v.Rule != "structure.interface-placement" {
			continue
		}
		if strings.Contains(v.Message, `"OrderRepository"`) {
			direct = true
		}
		if strings.Contains(v.Message, `"OrderRepo"`) {
			alias = true
		}
		if strings.Contains(v.Message, `"Service"`) || strings.Contains(v.Message, `"AdminOps"`) {
			t.Fatalf("consumer-defined interface should not be flagged: %+v", v)
		}
	}
	if !direct || !alias {
		t.Fatalf("direct=%v alias=%v, want both true; got %+v", direct, alias, got)
	}
}
