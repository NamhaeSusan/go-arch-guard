package tui

import (
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/core"
)

func TestWriteViolationsDoesNotMutateIndexOrder(t *testing.T) {
	index := ViolationIndex{
		"internal/order": {
			{Rule: "warning.rule", File: "b.go", EffectiveSeverity: core.Warning},
			{Rule: "error.rule", File: "a.go", EffectiveSeverity: core.Error},
		},
	}
	panel := NewDetailPanel(nil, index, nil, "example.com/app")
	node := &PkgNode{
		RelPath: "internal/order",
		IsLeaf:  true,
	}

	panel.Update(node)

	got := index["internal/order"]
	if got[0].Rule != "warning.rule" || got[1].Rule != "error.rule" {
		t.Fatalf("Update mutated violation order: %+v", got)
	}
}
