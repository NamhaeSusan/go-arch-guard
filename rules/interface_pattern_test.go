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
