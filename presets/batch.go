package presets

import (
	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/rules/dependency"
	"github.com/NamhaeSusan/go-arch-guard/rules/interfaces"
	"github.com/NamhaeSusan/go-arch-guard/rules/naming"
	"github.com/NamhaeSusan/go-arch-guard/rules/structural"
	"github.com/NamhaeSusan/go-arch-guard/rules/types"
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
			ModelPath:        "model",
			DTOAllowedLayers: []string{"job", "service"},
			TypePatterns: []core.TypePattern{
				{Dir: "job", FilePrefix: "job", TypeSuffix: "Job", RequireMethod: "Run"},
			},
			InterfacePatternExclude: map[string]bool{
				"model": true, "job": true,
			},
		},
	}
	if err := arch.Validate(); err != nil {
		panic("preset Batch: " + err.Error())
	}
	return arch
}

func RecommendedBatch() core.RuleSet {
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
