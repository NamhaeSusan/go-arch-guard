package interfaces_test

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/analyzer"
	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/rules/interfaces"
)

func TestPatternDetectsCoreViolations(t *testing.T) {
	root := writeFixture(t, "example.com/pattern-core", map[string]string{
		"internal/store/store.go": `package store

type Store interface {
	Get(id string) string
}

type Audit interface {
	Log(message string)
}

type StoreImpl struct{}

func (s *StoreImpl) Get(id string) string { return "" }

func NewStore() Store { return &StoreImpl{} }

func New() *StoreImpl { return &StoreImpl{} }
`,
	})

	violations := interfaces.NewPattern().Check(loadContext(t, root, flatArchitecture(), "example.com/pattern-core"))

	assertHasRule(t, violations, "interface.single-per-package")
	assertHasRule(t, violations, "interface.exported-impl")
	assertHasRule(t, violations, "interface.constructor-name")
	assertHasRule(t, violations, "interface.constructor-returns-interface")
}

func TestPatternMaxMethodsDisabledByDefault(t *testing.T) {
	root := writeFixture(t, "example.com/pattern-max-default", map[string]string{
		"internal/store/store.go": `package store

type Store interface {
	A()
	B()
	C()
	D()
	E()
	F()
	G()
	H()
	I()
	J()
	K()
}

type store struct{}

func New() Store { return &store{} }
`,
	})

	violations := interfaces.NewPattern().Check(loadContext(t, root, flatArchitecture(), "example.com/pattern-max-default"))

	assertLacksRule(t, violations, "interface.too-many-methods")
}

func TestPatternMaxMethodsOptIn(t *testing.T) {
	root := writeFixture(t, "example.com/pattern-max-opt-in", map[string]string{
		"internal/store/store.go": `package store

type Store interface {
	A()
	B()
}

type store struct{}

func New() Store { return &store{} }
`,
	})

	violations := interfaces.NewPattern(interfaces.WithMaxMethods(1)).Check(loadContext(t, root, flatArchitecture(), "example.com/pattern-max-opt-in"))

	assertHasRule(t, violations, "interface.too-many-methods")
}

func TestPatternHonorsInterfacePatternExclude(t *testing.T) {
	root := writeFixture(t, "example.com/pattern-exclude", map[string]string{
		"internal/model/model.go": `package model

type Model interface {
	ID() string
}

type ModelImpl struct{}

func (m *ModelImpl) ID() string { return "" }

func NewModel() Model { return &ModelImpl{} }
`,
	})
	arch := flatArchitecture()
	arch.Structure.InterfacePatternExclude = map[string]bool{"model": true}

	violations := interfaces.NewPattern().Check(loadContext(t, root, arch, "example.com/pattern-exclude"))

	if len(violations) != 0 {
		t.Fatalf("violations = %v, want none", violations)
	}
}

func TestPatternWithSeverity(t *testing.T) {
	rule := interfaces.NewPattern(interfaces.WithSeverity(core.Warning))
	spec := rule.Spec()

	if spec.DefaultSeverity != core.Warning {
		t.Fatalf("DefaultSeverity = %v, want Warning", spec.DefaultSeverity)
	}
	for _, v := range spec.Violations {
		if v.DefaultSeverity != core.Warning {
			t.Fatalf("%s DefaultSeverity = %v, want Warning", v.ID, v.DefaultSeverity)
		}
	}
}

func writeFixture(t *testing.T, module string, files map[string]string) string {
	t.Helper()

	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.21\n")
	for name, content := range files {
		writeFile(t, filepath.Join(root, name), content)
	}
	return root
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func loadContext(t *testing.T, root string, arch core.Architecture, module string) *core.Context {
	t.Helper()

	pkgs, err := analyzer.Load(root, "...")
	if err != nil {
		t.Fatalf("load packages: %v", err)
	}
	return core.NewContext(pkgs, module, root, arch, nil)
}

func flatArchitecture() core.Architecture {
	return core.Architecture{
		Layers: core.LayerModel{
			Sublayers: []string{"store", "model"},
			Direction: map[string][]string{
				"store": nil,
				"model": nil,
			},
		},
		Layout: core.LayoutModel{
			SharedDir: "pkg",
		},
	}
}

func domainArchitecture() core.Architecture {
	return core.Architecture{
		Layers: core.LayerModel{
			Sublayers: []string{"handler", "core/model", "orchestration", "wire"},
			Direction: map[string][]string{
				"handler":       {"core/model"},
				"core/model":    nil,
				"orchestration": {"core/model"},
				"wire":          nil,
			},
		},
		Layout: core.LayoutModel{
			DomainDir:        "domain",
			OrchestrationDir: "orchestration",
		},
	}
}

func assertHasRule(t *testing.T, violations []core.Violation, id string) {
	t.Helper()

	if slices.ContainsFunc(violations, func(v core.Violation) bool { return v.Rule == id }) {
		return
	}
	t.Fatalf("missing violation %q in %v", id, violations)
}

func assertLacksRule(t *testing.T, violations []core.Violation, id string) {
	t.Helper()

	if slices.ContainsFunc(violations, func(v core.Violation) bool { return v.Rule == id }) {
		t.Fatalf("unexpected violation %q in %v", id, violations)
	}
}
