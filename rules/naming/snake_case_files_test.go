package naming_test

import (
	"strings"
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/rules/naming"
)

func TestSnakeCaseFilesSpec(t *testing.T) {
	spec := naming.NewSnakeCaseFiles(naming.WithSeverity(core.Error)).Spec()

	if spec.ID != "naming.snake-case-file" {
		t.Fatalf("ID = %q, want naming.snake-case-file", spec.ID)
	}
	if spec.DefaultSeverity != core.Error {
		t.Fatalf("DefaultSeverity = %v, want Error", spec.DefaultSeverity)
	}
}

func TestSnakeCaseFilesFlagsNonSnakeCaseGoFile(t *testing.T) {
	ctx := tempContext(t, map[string]string{
		"internal/domain/order/app/createOrder.go": "package app\n\ntype Request struct{}\n",
	}, dddArch())

	got := naming.NewSnakeCaseFiles().Check(ctx)
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1: %+v", len(got), got)
	}
	if got[0].Rule != "naming.snake-case-file" || !strings.Contains(got[0].Fix, "create_order.go") {
		t.Fatalf("violation = %+v, want snake-case fix", got[0])
	}
}

func TestSnakeCaseFilesSkipsExcludedPath(t *testing.T) {
	ctx := core.NewContext(loadInvalid(t), "github.com/kimtaeyun/testproject-dc-invalid", "../../testdata/invalid", dddArch(), []string{"internal/domain/order/handler/..."})

	got := naming.NewSnakeCaseFiles().Check(ctx)
	for _, v := range got {
		if strings.Contains(v.File, "internal/domain/order/handler/") {
			t.Fatalf("excluded handler file was flagged: %+v", v)
		}
	}
}
