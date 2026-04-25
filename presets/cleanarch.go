package presets

import (
	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/rules/dependency"
	"github.com/NamhaeSusan/go-arch-guard/rules/interfaces"
	"github.com/NamhaeSusan/go-arch-guard/rules/naming"
	"github.com/NamhaeSusan/go-arch-guard/rules/structural"
	"github.com/NamhaeSusan/go-arch-guard/rules/types"
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
			ModelPath:        "entity",
			DTOAllowedLayers: []string{"handler", "usecase"},
			InterfacePatternExclude: map[string]bool{
				"handler": true, "entity": true,
			},
		},
	}
	if err := arch.Validate(); err != nil {
		panic("preset CleanArch: " + err.Error())
	}
	return arch
}

func RecommendedCleanArch() core.RuleSet {
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
