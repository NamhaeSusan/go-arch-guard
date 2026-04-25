package presets

import (
	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/rules/dependency"
	"github.com/NamhaeSusan/go-arch-guard/rules/interfaces"
	"github.com/NamhaeSusan/go-arch-guard/rules/naming"
	"github.com/NamhaeSusan/go-arch-guard/rules/structural"
	"github.com/NamhaeSusan/go-arch-guard/rules/types"
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
			ModelPath:        "domain",
			DTOAllowedLayers: []string{"handler", "usecase"},
			InterfacePatternExclude: map[string]bool{
				"handler": true, "domain": true,
			},
		},
	}
	if err := arch.Validate(); err != nil {
		panic("preset Hexagonal: " + err.Error())
	}
	return arch
}

func RecommendedHexagonal() core.RuleSet {
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
		interfaces.NewPattern(interfaces.WithMaxMethods(10)),
		interfaces.NewContainer(),
		interfaces.NewCrossDomainAnonymous(),
		types.NewTypePattern(),
		types.NewNoSetter(),
	)
}
