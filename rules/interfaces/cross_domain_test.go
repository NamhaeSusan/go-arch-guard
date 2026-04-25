package interfaces_test

import (
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/rules/interfaces"
)

func TestCrossDomainAnonymousDetectsAnonymousInterfaceOutsideDomain(t *testing.T) {
	root := writeFixture(t, "example.com/cross-domain-anon", map[string]string{
		"internal/domain/user/alias.go": `package user

type User struct {
	ID string
}
`,
		"internal/wire/wire.go": `package wire

import (
	"context"

	"example.com/cross-domain-anon/internal/domain/user"
)

type adapter struct {
	repo interface {
		Get(ctx context.Context) (*user.User, error)
	}
}
`,
	})

	violations := interfaces.NewCrossDomainAnonymous().Check(loadContext(t, root, domainArchitecture(), "example.com/cross-domain-anon"))

	assertHasRule(t, violations, "interface.cross-domain-anonymous")
	if violations[0].EffectiveSeverity != core.Error {
		t.Fatalf("EffectiveSeverity = %v, want Error", violations[0].EffectiveSeverity)
	}
}

func TestCrossDomainAnonymousSkipsNamedSameDomainAndOrchestration(t *testing.T) {
	root := writeFixture(t, "example.com/cross-domain-skip", map[string]string{
		"internal/domain/user/alias.go": `package user

type User struct {
	ID string
}
`,
		"internal/domain/user/handler/http.go": `package handler

import (
	"context"

	"example.com/cross-domain-skip/internal/domain/user"
)

type handler struct {
	repo interface {
		Get(ctx context.Context) (*user.User, error)
	}
}
`,
		"internal/orchestration/orch.go": `package orchestration

import (
	"context"

	"example.com/cross-domain-skip/internal/domain/user"
)

type service struct {
	repo interface {
		Get(ctx context.Context) (*user.User, error)
	}
}
`,
		"internal/wire/wire.go": `package wire

import (
	"context"

	"example.com/cross-domain-skip/internal/domain/user"
)

type Repo interface {
	Get(ctx context.Context) (*user.User, error)
}

type adapter struct {
	repo Repo
}
`,
	})

	violations := interfaces.NewCrossDomainAnonymous().Check(loadContext(t, root, domainArchitecture(), "example.com/cross-domain-skip"))

	assertLacksRule(t, violations, "interface.cross-domain-anonymous")
}

func TestCrossDomainAnonymousWithSeverity(t *testing.T) {
	rule := interfaces.NewCrossDomainAnonymous(interfaces.WithSeverity(core.Warning))

	if got := rule.Spec().DefaultSeverity; got != core.Warning {
		t.Fatalf("DefaultSeverity = %v, want Warning", got)
	}
}
