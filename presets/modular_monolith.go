package presets

import (
	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/rules/dependency"
	"github.com/NamhaeSusan/go-arch-guard/rules/interfaces"
	"github.com/NamhaeSusan/go-arch-guard/rules/naming"
	"github.com/NamhaeSusan/go-arch-guard/rules/structural"
	"github.com/NamhaeSusan/go-arch-guard/rules/types"
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
			ModelPath:        "core",
			DTOAllowedLayers: []string{"api", "application"},
			InterfacePatternExclude: map[string]bool{
				"api": true, "core": true,
			},
		},
	}
	if err := arch.Validate(); err != nil {
		panic("preset ModularMonolith: " + err.Error())
	}
	return arch
}

func RecommendedModularMonolith() core.RuleSet {
	return core.NewRuleSet(
		dependency.NewIsolation(),
		dependency.NewLayerDirection(),
		dependency.NewBlastRadius(),
		naming.NewNoStutter(),
		naming.NewImplSuffix(),
		naming.NewSnakeCaseFiles(),
		naming.NewNoLayerSuffix(),
		naming.NewNoHandMock(),
		naming.NewRepoFileInterface(),
		structural.NewPlacement(),
		structural.NewBannedPackage(),
		structural.NewInternalTopLevel(),
		interfaces.NewPattern(),
		interfaces.NewContainer(),
		interfaces.NewCrossDomainAnonymous(),
		types.NewTypePattern(),
		types.NewNoSetter(),
	)
}
