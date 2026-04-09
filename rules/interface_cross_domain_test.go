package rules_test

import (
	"path/filepath"
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/rules"
)

// TestCheckCrossDomainAnonymous_StructField_Flagged verifies that an anonymous
// interface used as a struct field type with a method that references a domain
// type from another package is flagged.
func TestCheckCrossDomainAnonymous_StructField_Flagged(t *testing.T) {
	root := t.TempDir()
	module := "example.com/cda-struct-field"

	writeTestFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.21\n")

	writeTestFile(t, filepath.Join(root, "internal", "domain", "user", "alias.go"),
		`package user

type User struct {
	ID   string
	Name string
}
`)

	writeTestFile(t, filepath.Join(root, "internal", "wire", "wire.go"),
		`package wire

import (
	"context"

	"example.com/cda-struct-field/internal/domain/user"
)

type adapter struct {
	repo interface {
		GetByID(ctx context.Context, id string) (*user.User, error)
	}
}

func (a *adapter) Use(ctx context.Context) {
	_, _ = a.repo.GetByID(ctx, "x")
}
`)

	pkgs := loadTestPackages(t, root)
	vs := rules.CheckCrossDomainAnonymous(pkgs)

	found := false
	for _, v := range vs {
		if v.Rule == "interface.cross-domain-anonymous" {
			found = true
			if v.Severity != rules.Error {
				t.Errorf("expected Error severity, got %v", v.Severity)
			}
			break
		}
	}
	if !found {
		t.Errorf("expected interface.cross-domain-anonymous violation, got %d violations", len(vs))
		for _, v := range vs {
			t.Log(v.String())
		}
	}
}

// TestCheckCrossDomainAnonymous_NoDomainRef_NotFlagged verifies that an
// anonymous interface that does NOT reference any domain type is not flagged
// (e.g. interface{ String() string }).
func TestCheckCrossDomainAnonymous_NoDomainRef_NotFlagged(t *testing.T) {
	root := t.TempDir()
	module := "example.com/cda-no-ref"

	writeTestFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.21\n")

	writeTestFile(t, filepath.Join(root, "internal", "wire", "wire.go"),
		`package wire

type adapter struct {
	stringer interface {
		String() string
	}
}
`)

	pkgs := loadTestPackages(t, root)
	vs := rules.CheckCrossDomainAnonymous(pkgs)

	for _, v := range vs {
		if v.Rule == "interface.cross-domain-anonymous" {
			t.Errorf("did not expect violation, got: %s", v.String())
		}
	}
}

// TestCheckCrossDomainAnonymous_NamedInterface_NotFlagged verifies that a
// named interface (top-level type declaration) is NOT flagged by this rule
// even if it references domain types. This rule only targets anonymous
// interfaces — named interfaces fall under different rule territory.
func TestCheckCrossDomainAnonymous_NamedInterface_NotFlagged(t *testing.T) {
	root := t.TempDir()
	module := "example.com/cda-named"

	writeTestFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.21\n")

	writeTestFile(t, filepath.Join(root, "internal", "domain", "user", "alias.go"),
		`package user

type User struct {
	ID string
}
`)

	writeTestFile(t, filepath.Join(root, "internal", "wire", "wire.go"),
		`package wire

import (
	"context"

	"example.com/cda-named/internal/domain/user"
)

type Repo interface {
	GetByID(ctx context.Context, id string) (*user.User, error)
}

type adapter struct {
	repo Repo
}
`)

	pkgs := loadTestPackages(t, root)
	vs := rules.CheckCrossDomainAnonymous(pkgs)

	for _, v := range vs {
		if v.Rule == "interface.cross-domain-anonymous" {
			t.Errorf("named interface should not be flagged, got: %s", v.String())
		}
	}
}

// TestCheckCrossDomainAnonymous_SameDomain_NotFlagged verifies that an
// anonymous interface inside the same domain (e.g. user/handler using user.User)
// is NOT flagged.
func TestCheckCrossDomainAnonymous_SameDomain_NotFlagged(t *testing.T) {
	root := t.TempDir()
	module := "example.com/cda-same-domain"

	writeTestFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.21\n")

	writeTestFile(t, filepath.Join(root, "internal", "domain", "user", "alias.go"),
		`package user

type User struct {
	ID string
}
`)

	writeTestFile(t, filepath.Join(root, "internal", "domain", "user", "handler", "http", "handler.go"),
		`package http

import (
	"context"

	"example.com/cda-same-domain/internal/domain/user"
)

type handler struct {
	repo interface {
		GetByID(ctx context.Context, id string) (*user.User, error)
	}
}
`)

	pkgs := loadTestPackages(t, root)
	vs := rules.CheckCrossDomainAnonymous(pkgs)

	for _, v := range vs {
		if v.Rule == "interface.cross-domain-anonymous" {
			t.Errorf("same-domain anonymous interface should not be flagged, got: %s", v.String())
		}
	}
}

