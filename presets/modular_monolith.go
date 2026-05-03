package presets

import (
	"github.com/NamhaeSusan/go-arch-guard/core"
)

func ModularMonolith() core.Architecture {
	domainDir, orchestrationDir, sharedDir := domainLayout()
	arch := core.Architecture{
		Layers: core.LayerModel{
			Sublayers: []string{"api", "application", "core", "infrastructure"},
			Direction: map[string][]string{
				"api":            {"application"},
				"application":    {"core"},
				"core":           {},
				"infrastructure": {"core"},
			},
			PkgRestricted:    map[string]bool{"core": true},
			InternalTopLevel: domainTopLevel(),
			LayerDirNames: map[string]bool{
				"api": true, "application": true, "core": true,
				"infrastructure": true,
				"controller":     true, "service": true, "entity": true,
				"store": true, "persistence": true,
			},
		},
		Layout: core.LayoutModel{
			DomainDir:        domainDir,
			OrchestrationDir: orchestrationDir,
			SharedDir:        sharedDir,
		},
		Naming: core.NamingPolicy{
			BannedPkgNames: defaultBannedPkgNames(),
			LegacyPkgNames: defaultLegacyPkgNames(),
			AliasFileName:  "alias.go",
		},
		Structure: core.StructurePolicy{
			ModelPath: "core",
			InterfacePatternExclude: map[string]bool{
				"api": true, "core": true,
			},
		},
	}
	return mustValidatePreset("ModularMonolith", arch)
}

func RecommendedModularMonolith() core.RuleSet {
	return recommendedRules(true, false, false, true)
}
