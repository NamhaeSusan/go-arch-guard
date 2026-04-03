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
