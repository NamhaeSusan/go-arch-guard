package structural_test

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/analyzer"
	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/rules/structural"
	"golang.org/x/tools/go/packages"
)

var (
	repoIfaceInvalidOnce sync.Once
	repoIfaceInvalidPkgs []*packages.Package
	repoIfaceInvalidErr  error
)

func loadInvalidForRepoIface(t *testing.T) []*packages.Package {
	t.Helper()
	repoIfaceInvalidOnce.Do(func() {
		repoIfaceInvalidPkgs, repoIfaceInvalidErr = analyzer.Load("../../testdata/invalid", "internal/...")
	})
	if repoIfaceInvalidErr != nil {
		t.Fatal(repoIfaceInvalidErr)
	}
	return repoIfaceInvalidPkgs
}

func invalidContextForRepoIface(t *testing.T) *core.Context {
	t.Helper()
	return core.NewContext(loadInvalidForRepoIface(t), "github.com/kimtaeyun/testproject-dc-invalid", "../../testdata/invalid", dddArch(), nil)
}

func tempContextForRepoIface(t *testing.T, files map[string]string, arch core.Architecture) *core.Context {
	t.Helper()
	root := t.TempDir()
	writeRepoIfaceFile(t, filepath.Join(root, "go.mod"), "module example.com/structuraltest\n\ngo 1.25.0\n")
	for name, content := range files {
		writeRepoIfaceFile(t, filepath.Join(root, name), content)
	}
	pkgs, err := analyzer.Load(root, "internal/...")
	if err != nil {
		t.Fatal(err)
	}
	return core.NewContext(pkgs, "example.com/structuraltest", root, arch, nil)
}

func writeRepoIfaceFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestRepoFileInterfaceSpec(t *testing.T) {
	spec := structural.NewRepoFileInterface(structural.WithSeverity(core.Warning)).Spec()

	if spec.ID != "structural.repo-file-interface" {
		t.Fatalf("ID = %q, want structural.repo-file-interface", spec.ID)
	}
	if spec.DefaultSeverity != core.Warning {
		t.Fatalf("DefaultSeverity = %v, want Warning", spec.DefaultSeverity)
	}
	want := []string{"structural.repo-file-interface-missing", "structural.repo-file-extra-interface", "structural.interface-placement"}
	got := spec.ViolationIDs()
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("ViolationIDs = %v, want %v", got, want)
		}
	}
}

func TestRepoFileInterfaceFlagsRepoFilenameContract(t *testing.T) {
	ctx := tempContextForRepoIface(t, map[string]string{
		"internal/domain/order/core/repo/order.go": "package repo\n\nfunc FindOrder() {}\n",
	}, dddArch())

	got := structural.NewRepoFileInterface().Check(ctx)
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1: %+v", len(got), got)
	}
	if got[0].Rule != "structural.repo-file-interface-missing" || got[0].DefaultSeverity != core.Error {
		t.Fatalf("violation = %+v, want repo-file-interface Error", got[0])
	}
}

func TestRepoFileInterfaceFlagsExtraInterfaces(t *testing.T) {
	ctx := tempContextForRepoIface(t, map[string]string{
		"internal/domain/order/core/repo/review.go": "package repo\n\ntype Review interface { Find() }\ntype Helper interface { Assist() }\n",
	}, dddArch())

	got := structural.NewRepoFileInterface().Check(ctx)
	if len(got) != 1 || got[0].Rule != "structural.repo-file-extra-interface" {
		t.Fatalf("got %+v, want one extra-interface violation", got)
	}
}

func TestRepoFileInterfaceExtraInterfaceFixUsesSnakeCase(t *testing.T) {
	ctx := tempContextForRepoIface(t, map[string]string{
		"internal/domain/order/core/repo/order.go": "package repo\n\ntype Order interface { Find() }\ntype UserRepository interface { Save() }\n",
	}, dddArch())

	got := structural.NewRepoFileInterface().Check(ctx)
	var found bool
	for _, v := range got {
		if v.Rule != "structural.repo-file-extra-interface" {
			continue
		}
		found = true
		if !strings.Contains(v.Fix, "user_repository.go") {
			t.Fatalf("Fix = %q, want it to suggest user_repository.go", v.Fix)
		}
		if strings.Contains(v.Fix, "userrepository.go") {
			t.Fatalf("Fix = %q, must not suggest userrepository.go", v.Fix)
		}
	}
	if !found {
		t.Fatalf("expected extra-interface violation, got %+v", got)
	}
}

func TestRepoFileInterfaceNoPortSublayerEmitsMetaDisabledByConfig(t *testing.T) {
	arch := dddArch()
	// Strip every port-shaped sublayer (named "repo"/"gateway" or listed in PortLayers)
	// so analysisutil.HasPortSublayer returns false.
	arch.Layers.PortLayers = nil
	arch.Layers.Sublayers = []string{"handler", "app", "core", "core/model", "core/svc", "event", "infra"}
	ctx := core.NewContext(nil, "github.com/example/app", "../../testdata/valid", arch, nil)

	got := structural.NewRepoFileInterface().Check(ctx)
	if len(got) != 1 || got[0].Rule != "meta.rule-disabled-by-config" {
		t.Fatalf("expected exactly 1 meta.rule-disabled-by-config violation, got %+v", got)
	}
	if !strings.Contains(got[0].Message, "PortLayers is empty") {
		t.Fatalf("meta message should mention PortLayers, got %q", got[0].Message)
	}
}

func TestRepoFileInterfaceWithRepoPortSuffixesOverrideMatchesAlternateVocabulary(t *testing.T) {
	// Default suffixes are ["Repository", "Repo"]. With WithRepoPortSuffixes,
	// the rule should match a project that names its ports *Gateway instead.
	ctx := tempContextForRepoIface(t, map[string]string{
		"internal/domain/order/core/svc/svc.go": `package svc

type OrderGateway interface {
	Find() string
}
`,
	}, dddArch())

	got := structural.NewRepoFileInterface(structural.WithRepoPortSuffixes("Gateway")).Check(ctx)

	var found bool
	for _, v := range got {
		if v.Rule == "structural.interface-placement" && strings.Contains(v.Message, `"OrderGateway"`) {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected interface-placement violation for OrderGateway, got %+v", got)
	}
}

func TestRepoFileInterfaceWithRepoPortSuffixesEmptyFallsBackToDefault(t *testing.T) {
	// Empty suffix slice should fall back to the default ["Repository", "Repo"].
	ctx := tempContextForRepoIface(t, map[string]string{
		"internal/domain/order/core/svc/svc.go": `package svc

type OrderRepository interface {
	Find() string
}
`,
	}, dddArch())

	got := structural.NewRepoFileInterface(structural.WithRepoPortSuffixes()).Check(ctx)

	var found bool
	for _, v := range got {
		if v.Rule == "structural.interface-placement" && strings.Contains(v.Message, `"OrderRepository"`) {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected interface-placement violation for OrderRepository (default suffixes), got %+v", got)
	}
}

func TestRepoFileInterfaceFlagsRepositoryPortOutsidePortLayer(t *testing.T) {
	got := structural.NewRepoFileInterface().Check(invalidContextForRepoIface(t))

	var direct, alias bool
	for _, v := range got {
		if v.Rule != "structural.interface-placement" {
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
