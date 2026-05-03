package presets

import (
	"github.com/NamhaeSusan/go-arch-guard/core"
)

func Batch() core.Architecture {
	arch := core.Architecture{
		Layers: core.LayerModel{
			Sublayers: []string{"job", "service", "store", "model"},
			Direction: map[string][]string{
				"job":     {"service", "model"},
				"service": {"store", "model"},
				"store":   {"model"},
				"model":   {},
			},
			PkgRestricted: map[string]bool{"model": true},
			InternalTopLevel: map[string]bool{
				"job": true, "service": true,
				"store": true, "model": true, "pkg": true,
			},
			LayerDirNames: map[string]bool{
				"job": true, "service": true, "store": true, "model": true,
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
				{Dir: "job", FilePrefix: "job", TypeSuffix: "Job", RequireMethod: "Run"},
			},
			InterfacePatternExclude: map[string]bool{
				"model": true, "job": true,
			},
		},
	}
	return mustValidatePreset("Batch", arch)
}

func RecommendedBatch() core.RuleSet {
	return recommendedRules(false, false, false, false)
}
