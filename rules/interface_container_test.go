package rules_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/rules"
)

// TestCheckContainerInterface_FieldOnly_DetectsSmell verifies that an interface
// used only as a struct field type (no parameter, no return) is flagged as a
// container-only smell.
func TestCheckContainerInterface_FieldOnly_DetectsSmell(t *testing.T) {
	root := t.TempDir()
	module := "example.com/container-field-only"

	writeTestFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.21\n")

	// repo interface used only as struct field — never as param or return.
	writeTestFile(t, filepath.Join(root, "internal", "wire", "wire.go"),
		`package wire

type userRepo interface {
	GetByID(id string) string
}

type holder struct {
	r userRepo
}
`)

	pkgs := loadTestPackages(t, root)
	vs := rules.CheckContainerInterface(pkgs)

	found := false
	for _, v := range vs {
		if v.Rule == "interface.container-only" {
			found = true
			if v.Severity != rules.Warning {
				t.Errorf("expected Warning severity, got %v", v.Severity)
			}
			break
		}
	}
	if !found {
		t.Errorf("expected interface.container-only violation, got %d violations", len(vs))
		for _, v := range vs {
			t.Log(v.String())
		}
	}
}

// TestCheckContainerInterface_AsParameter_NoSmell verifies that an interface
// used as a function parameter is treated as a legitimate consumer-defined
// interface and is NOT flagged.
func TestCheckContainerInterface_AsParameter_NoSmell(t *testing.T) {
	root := t.TempDir()
	module := "example.com/container-as-param"

	writeTestFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.21\n")

	writeTestFile(t, filepath.Join(root, "internal", "wire", "wire.go"),
		`package wire

type userRepo interface {
	GetByID(id string) string
}

type holder struct {
	r userRepo
}

func newHolder(r userRepo) *holder {
	return &holder{r: r}
}
`)

	pkgs := loadTestPackages(t, root)
	vs := rules.CheckContainerInterface(pkgs)

	for _, v := range vs {
		if v.Rule == "interface.container-only" {
			t.Errorf("did not expect container-only violation, got: %s", v.String())
		}
	}
}

// TestCheckContainerInterface_AsReturnType_NoSmell verifies that an interface
// used as a function return type is NOT flagged.
func TestCheckContainerInterface_AsReturnType_NoSmell(t *testing.T) {
	root := t.TempDir()
	module := "example.com/container-as-return"

	writeTestFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.21\n")

	writeTestFile(t, filepath.Join(root, "internal", "store", "store.go"),
		`package store

type Reader interface {
	Read() string
}

type fileReader struct{}

func (f *fileReader) Read() string { return "" }

func NewReader() Reader {
	return &fileReader{}
}
`)

	pkgs := loadTestPackages(t, root)
	vs := rules.CheckContainerInterface(pkgs)

	for _, v := range vs {
		if v.Rule == "interface.container-only" {
			t.Errorf("did not expect container-only violation, got: %s", v.String())
		}
	}
}

// TestCheckContainerInterface_TypeAlias_NoSmell verifies that a type alias
// (not a new interface declaration) is NOT flagged even if used as a field.
func TestCheckContainerInterface_TypeAlias_NoSmell(t *testing.T) {
	root := t.TempDir()
	module := "example.com/container-alias"

	writeTestFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.21\n")

	writeTestFile(t, filepath.Join(root, "internal", "core", "core.go"),
		`package core

type Reader interface {
	Read() string
}
`)

	writeTestFile(t, filepath.Join(root, "internal", "wire", "wire.go"),
		`package wire

import "example.com/container-alias/internal/core"

type Reader = core.Reader

type holder struct {
	r Reader
}
`)

	pkgs := loadTestPackages(t, root)
	vs := rules.CheckContainerInterface(pkgs)

	for _, v := range vs {
		if v.Rule == "interface.container-only" {
			t.Errorf("did not expect container-only violation for type alias, got: %s", v.String())
		}
	}
}

