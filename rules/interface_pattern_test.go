package rules_test

import (
	"path/filepath"
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/rules"
)

func TestCheckInterfacePattern_Valid(t *testing.T) {
	root := t.TempDir()
	module := "example.com/ip-valid"

	writeTestFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.21\n")

	// store package: unexported impl + New() returns interface → 0 violations
	writeTestFile(t, filepath.Join(root, "internal", "store", "store.go"),
		`package store

type Store interface {
	Get(id string) string
}

type store struct{}

func (s *store) Get(id string) string { return "" }

func New() Store { return &store{} }
`)

	pkgs := loadTestPackages(t, root)
	vs := rules.CheckInterfacePattern(pkgs, rules.WithModel(rules.ConsumerWorker()))

	if len(vs) != 0 {
		for _, v := range vs {
			t.Log(v.String())
		}
		t.Errorf("expected 0 violations, got %d", len(vs))
	}
}

func TestCheckInterfacePattern_ExportedImpl(t *testing.T) {
	root := t.TempDir()
	module := "example.com/ip-exported-impl"

	writeTestFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.21\n")

	// store package: exported struct implements interface → violation
	writeTestFile(t, filepath.Join(root, "internal", "store", "store.go"),
		`package store

type Store interface {
	Get(id string) string
}

type DBStore struct{}

func (s *DBStore) Get(id string) string { return "" }

func New() Store { return &DBStore{} }
`)

	pkgs := loadTestPackages(t, root)
	vs := rules.CheckInterfacePattern(pkgs, rules.WithModel(rules.ConsumerWorker()))

	found := false
	for _, v := range vs {
		if v.Rule == "interface.exported-impl" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected interface.exported-impl violation")
	}
}

func TestCheckInterfacePattern_ExcludedLayer(t *testing.T) {
	root := t.TempDir()
	module := "example.com/ip-excluded"

	writeTestFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.21\n")

	// model package is excluded in ConsumerWorker — no violations even with exported struct
	writeTestFile(t, filepath.Join(root, "internal", "model", "model.go"),
		`package model

type Order struct {
	ID   string
	Name string
}
`)

	pkgs := loadTestPackages(t, root)
	vs := rules.CheckInterfacePattern(pkgs, rules.WithModel(rules.ConsumerWorker()))

	if len(vs) != 0 {
		for _, v := range vs {
			t.Log(v.String())
		}
		t.Errorf("expected 0 violations, got %d", len(vs))
	}
}

func TestCheckInterfacePattern_ConstructorName(t *testing.T) {
	root := t.TempDir()
	module := "example.com/ip-ctor-name"

	writeTestFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.21\n")

	// store package: NewStore() instead of New() → constructor-name violation
	writeTestFile(t, filepath.Join(root, "internal", "store", "store.go"),
		`package store

type Store interface {
	Get(id string) string
}

type store struct{}

func (s *store) Get(id string) string { return "" }

func NewStore() Store { return &store{} }
`)

	pkgs := loadTestPackages(t, root)
	vs := rules.CheckInterfacePattern(pkgs, rules.WithModel(rules.ConsumerWorker()))

	found := false
	for _, v := range vs {
		if v.Rule == "interface.constructor-name" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected interface.constructor-name violation")
	}
}

func TestCheckInterfacePattern_ConstructorNameValid(t *testing.T) {
	root := t.TempDir()
	module := "example.com/ip-ctor-name-valid"

	writeTestFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.21\n")

	// store package: exact New() → no constructor-name violation
	writeTestFile(t, filepath.Join(root, "internal", "store", "store.go"),
		`package store

type Store interface {
	Get(id string) string
}

type store struct{}

func (s *store) Get(id string) string { return "" }

func New() Store { return &store{} }
`)

	pkgs := loadTestPackages(t, root)
	vs := rules.CheckInterfacePattern(pkgs, rules.WithModel(rules.ConsumerWorker()))

	for _, v := range vs {
		if v.Rule == "interface.constructor-name" {
			t.Errorf("unexpected constructor-name violation: %s", v.String())
		}
	}
}

