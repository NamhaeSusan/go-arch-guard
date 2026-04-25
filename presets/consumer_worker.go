package presets

import (
	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/rules/dependency"
	"github.com/NamhaeSusan/go-arch-guard/rules/interfaces"
	"github.com/NamhaeSusan/go-arch-guard/rules/naming"
	"github.com/NamhaeSusan/go-arch-guard/rules/structural"
	"github.com/NamhaeSusan/go-arch-guard/rules/types"
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
			ModelPath:        "model",
			DTOAllowedLayers: []string{"worker", "service"},
			TypePatterns: []core.TypePattern{
				{Dir: "worker", FilePrefix: "worker", TypeSuffix: "Worker", RequireMethod: "Process"},
			},
			InterfacePatternExclude: map[string]bool{
				"model": true, "worker": true,
			},
		},
	}
	if err := arch.Validate(); err != nil {
		panic("preset ConsumerWorker: " + err.Error())
	}
	return arch
}

func RecommendedConsumerWorker() core.RuleSet {
	return core.NewRuleSet(
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
		types.NewTypePattern(),
		types.NewNoSetter(),
	)
}
