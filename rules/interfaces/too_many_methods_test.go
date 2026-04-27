package interfaces_test

import (
	"strings"
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/rules/interfaces"
)

func TestTooManyMethodsSpec(t *testing.T) {
	spec := interfaces.NewTooManyMethods(interfaces.WithSeverity(core.Warning)).Spec()

	if spec.ID != "interfaces.too-many-methods" {
		t.Fatalf("Spec().ID = %q, want interfaces.too-many-methods", spec.ID)
	}
	if spec.DefaultSeverity != core.Warning {
		t.Fatalf("Spec().DefaultSeverity = %v, want Warning", spec.DefaultSeverity)
	}
}

func TestTooManyMethodsDefaultCapIs10(t *testing.T) {
	root := writeFixture(t, "example.com/too-many-default", map[string]string{
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

	violations := interfaces.NewTooManyMethods().Check(loadContext(t, root, flatArchitecture(), "example.com/too-many-default"))

	var found bool
	for _, v := range violations {
		if v.Rule == "interfaces.too-many-methods" && strings.Contains(v.Message, "11 methods") && strings.Contains(v.Message, "at most 10") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected interfaces.too-many-methods violation with default cap 10, got %+v", violations)
	}
}

func TestTooManyMethodsRespectsCustomCap(t *testing.T) {
	root := writeFixture(t, "example.com/too-many-custom", map[string]string{
		"internal/store/store.go": `package store

type Store interface {
	A()
	B()
}

type store struct{}

func New() Store { return &store{} }
`,
	})

	violations := interfaces.NewTooManyMethods(interfaces.WithMaxMethods(1)).
		Check(loadContext(t, root, flatArchitecture(), "example.com/too-many-custom"))

	assertHasRule(t, violations, "interfaces.too-many-methods")
}

func TestTooManyMethodsDoesNotFlagBelowCap(t *testing.T) {
	root := writeFixture(t, "example.com/too-many-ok", map[string]string{
		"internal/store/store.go": `package store

type Store interface {
	A()
	B()
	C()
}

type store struct{}

func New() Store { return &store{} }
`,
	})

	violations := interfaces.NewTooManyMethods(interfaces.WithMaxMethods(5)).
		Check(loadContext(t, root, flatArchitecture(), "example.com/too-many-ok"))

	assertLacksRule(t, violations, "interfaces.too-many-methods")
}

func TestTooManyMethodsTreatsZeroAsDefault(t *testing.T) {
	// WithMaxMethods(0) means "use default" — same fixture as default test
	// but with the explicit (forgiving) zero value.
	root := writeFixture(t, "example.com/too-many-zero", map[string]string{
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

	violations := interfaces.NewTooManyMethods(interfaces.WithMaxMethods(0)).
		Check(loadContext(t, root, flatArchitecture(), "example.com/too-many-zero"))

	assertHasRule(t, violations, "interfaces.too-many-methods")
}

func TestTooManyMethodsDoesNotEmitMetaWhenOptionMissing(t *testing.T) {
	// Sanity check: NewTooManyMethods() without WithMaxMethods must not emit
	// meta.rule-disabled-by-config — the rule has a sensible default and is
	// active out of the box. The legacy meta emit (which lived on Pattern's
	// missing-WithMaxMethods case) is gone.
	root := writeFixture(t, "example.com/too-many-no-meta", map[string]string{
		"internal/store/store.go": `package store

type Store interface {
	A()
}

type store struct{}

func New() Store { return &store{} }
`,
	})

	got := interfaces.NewTooManyMethods().Check(loadContext(t, root, flatArchitecture(), "example.com/too-many-no-meta"))
	for _, v := range got {
		if v.Rule == "meta.rule-disabled-by-config" {
			t.Fatalf("NewTooManyMethods must not emit meta.rule-disabled-by-config; got %+v", v)
		}
	}
}
