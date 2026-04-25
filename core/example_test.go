package core_test

import (
	"fmt"

	"github.com/NamhaeSusan/go-arch-guard/core"
)

// noDBInHandler is a hand-written example of a custom rule. Real rules
// would inspect ctx.Pkgs(); this fake scans a hard-coded import list to
// keep the example hermetic.
type noDBInHandler struct {
	imports []importEdge
}

type importEdge struct {
	From, To, File string
	Line           int
}

func (r *noDBInHandler) Spec() core.RuleSpec {
	return core.RuleSpec{
		ID:              "team.no-db-in-handler",
		Description:     "handlers must not import database/sql",
		DefaultSeverity: core.Error,
		Violations: []core.ViolationSpec{
			{ID: "team.no-db-in-handler", Description: "handler imports database/sql", DefaultSeverity: core.Error},
		},
	}
}

func (r *noDBInHandler) Check(ctx *core.Context) []core.Violation {
	var out []core.Violation
	for _, edge := range r.imports {
		if edge.From == "internal/handler" && edge.To == "database/sql" {
			out = append(out, core.Violation{
				File:    edge.File,
				Line:    edge.Line,
				Rule:    "team.no-db-in-handler",
				Message: "handler must not import database/sql",
				Fix:     "move DB calls to internal/repo and inject through interface",
			})
		}
	}
	return out
}

func Example() {
	arch := core.Architecture{
		Layers: core.LayerModel{
			Sublayers: []string{"handler", "core"},
			Direction: map[string][]string{
				"handler": {"core"},
				"core":    {},
			},
		},
	}
	ctx := core.NewContext(nil, "github.com/example/app", "/repo", arch, nil)

	rule := &noDBInHandler{
		imports: []importEdge{
			{From: "internal/handler", To: "database/sql", File: "internal/handler/users.go", Line: 7},
		},
	}

	violations := core.Run(ctx, core.RuleSet{}.With(rule))
	for _, v := range violations {
		fmt.Println(v)
	}
	// Output:
	// [ERROR] violation: handler must not import database/sql (file: internal/handler/users.go:7, rule: team.no-db-in-handler, fix: move DB calls to internal/repo and inject through interface)
}
