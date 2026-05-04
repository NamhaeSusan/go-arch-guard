package naming_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/analyzer"
	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/presets"
	"github.com/NamhaeSusan/go-arch-guard/rules/naming"
)

func TestConstructorNameSpec(t *testing.T) {
	spec := naming.NewConstructorName(naming.WithSeverity(core.Error)).Spec()

	if spec.ID != "infra.constructor-name" {
		t.Fatalf("ID = %q, want infra.constructor-name", spec.ID)
	}
	if spec.DefaultSeverity != core.Error {
		t.Fatalf("DefaultSeverity = %v, want Error", spec.DefaultSeverity)
	}
}

func TestConstructorNameFlagsNewVariantsReturningSamePackageTypes(t *testing.T) {
	ctx := infraContext(t, map[string]string{
		"internal/domain/order/infra/memory/store.go": `package memory

type Store struct{}

func NewStore() *Store { return &Store{} }
func NewRepository() Store { return Store{} }
func NewOrderMemoryStore() (*Store, error) { return &Store{}, nil }
`,
	})

	got := naming.NewConstructorName().Check(ctx)

	for _, name := range []string{"NewStore", "NewRepository", "NewOrderMemoryStore"} {
		assertConstructorViolation(t, got, name)
	}
}

func TestConstructorNameAllowsNewAndNonConstructors(t *testing.T) {
	ctx := infraContext(t, map[string]string{
		"internal/domain/order/infra/memory/store.go": `package memory

type Store struct{}
type Repository interface{ Save() error }

func New() *Store { return &Store{} }
func NewRepository() Repository { return nil }
func NewMetricName() string { return "" }
func BuildStore() *Store { return &Store{} }
`,
	})

	got := naming.NewConstructorName().Check(ctx)
	if len(got) != 0 {
		t.Fatalf("expected no constructor-name violations, got %+v", got)
	}
}

func TestConstructorNameAllowsConfiguredNames(t *testing.T) {
	ctx := infraContext(t, map[string]string{
		"internal/domain/order/infra/memory/store.go": `package memory

type Store struct{}

func NewStore() *Store { return &Store{} }
`,
	})

	got := naming.NewConstructorName(naming.WithAllowedConstructorNames("New", "NewStore")).Check(ctx)
	if len(got) != 0 {
		t.Fatalf("configured constructor name should be allowed, got %+v", got)
	}
}

func TestConstructorNameOnlyChecksInfraPackages(t *testing.T) {
	ctx := infraContext(t, map[string]string{
		"internal/domain/order/app/service.go": `package app

type Service struct{}

func NewService() *Service { return &Service{} }
`,
	})

	got := naming.NewConstructorName().Check(ctx)
	if len(got) != 0 {
		t.Fatalf("non-infra package should not be checked, got %+v", got)
	}
}

func infraContext(t *testing.T, files map[string]string) *core.Context {
	t.Helper()
	root := t.TempDir()
	writeInfraFile(t, filepath.Join(root, "go.mod"), "module example.com/shop\n\ngo 1.25.0\n")
	for name, content := range files {
		writeInfraFile(t, filepath.Join(root, name), content)
	}
	pkgs, err := analyzer.Load(root, "internal/...")
	if err != nil {
		t.Fatal(err)
	}
	return core.NewContext(pkgs, "example.com/shop", root, presets.DDD(), nil)
}

func writeInfraFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func assertConstructorViolation(t *testing.T, violations []core.Violation, name string) {
	t.Helper()
	for _, v := range violations {
		if v.Rule == "infra.constructor-name" && strings.Contains(v.Message, name) {
			if v.DefaultSeverity != core.Warning || v.EffectiveSeverity != core.Warning {
				t.Fatalf("severity = default %v effective %v, want Warning/Warning", v.DefaultSeverity, v.EffectiveSeverity)
			}
			return
		}
	}
	t.Fatalf("expected infra.constructor-name violation for %s, got %+v", name, violations)
}
