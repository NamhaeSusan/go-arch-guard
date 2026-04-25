package presets

import (
	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/rules/dependency"
	"github.com/NamhaeSusan/go-arch-guard/rules/interfaces"
	"github.com/NamhaeSusan/go-arch-guard/rules/naming"
	"github.com/NamhaeSusan/go-arch-guard/rules/structural"
	"github.com/NamhaeSusan/go-arch-guard/rules/types"
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
			ModelPath:        "model",
			DTOAllowedLayers: []string{"handler", "service"},
			InterfacePatternExclude: map[string]bool{
				"handler": true, "model": true,
			},
		},
	}
	if err := arch.Validate(); err != nil {
		panic("preset Layered: " + err.Error())
	}
	return arch
}

func RecommendedLayered() core.RuleSet {
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
