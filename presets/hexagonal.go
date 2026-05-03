package presets

import (
	"github.com/NamhaeSusan/go-arch-guard/core"
)

func Hexagonal() core.Architecture {
	domainDir, orchestrationDir, sharedDir := domainLayout()
	arch := core.Architecture{
		Layers: core.LayerModel{
			Sublayers: []string{"handler", "usecase", "port", "domain", "adapter"},
			Direction: map[string][]string{
				"handler": {"usecase"},
				"usecase": {"port", "domain"},
				"port":    {"domain"},
				"domain":  {},
				"adapter": {"port", "domain"},
			},
			PkgRestricted:    map[string]bool{"domain": true},
			InternalTopLevel: domainTopLevel(),
			LayerDirNames: map[string]bool{
				"handler": true, "usecase": true, "port": true,
				"domain": true, "adapter": true,
				"controller": true, "service": true, "entity": true,
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
			ModelPath: "domain",
			InterfacePatternExclude: map[string]bool{
				"handler": true, "domain": true,
			},
		},
	}
	return mustValidatePreset("Hexagonal", arch)
}

func RecommendedHexagonal() core.RuleSet {
	return recommendedRules(true, false, false, true)
}
