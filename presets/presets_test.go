package presets_test

import (
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/presets"
)

func TestArchitecturesValidate(t *testing.T) {
	tests := map[string]func() core.Architecture{
		"DDD":             presets.DDD,
		"CleanArch":       presets.CleanArch,
		"Layered":         presets.Layered,
		"Hexagonal":       presets.Hexagonal,
		"ModularMonolith": presets.ModularMonolith,
		"ConsumerWorker":  presets.ConsumerWorker,
		"Batch":           presets.Batch,
		"EventPipeline":   presets.EventPipeline,
	}

	for name, build := range tests {
		t.Run(name, func(t *testing.T) {
			arch := build()
			if err := arch.Validate(); err != nil {
				t.Fatalf("Validate() error = %v", err)
			}
		})
	}
}

func TestRecommendedRuleSetsAreNonEmpty(t *testing.T) {
	tests := map[string]func() core.RuleSet{
		"DDD":             presets.RecommendedDDD,
		"CleanArch":       presets.RecommendedCleanArch,
		"Layered":         presets.RecommendedLayered,
		"Hexagonal":       presets.RecommendedHexagonal,
		"ModularMonolith": presets.RecommendedModularMonolith,
		"ConsumerWorker":  presets.RecommendedConsumerWorker,
		"Batch":           presets.RecommendedBatch,
		"EventPipeline":   presets.RecommendedEventPipeline,
	}

	for name, build := range tests {
		t.Run(name, func(t *testing.T) {
			if got := len(build().Rules()); got == 0 {
				t.Fatal("recommended RuleSet is empty")
			}
		})
	}
}

// TestRecommendedRuleSetsContainCoreRules pins the membership of each
// recommended bundle so a silent rule deletion (e.g. dropping
// dependency.NewLayerDirection from RecommendedDDD) breaks this test.
func TestRecommendedRuleSetsContainCoreRules(t *testing.T) {
	// Core rule-type IDs that every recommended bundle must carry.
	commonRules := []string{
		"dependency.layer-direction",
		"dependency.blast-radius",
		"naming.no-stutter",
		"naming.no-impl-suffix",
		"interfaces.pattern",
		"interfaces.container",
		"types.no-setter",
	}

	cases := []struct {
		name    string
		build   func() core.RuleSet
		require []string // additional rule IDs this preset must include
		exclude []string // rule IDs this preset must NOT include
	}{
		{
			name:    "DDD",
			build:   presets.RecommendedDDD,
			require: []string{"dependency.isolation", "structural.alias", "structural.model-required"},
		},
		{
			name:    "CleanArch",
			build:   presets.RecommendedCleanArch,
			require: []string{"dependency.isolation"},
			exclude: []string{"structural.alias", "structural.model-required"},
		},
		{
			name:    "Hexagonal",
			build:   presets.RecommendedHexagonal,
			require: []string{"dependency.isolation"},
			exclude: []string{"structural.alias", "structural.model-required"},
		},
		{
			name:    "ConsumerWorker",
			build:   presets.RecommendedConsumerWorker,
			require: []string{"types.type-pattern"},
			exclude: []string{"dependency.isolation", "structural.alias"},
		},
		{
			name:    "Batch",
			build:   presets.RecommendedBatch,
			require: []string{"types.type-pattern"},
			exclude: []string{"dependency.isolation", "structural.alias"},
		},
		{
			name:    "EventPipeline",
			build:   presets.RecommendedEventPipeline,
			require: []string{"types.type-pattern"},
			exclude: []string{"dependency.isolation", "structural.alias"},
		},
	}

	// `tx.boundary` is opt-in — must NOT appear in any recommended bundle.
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ids := make(map[string]bool)
			for _, r := range tc.build().Rules() {
				ids[r.Spec().ID] = true
			}
			for _, id := range commonRules {
				if !ids[id] {
					t.Errorf("missing common rule %q in Recommended%s", id, tc.name)
				}
			}
			for _, id := range tc.require {
				if !ids[id] {
					t.Errorf("missing required rule %q in Recommended%s", id, tc.name)
				}
			}
			for _, id := range tc.exclude {
				if ids[id] {
					t.Errorf("unexpected rule %q in Recommended%s", id, tc.name)
				}
			}
			if ids["tx.boundary"] {
				t.Errorf("tx.boundary must remain opt-in; found in Recommended%s", tc.name)
			}
		})
	}
}
