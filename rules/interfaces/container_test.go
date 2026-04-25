package interfaces_test

import (
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/rules/interfaces"
)

func TestContainerDetectsFieldOnlyInterface(t *testing.T) {
	root := writeFixture(t, "example.com/container-field-only", map[string]string{
		"internal/wire/wire.go": `package wire

type repo interface {
	Get() string
}

type holder struct {
	r repo
}
`,
	})

	violations := interfaces.NewContainer().Check(loadContext(t, root, flatArchitecture(), "example.com/container-field-only"))

	assertHasRule(t, violations, "interface.container-only")
	if violations[0].EffectiveSeverity != core.Warning {
		t.Fatalf("EffectiveSeverity = %v, want Warning", violations[0].EffectiveSeverity)
	}
}

func TestContainerSkipsParameterAndReturnUsage(t *testing.T) {
	root := writeFixture(t, "example.com/container-real-abstraction", map[string]string{
		"internal/wire/wire.go": `package wire

type repo interface {
	Get() string
}

type holder struct {
	r repo
}

func newHolder(r repo) *holder {
	return &holder{r: r}
}
`,
		"internal/store/store.go": `package store

type Reader interface {
	Read() string
}

type reader struct{}

func New() Reader { return &reader{} }
`,
	})

	violations := interfaces.NewContainer().Check(loadContext(t, root, flatArchitecture(), "example.com/container-real-abstraction"))

	assertLacksRule(t, violations, "interface.container-only")
}

func TestContainerHandlesGenericTypeArguments(t *testing.T) {
	root := writeFixture(t, "example.com/container-generics", map[string]string{
		"internal/wire/wire.go": `package wire

type Reader interface {
	Read() string
}

type Cache[T any] struct {
	v T
}

func New() *Cache[Reader] { return &Cache[Reader]{} }
`,
	})

	violations := interfaces.NewContainer().Check(loadContext(t, root, flatArchitecture(), "example.com/container-generics"))

	assertLacksRule(t, violations, "interface.container-only")
}

func TestContainerWithSeverity(t *testing.T) {
	rule := interfaces.NewContainer(interfaces.WithSeverity(core.Error))

	if got := rule.Spec().DefaultSeverity; got != core.Error {
		t.Fatalf("DefaultSeverity = %v, want Error", got)
	}
}
