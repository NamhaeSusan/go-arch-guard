package types_test

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/analyzer"
	"github.com/NamhaeSusan/go-arch-guard/core"
	types "github.com/NamhaeSusan/go-arch-guard/rules/types"
	"golang.org/x/tools/go/packages"
)

var (
	typesOnce sync.Once
	typesPkgs []*packages.Package
	typesErr  error
)

func loadTypesFixture(t *testing.T) []*packages.Package {
	t.Helper()
	typesOnce.Do(func() {
		typesPkgs, typesErr = analyzer.Load("../../testdata/types", "internal/...", "mocks/...")
	})
	if typesErr != nil {
		t.Fatal(typesErr)
	}
	return typesPkgs
}

func newFixtureContext(t *testing.T, arch core.Architecture, exclude []string) *core.Context {
	t.Helper()
	return core.NewContext(loadTypesFixture(t), "github.com/kimtaeyun/testproject-types", "../../testdata/types", arch, exclude)
}

func TestNoSetterSpec(t *testing.T) {
	spec := types.NewNoSetter(types.WithSeverity(core.Error)).Spec()

	if spec.ID != "types.no-setter" {
		t.Fatalf("ID = %q, want types.no-setter", spec.ID)
	}
	if spec.DefaultSeverity != core.Error {
		t.Fatalf("DefaultSeverity = %v, want Error", spec.DefaultSeverity)
	}
	if len(spec.Violations) != 1 || spec.Violations[0].ID != "setter.forbidden" {
		t.Fatalf("Violations = %+v, want setter.forbidden", spec.Violations)
	}
}

func TestNoSetterFlagsClassicSetters(t *testing.T) {
	ctx := newFixtureContext(t, core.Architecture{}, nil)
	got := types.NewNoSetter().Check(ctx)

	var setters []core.Violation
	for _, v := range got {
		if v.Rule == "setter.forbidden" && strings.Contains(v.File, "internal/model/order.go") {
			setters = append(setters, v)
		}
	}
	if len(setters) != 3 {
		t.Fatalf("expected exactly 3 setter violations from order.go, got %d: %+v", len(setters), setters)
	}
	for _, v := range setters {
		if v.DefaultSeverity != core.Warning || v.EffectiveSeverity != core.Warning {
			t.Fatalf("setter severity = default %v effective %v, want Warning/Warning", v.DefaultSeverity, v.EffectiveSeverity)
		}
	}
}

func TestNoSetterHandlesGenericReceivers(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/gen\n\ngo 1.25.0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "internal", "model"), 0o755); err != nil {
		t.Fatal(err)
	}
	src := `package model

type Stack[T any] struct{ items []T }
func (s *Stack[T]) SetTop(v T) { s.items[len(s.items)-1] = v }

type Builder[T any] struct{ v T }
func (b *Builder[T]) SetValue(v T) *Builder[T] { b.v = v; return b }

type Pair[K comparable, V any] struct{ k K; v V }
func (p *Pair[K, V]) SetKey(k K) { p.k = k }
`
	if err := os.WriteFile(filepath.Join(root, "internal", "model", "gen.go"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	pkgs, err := analyzer.Load(root, "internal/...")
	if err != nil {
		t.Fatal(err)
	}
	ctx := core.NewContext(pkgs, "example.com/gen", root, core.Architecture{}, nil)
	got := types.NewNoSetter().Check(ctx)

	var setTop, setKey, setValue int
	for _, v := range got {
		switch {
		case strings.Contains(v.Message, "SetTop"):
			setTop++
		case strings.Contains(v.Message, "SetKey"):
			setKey++
		case strings.Contains(v.Message, "SetValue"):
			setValue++
		}
	}
	if setTop != 1 {
		t.Fatalf("Stack[T].SetTop: got %d violations, want 1: %+v", setTop, got)
	}
	if setKey != 1 {
		t.Fatalf("Pair[K,V].SetKey: got %d violations, want 1: %+v", setKey, got)
	}
	if setValue != 0 {
		t.Fatalf("Builder[T].SetValue is fluent: got %d violations, want 0: %+v", setValue, got)
	}
}

func TestNoSetterSkipsFluentBuildersAndMocks(t *testing.T) {
	ctx := newFixtureContext(t, core.Architecture{}, nil)
	got := types.NewNoSetter().Check(ctx)

	for _, v := range got {
		if strings.Contains(v.File, "builder") {
			t.Fatalf("fluent builder should not be flagged: %+v", v)
		}
		if strings.Contains(v.File, "mocks") {
			t.Fatalf("mocks should be auto-excluded: %+v", v)
		}
	}
}
