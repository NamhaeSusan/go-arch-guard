package presets

import (
	"github.com/NamhaeSusan/go-arch-guard/core"
)

func CleanArch() core.Architecture {
	domainDir, orchestrationDir, sharedDir := domainLayout()
	arch := core.Architecture{
		Layers: core.LayerModel{
			Sublayers: []string{"handler", "usecase", "entity", "gateway", "infra"},
			Direction: map[string][]string{
				"handler": {"usecase"},
				"usecase": {"entity", "gateway"},
				"entity":  {},
				"gateway": {"entity"},
				"infra":   {"gateway", "entity"},
			},
			PkgRestricted:    map[string]bool{"entity": true},
			InternalTopLevel: domainTopLevel(),
			LayerDirNames: map[string]bool{
				"handler": true, "usecase": true, "entity": true,
				"gateway": true, "infra": true,
				"service": true, "controller": true,
				"store": true, "persistence": true, "domain": true,
			},
			PortLayers:     []string{"gateway"},
			ContractLayers: []string{"gateway"},
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
			ModelPath: "entity",
			InterfacePatternExclude: map[string]bool{
				"handler": true, "entity": true,
			},
		},
	}
	return mustValidatePreset("CleanArch", arch)
}

func RecommendedCleanArch() core.RuleSet {
	return recommendedRules(true, false, false, true)
}
