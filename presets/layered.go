package presets

import (
	"github.com/NamhaeSusan/go-arch-guard/core"
)

func Layered() core.Architecture {
	domainDir, orchestrationDir, sharedDir := domainLayout()
	arch := core.Architecture{
		Layers: core.LayerModel{
			Sublayers: []string{"handler", "service", "repository", "model"},
			Direction: map[string][]string{
				"handler":    {"service"},
				"service":    {"repository", "model"},
				"repository": {"model"},
				"model":      {},
			},
			PkgRestricted:    map[string]bool{"model": true},
			InternalTopLevel: domainTopLevel(),
			LayerDirNames: map[string]bool{
				"handler": true, "service": true, "repository": true, "model": true,
				"controller": true, "entity": true, "store": true,
				"persistence": true, "domain": true,
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
			ModelPath: "model",
			InterfacePatternExclude: map[string]bool{
				"handler": true, "model": true,
			},
		},
	}
	return mustValidatePreset("Layered", arch)
}

func RecommendedLayered() core.RuleSet {
	return recommendedRules(true, false, false, true)
}
