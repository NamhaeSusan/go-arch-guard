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
