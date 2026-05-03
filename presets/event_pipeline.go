package presets

import (
	"github.com/NamhaeSusan/go-arch-guard/core"
)

func EventPipeline() core.Architecture {
	arch := core.Architecture{
		Layers: core.LayerModel{
			Sublayers: []string{
				"command", "aggregate", "event", "projection",
				"eventstore", "readstore", "model",
			},
			Direction: map[string][]string{
				"command":    {"aggregate", "eventstore", "model"},
				"aggregate":  {"event", "model"},
				"event":      {"model"},
				"projection": {"event", "readstore", "model"},
				"eventstore": {"event", "model"},
				"readstore":  {"model"},
				"model":      {},
			},
			PkgRestricted: map[string]bool{"model": true, "event": true},
			InternalTopLevel: map[string]bool{
				"command": true, "aggregate": true, "event": true,
				"projection": true, "eventstore": true, "readstore": true,
				"model": true, "pkg": true,
			},
			LayerDirNames: map[string]bool{
				"command": true, "aggregate": true, "event": true,
				"projection": true, "eventstore": true, "readstore": true,
				"model": true,
			},
		},
		Layout: core.LayoutModel{SharedDir: "pkg"},
		Naming: core.NamingPolicy{
			BannedPkgNames: defaultBannedPkgNames(),
			LegacyPkgNames: defaultLegacyPkgNames(),
		},
		Structure: core.StructurePolicy{
			ModelPath: "model",
			TypePatterns: []core.TypePattern{
				{Dir: "command", FilePrefix: "command", TypeSuffix: "Command", RequireMethod: "Execute"},
				{Dir: "aggregate", FilePrefix: "aggregate", TypeSuffix: "Aggregate", RequireMethod: "Apply"},
			},
			InterfacePatternExclude: map[string]bool{
				"model": true, "event": true, "command": true, "aggregate": true,
			},
		},
	}
	return mustValidatePreset("EventPipeline", arch)
}

func RecommendedEventPipeline() core.RuleSet {
	return recommendedRules(false, false, false, false)
}
