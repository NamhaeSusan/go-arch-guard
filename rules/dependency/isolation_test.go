package dependency_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/analyzer"
	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/rules/dependency"
)

func TestIsolationSpec(t *testing.T) {
	rule := dependency.NewIsolation(dependency.WithSeverity(core.Warning))

	spec := rule.Spec()
	if spec.ID != "dependency.isolation" {
		t.Fatalf("Spec().ID = %q, want dependency.isolation", spec.ID)
	}
	if spec.DefaultSeverity != core.Warning {
		t.Fatalf("Spec().DefaultSeverity = %v, want %v", spec.DefaultSeverity, core.Warning)
	}

	for _, id := range []string{
		"isolation.cross-domain",
		"isolation.cmd-deep-import",
		"isolation.orchestration-deep-import",
		"isolation.pkg-imports-domain",
		"isolation.pkg-imports-orchestration",
		"isolation.domain-imports-orchestration",
		"isolation.stray-imports-orchestration",
		"isolation.stray-imports-domain",
		"isolation.transport-imports-domain",
		"isolation.transport-imports-orchestration",
		"isolation.transport-imports-unclassified",
	} {
		if !slices.Contains(spec.ViolationIDs(), id) {
			t.Fatalf("Spec().ViolationIDs() missing %q", id)
		}
	}
}

func TestIsolationValidProject(t *testing.T) {
	ctx := loadContext(t, "../../testdata/valid", "github.com/kimtaeyun/testproject-dc", dddArchitecture(), "internal/...", "cmd/...")

	violations := dependency.NewIsolation().Check(ctx)
	if len(violations) > 0 {
		for _, v := range violations {
			t.Log(v.String())
		}
		t.Fatalf("expected no violations, got %d", len(violations))
	}
}

func TestIsolationInvalidProject(t *testing.T) {
	ctx := loadContext(t, "../../testdata/invalid", "github.com/kimtaeyun/testproject-dc-invalid", dddArchitecture(), "internal/...", "cmd/...")

	violations := dependency.NewIsolation().Check(ctx)

	for _, id := range []string{
		"isolation.cross-domain",
		"isolation.cmd-deep-import",
		"isolation.orchestration-deep-import",
		"isolation.pkg-imports-domain",
		"isolation.pkg-imports-orchestration",
		"isolation.domain-imports-orchestration",
		"isolation.stray-imports-orchestration",
		"isolation.stray-imports-domain",
	} {
		assertHasRule(t, violations, id)
	}
}

func TestIsolationDDDAppServerTransport(t *testing.T) {
	ctx := loadContext(t, "../../testdata/ddd-appserver", "github.com/kimtaeyun/testproject-ddd-appserver", dddArchitecture(), "internal/...")

	violations := dependency.NewIsolation().Check(ctx)

	assertHasRule(t, violations, "isolation.transport-imports-domain")
}

func TestIsolationFlatLayoutEmitsMetaWarning(t *testing.T) {
	ctx := loadFlatLayoutContext(t)
	violations := dependency.NewIsolation().Check(ctx)
	assertExactlyOneMetaLayoutNotSupported(t, violations, "dependency.isolation")
}

func TestIsolationDDDProjectHasNoMetaLayoutWarning(t *testing.T) {
	ctx := loadContext(t, "../../testdata/valid", "github.com/kimtaeyun/testproject-dc", dddArchitecture(), "internal/...")
	for _, v := range dependency.NewIsolation().Check(ctx) {
		if v.Rule == "meta.layout-not-supported" {
			t.Fatalf("internal/-based project must not emit meta.layout-not-supported: %s", v.String())
		}
	}
}