func TestCheckInterfacePattern_ConstructorReturnsConcrete(t *testing.T) {
	root := t.TempDir()
	module := "example.com/ip-ctor-concrete"

	writeTestFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.21\n")

	// store package: New() returns *store → constructor-returns-interface violation
	writeTestFile(t, filepath.Join(root, "internal", "store", "store.go"),
		`package store

type Store interface {
	Get(id string) string
}

type store struct{}

func (s *store) Get(id string) string { return "" }

func New() *store { return &store{} }
`)

	pkgs := loadTestPackages(t, root)
	vs := rules.CheckInterfacePattern(pkgs, rules.WithModel(rules.ConsumerWorker()))

	found := false
	for _, v := range vs {
		if v.Rule == "interface.constructor-returns-interface" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected interface.constructor-returns-interface violation")
	}
}

func TestCheckInterfacePattern_SinglePerPackage_Violation(t *testing.T) {
	root := t.TempDir()
	module := "example.com/sp-bad"

	writeTestFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.21\n")
	writeTestFile(t, filepath.Join(root, "internal", "store", "store.go"),
		`package store

type OrderStore interface {
	FindOrder() error
}

type UserStore interface {
	FindUser() error
}
`)

	pkgs := loadTestPackages(t, root)
	violations := rules.CheckInterfacePattern(pkgs, rules.WithModel(rules.ConsumerWorker()))
	found := false
	for _, v := range violations {
		if v.Rule == "interface.single-per-package" {
			found = true
			if v.Severity != rules.Error {
				t.Errorf("expected Error severity (default), got %v", v.Severity)
			}
		}
	}
	if !found {
		t.Error("expected interface.single-per-package violation")
	}
}

func TestCheckInterfacePattern_SinglePerPackage_Valid(t *testing.T) {
	root := t.TempDir()
	module := "example.com/sp-valid"

	writeTestFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.21\n")
	writeTestFile(t, filepath.Join(root, "internal", "store", "store.go"),
		`package store

type Store interface {
	Find() error
}
`)

	pkgs := loadTestPackages(t, root)
	violations := rules.CheckInterfacePattern(pkgs, rules.WithModel(rules.ConsumerWorker()))
	for _, v := range violations {
		if v.Rule == "interface.single-per-package" {
			t.Errorf("unexpected single-per-package violation")
		}
	}
}

// TestCheckInterfacePattern_ExportedImpl_SignatureMismatch verifies that a struct
// with a method whose name matches an interface method but whose signature differs
// is NOT flagged as implementing the interface.
func TestCheckInterfacePattern_ExportedImpl_SignatureMismatch(t *testing.T) {
	root := t.TempDir()
	module := "example.com/ip-sig-mismatch"

	writeTestFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.21\n")

	// StoreImpl.Find has different parameter types than Store.Find — not an impl.
	writeTestFile(t, filepath.Join(root, "internal", "store", "store.go"),
		`package store

type Store interface {
	Find(id string) string
}

type StoreImpl struct{}

func (s *StoreImpl) Find(limit, offset int) []string { return nil }
`)

	pkgs := loadTestPackages(t, root)
	vs := rules.CheckInterfacePattern(pkgs, rules.WithModel(rules.ConsumerWorker()))

	for _, v := range vs {
		if v.Rule == "interface.exported-impl" {
			t.Errorf("unexpected interface.exported-impl violation when signatures do not match: %s", v.String())
		}
	}
}