// TestCheckContainerInterface_UnusedInterface_NoSmell verifies that an interface
// that is not used at all is NOT flagged. That is a different smell category
// (dead interface), out of scope for container-only.
func TestCheckContainerInterface_UnusedInterface_NoSmell(t *testing.T) {
	root := t.TempDir()
	module := "example.com/container-unused"

	writeTestFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.21\n")

	writeTestFile(t, filepath.Join(root, "internal", "store", "store.go"),
		`package store

type Orphan interface {
	Foo()
}
`)

	pkgs := loadTestPackages(t, root)
	vs := rules.CheckContainerInterface(pkgs)

	for _, v := range vs {
		if v.Rule == "interface.container-only" {
			t.Errorf("did not expect container-only violation for unused interface, got: %s", v.String())
		}
	}
}

// TestCheckContainerInterface_TestFile_Skipped verifies that interfaces declared
// in _test.go files are skipped (mock/fake fixtures often use this shape).
func TestCheckContainerInterface_TestFile_Skipped(t *testing.T) {
	root := t.TempDir()
	module := "example.com/container-test-skip"

	writeTestFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.21\n")

	// non-test file: a real type so the package is not empty
	writeTestFile(t, filepath.Join(root, "internal", "store", "store.go"),
		`package store

type Real struct{}
`)

	// test file: container-only interface — should NOT trigger
	writeTestFile(t, filepath.Join(root, "internal", "store", "store_test.go"),
		`package store

type fakeRepo interface {
	Get() string
}

type fakeHolder struct {
	r fakeRepo
}
`)

	pkgs := loadTestPackages(t, root)
	vs := rules.CheckContainerInterface(pkgs)

	for _, v := range vs {
		if v.Rule == "interface.container-only" {
			t.Errorf("did not expect container-only violation in test file, got: %s", v.String())
		}
	}
}

// TestCheckContainerInterface_TwoFieldsTwoInterfaces verifies that the rule
// independently evaluates each interface and only flags the container-only one.
func TestCheckContainerInterface_TwoFieldsTwoInterfaces(t *testing.T) {
	root := t.TempDir()
	module := "example.com/container-mixed"

	writeTestFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.21\n")

	writeTestFile(t, filepath.Join(root, "internal", "wire", "wire.go"),
		`package wire

// container-only — should be flagged
type badRepo interface {
	Foo() string
}

// used as parameter — should NOT be flagged
type goodRepo interface {
	Bar() string
}

type holder struct {
	bad  badRepo
	good goodRepo
}

func newHolder(g goodRepo) *holder {
	return &holder{good: g}
}
`)

	pkgs := loadTestPackages(t, root)
	vs := rules.CheckContainerInterface(pkgs)

	badFound := false
	for _, v := range vs {
		if v.Rule != "interface.container-only" {
			continue
		}
		if v.Message == "" {
			t.Error("violation message is empty")
		}
		// expect only badRepo to be flagged, not goodRepo
		if strings.Contains(v.Message, "badRepo") {
			badFound = true
		}
		if strings.Contains(v.Message, "goodRepo") {
			t.Errorf("goodRepo should not be flagged, got: %s", v.String())
		}
	}
	if !badFound {
		t.Errorf("expected badRepo to be flagged as container-only, got %d violations", len(vs))
		for _, v := range vs {
			t.Log(v.String())
		}
	}
}

// TestCheckContainerInterface_SeverityOverride verifies that WithSeverity(Error)
// upgrades the severity if the caller wants strict mode.
func TestCheckContainerInterface_SeverityOverride(t *testing.T) {
	root := t.TempDir()
	module := "example.com/container-severity"

	writeTestFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.21\n")

	writeTestFile(t, filepath.Join(root, "internal", "wire", "wire.go"),
		`package wire

type repo interface {
	Get() string
}

type holder struct {
	r repo
}
`)

	pkgs := loadTestPackages(t, root)
	vs := rules.CheckContainerInterface(pkgs, rules.WithSeverity(rules.Error))

	found := false
	for _, v := range vs {
		if v.Rule == "interface.container-only" {
			found = true
			if v.Severity != rules.Error {
				t.Errorf("expected Error severity after WithSeverity(Error), got %v", v.Severity)
			}
		}
	}
	if !found {
		t.Error("expected container-only violation")
	}
}
