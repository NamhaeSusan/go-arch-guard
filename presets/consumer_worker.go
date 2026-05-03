package presets

import (
	"github.com/NamhaeSusan/go-arch-guard/core"
)

func ConsumerWorker() core.Architecture {
	arch := core.Architecture{
		Layers: core.LayerModel{
			Sublayers: []string{"worker", "service", "store", "model"},
			Direction: map[string][]string{
				"worker":  {"service", "model"},
				"service": {"store", "model"},
				"store":   {"model"},
				"model":   {},
			},
			PkgRestricted: map[string]bool{"model": true},
			InternalTopLevel: map[string]bool{
				"worker": true, "service": true,
				"store": true, "model": true, "pkg": true,
			},
			LayerDirNames: map[string]bool{
				"worker": true, "service": true, "store": true, "model": true,
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
				{Dir: "worker", FilePrefix: "worker", TypeSuffix: "Worker", RequireMethod: "Process"},
			},
			InterfacePatternExclude: map[string]bool{
				"model": true, "worker": true,
			},
		},
	}
	return mustValidatePreset("ConsumerWorker", arch)
}

func RecommendedConsumerWorker() core.RuleSet {
	return recommendedRules(false, false, false, false)
}