// TestCheckInterfacePattern_ExportedImpl_SignatureMatch verifies that a struct
// whose method names AND signatures match the interface is still flagged.
func TestCheckInterfacePattern_ExportedImpl_SignatureMatch(t *testing.T) {
	root := t.TempDir()
	module := "example.com/ip-sig-match"

	writeTestFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.21\n")

	writeTestFile(t, filepath.Join(root, "internal", "store", "store.go"),
		`package store

type Store interface {
	Find(id string) string
}

type StoreImpl struct{}

func (s *StoreImpl) Find(id string) string { return "" }
`)

	pkgs := loadTestPackages(t, root)
	vs := rules.CheckInterfacePattern(pkgs, rules.WithModel(rules.ConsumerWorker()))

	found := false
	for _, v := range vs {
		if v.Rule == "interface.exported-impl" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected interface.exported-impl violation when signatures match")
	}
}

// TestCheckInterfacePattern_ExportedImpl_AliasInterface verifies that an
// exported alias of an anonymous interface is still treated as an interface,
// so an exported struct implementing it is flagged. Regression for the
// Go 1.26 *types.Alias case that was silently skipped.
func TestCheckInterfacePattern_ExportedImpl_AliasInterface(t *testing.T) {
	root := t.TempDir()
	module := "example.com/ip-alias-iface"

	writeTestFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.21\n")

	writeTestFile(t, filepath.Join(root, "internal", "store", "store.go"),
		`package store

type Store = interface {
	Get(id string) string
}

type StoreImpl struct{}

func (s *StoreImpl) Get(id string) string { return "" }
`)

	pkgs := loadTestPackages(t, root)
	vs := rules.CheckInterfacePattern(pkgs, rules.WithModel(rules.ConsumerWorker()))

	found := false
	for _, v := range vs {
		if v.Rule == "interface.exported-impl" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected interface.exported-impl violation for alias-based interface")
	}
}

// TestCheckInterfacePattern_ExportedImpl_IllTyped verifies that when a package
// has unrelated type errors (pkg.Types missing or objects unresolvable), the
// rule still reports via the AST name-based fallback instead of silently
// dropping coverage.
func TestCheckInterfacePattern_ExportedImpl_IllTyped(t *testing.T) {
	root := t.TempDir()
	module := "example.com/ip-illtyped"

	writeTestFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.21\n")

	// Exported struct implements the interface (same name + signature), but the
	// file also contains an unrelated type error so the package is IllTyped.
	// The fallback must still emit interface.exported-impl.
	// The method receiver references an undefined parameter type, so the
	// *types.Named for DBStore lacks a usable Get method and types.Implements
	// returns false. But AST inspection still shows method name Get matching.
	// The fallback must emit the violation.
	writeTestFile(t, filepath.Join(root, "internal", "store", "store.go"),
		`package store

type Store interface {
	Get(id string) string
}

type DBStore struct{}

func (s *DBStore) Get(id UndefinedType) string { return "" }
`)

	pkgs := loadTestPackages(t, root)
	vs := rules.CheckInterfacePattern(pkgs, rules.WithModel(rules.ConsumerWorker()))

	found := false
	for _, v := range vs {
		if v.Rule == "interface.exported-impl" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected interface.exported-impl violation in ill-typed package (via AST fallback)")
	}
}

func TestCheckInterfacePattern_ConstructorReturnsInterfaceValid(t *testing.T) {
	root := t.TempDir()
	module := "example.com/ip-ctor-iface-valid"

	writeTestFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.21\n")

	// store package: New() returns Store interface → no violation
	writeTestFile(t, filepath.Join(root, "internal", "store", "store.go"),
		`package store

type Store interface {
	Get(id string) string
}

type store struct{}

func (s *store) Get(id string) string { return "" }

func New() Store { return &store{} }
`)

	pkgs := loadTestPackages(t, root)
	vs := rules.CheckInterfacePattern(pkgs, rules.WithModel(rules.ConsumerWorker()))

	for _, v := range vs {
		if v.Rule == "interface.constructor-returns-interface" {
			t.Errorf("unexpected constructor-returns-interface violation: %s", v.String())
		}
	}
}