func TestIsolationExclude(t *testing.T) {
	ctx := loadContextWithExclude(t,
		"../../testdata/invalid",
		"github.com/kimtaeyun/testproject-dc-invalid",
		dddArchitecture(),
		[]string{"internal/config/..."},
		"internal/...",
	)

	violations := dependency.NewIsolation().Check(ctx)
	for _, v := range violations {
		if v.File == "internal/config/config.go" ||
			v.File == "internal/config/domain_alias.go" ||
			v.File == "internal/config/orchestration.go" {
			t.Fatalf("expected config package to be excluded, got %s", v.String())
		}
	}
}

func loadFlatLayoutContext(t *testing.T) *core.Context {
	t.Helper()
	return loadContext(t, "../../testdata/flat", "github.com/kimtaeyun/testproject-flat", dddArchitecture(), "...")
}

func assertExactlyOneMetaLayoutNotSupported(t *testing.T, violations []core.Violation, ruleID string) {
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

func loadContext(t *testing.T, root, module string, arch core.Architecture, patterns ...string) *core.Context {
	t.Helper()
	return loadContextWithExclude(t, root, module, arch, nil, patterns...)
}

func loadContextWithExclude(t *testing.T, root, module string, arch core.Architecture, exclude []string, patterns ...string) *core.Context {
	t.Helper()
	pkgs, err := analyzer.Load(root, patterns...)
	if err != nil {
		t.Fatal(err)
	}
	if len(pkgs) == 0 {
		t.Fatal("no packages loaded")
	}
	return core.NewContext(pkgs, module, root, arch, exclude)
}

func assertHasRule(t *testing.T, violations []core.Violation, rule string) {
	t.Helper()
	for _, v := range violations {
		if v.Rule == rule {
			return
		}
	}
	t.Fatalf("expected rule %q, got %v", rule, ruleIDs(violations))
}

func ruleIDs(violations []core.Violation) []string {
	seen := make(map[string]bool)
	for _, v := range violations {
		seen[v.Rule] = true
	}
	ids := make([]string, 0, len(seen))
	for id := range seen {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	return ids
}

func dddArchitecture() core.Architecture {
	return core.Architecture{
		Layers: core.LayerModel{
			Sublayers: []string{
				"handler", "app", "core", "core/model",
				"core/repo", "core/svc", "event", "infra",
			},
			Direction: map[string][]string{
				"handler":    {"app"},
				"app":        {"core/model", "core/repo", "core/svc", "event"},
				"core":       {"core/model"},
				"core/model": {},
				"core/repo":  {"core/model"},
				"core/svc":   {"core/model"},
				"event":      {"core/model"},
				"infra":      {"core/repo", "core/model", "event"},
			},
			PortLayers:     []string{"core/repo"},
			ContractLayers: []string{"core/repo", "core/svc"},
			PkgRestricted: map[string]bool{
				"core": true, "core/model": true,
				"core/repo": true, "core/svc": true, "event": true,
			},
			InternalTopLevel: map[string]bool{
				"domain": true, "orchestration": true, "pkg": true,
				"app": true, "server": true,
			},
			LayerDirNames: map[string]bool{
				"handler": true, "app": true, "core": true,
				"model": true, "repo": true, "svc": true,
				"event": true, "infra": true,
				"service": true, "controller": true,
				"entity": true, "store": true, "persistence": true,
				"domain": true,
			},
		},
		Layout: core.LayoutModel{
			DomainDir:        "domain",
			OrchestrationDir: "orchestration",
			SharedDir:        "pkg",
			AppDir:           "app",
			ServerDir:        "server",
		},
		Naming: core.NamingPolicy{
			BannedPkgNames: []string{"util", "common", "misc", "helper", "shared", "services"},
			LegacyPkgNames: []string{"router", "bootstrap"},
			AliasFileName:  "alias.go",
		},
		Structure: core.StructurePolicy{
			RequireAlias:            true,
			RequireModel:            true,
			ModelPath:               "core/model",
			DTOAllowedLayers:        []string{"handler", "app"},
			InterfacePatternExclude: map[string]bool{"handler": true, "app": true, "core/model": true, "core/repo": true, "event": true},
		},
	}
}