// TestCheckCrossDomainAnonymous_OrchestrationFlagged verifies that
// orchestration packages are also subject to the rule (any non-domain consumer
// is treated as cross-domain).
func TestCheckCrossDomainAnonymous_OrchestrationFlagged(t *testing.T) {
	root := t.TempDir()
	module := "example.com/cda-orchestration"

	writeTestFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.21\n")

	writeTestFile(t, filepath.Join(root, "internal", "domain", "draft", "alias.go"),
		`package draft

type ReviewDraft struct {
	ID int64
}
`)

	writeTestFile(t, filepath.Join(root, "internal", "orchestration", "submit.go"),
		`package orchestration

import (
	"context"

	"example.com/cda-orchestration/internal/domain/draft"
)

type service struct {
	drafts interface {
		GetDraftByID(ctx context.Context, id int64) (*draft.ReviewDraft, error)
	}
}
`)

	pkgs := loadTestPackages(t, root)
	vs := rules.CheckCrossDomainAnonymous(pkgs)

	found := false
	for _, v := range vs {
		if v.Rule == "interface.cross-domain-anonymous" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected violation in orchestration, got %d", len(vs))
		for _, v := range vs {
			t.Log(v.String())
		}
	}
}

// TestCheckCrossDomainAnonymous_TestFile_Skipped verifies that interfaces in
// _test.go files are not flagged (test mocks/fakes commonly use this shape).
func TestCheckCrossDomainAnonymous_TestFile_Skipped(t *testing.T) {
	root := t.TempDir()
	module := "example.com/cda-test-skip"

	writeTestFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.21\n")

	writeTestFile(t, filepath.Join(root, "internal", "domain", "user", "alias.go"),
		`package user

type User struct {
	ID string
}
`)

	// Real file: a basic struct so the package compiles
	writeTestFile(t, filepath.Join(root, "internal", "wire", "main.go"),
		`package wire

type Real struct{}
`)

	// Test file with anonymous interface — should NOT trigger
	writeTestFile(t, filepath.Join(root, "internal", "wire", "wire_test.go"),
		`package wire

import (
	"context"
	"testing"

	"example.com/cda-test-skip/internal/domain/user"
)

type fakeAdapter struct {
	repo interface {
		GetByID(ctx context.Context, id string) (*user.User, error)
	}
}

func TestStub(t *testing.T) {
	_ = fakeAdapter{}
}
`)

	pkgs := loadTestPackages(t, root)
	vs := rules.CheckCrossDomainAnonymous(pkgs)

	for _, v := range vs {
		if v.Rule == "interface.cross-domain-anonymous" {
			t.Errorf("test file violation should be skipped, got: %s", v.String())
		}
	}
}

// TestCheckCrossDomainAnonymous_FuncParameter_Flagged verifies that anonymous
// interfaces appearing as function parameter types are also flagged.
func TestCheckCrossDomainAnonymous_FuncParameter_Flagged(t *testing.T) {
	root := t.TempDir()
	module := "example.com/cda-func-param"

	writeTestFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.21\n")

	writeTestFile(t, filepath.Join(root, "internal", "domain", "user", "alias.go"),
		`package user

type User struct {
	ID string
}
`)

	writeTestFile(t, filepath.Join(root, "internal", "wire", "wire.go"),
		`package wire

import (
	"context"

	"example.com/cda-func-param/internal/domain/user"
)

func process(repo interface {
	GetByID(ctx context.Context, id string) (*user.User, error)
}) {
	_ = repo
}
`)

	pkgs := loadTestPackages(t, root)
	vs := rules.CheckCrossDomainAnonymous(pkgs)

	found := false
	for _, v := range vs {
		if v.Rule == "interface.cross-domain-anonymous" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected violation for anon iface as func param, got %d", len(vs))
		for _, v := range vs {
			t.Log(v.String())
		}
	}
}

// TestCheckCrossDomainAnonymous_EmptyInterface_NotFlagged verifies that
// interface{} (empty interface) is not flagged.
func TestCheckCrossDomainAnonymous_EmptyInterface_NotFlagged(t *testing.T) {
	root := t.TempDir()
	module := "example.com/cda-empty"

	writeTestFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.21\n")

	writeTestFile(t, filepath.Join(root, "internal", "wire", "wire.go"),
		`package wire

type bag struct {
	value interface{}
}
`)

	pkgs := loadTestPackages(t, root)
	vs := rules.CheckCrossDomainAnonymous(pkgs)

	for _, v := range vs {
		if v.Rule == "interface.cross-domain-anonymous" {
			t.Errorf("empty interface should not be flagged, got: %s", v.String())
		}
	}
}

// TestCheckCrossDomainAnonymous_PointerWrapper_Flagged verifies that pointer-
// or slice-wrapped anonymous interfaces are still flagged.
func TestCheckCrossDomainAnonymous_PointerWrapper_Flagged(t *testing.T) {
	root := t.TempDir()
	module := "example.com/cda-ptr-wrap"

	writeTestFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.21\n")

	writeTestFile(t, filepath.Join(root, "internal", "domain", "user", "alias.go"),
		`package user

type User struct {
	ID string
}
`)

	writeTestFile(t, filepath.Join(root, "internal", "wire", "wire.go"),
		`package wire

import (
	"context"

	"example.com/cda-ptr-wrap/internal/domain/user"
)

type adapter struct {
	repos []interface {
		GetByID(ctx context.Context, id string) (*user.User, error)
	}
}
`)

	pkgs := loadTestPackages(t, root)
	vs := rules.CheckCrossDomainAnonymous(pkgs)

	found := false
	for _, v := range vs {
		if v.Rule == "interface.cross-domain-anonymous" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected violation for slice-wrapped anonymous iface, got %d", len(vs))
		for _, v := range vs {
			t.Log(v.String())
		}
	}
}
