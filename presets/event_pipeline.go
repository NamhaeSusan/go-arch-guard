package presets

import (
	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/rules/dependency"
	"github.com/NamhaeSusan/go-arch-guard/rules/interfaces"
	"github.com/NamhaeSusan/go-arch-guard/rules/naming"
	"github.com/NamhaeSusan/go-arch-guard/rules/structural"
	"github.com/NamhaeSusan/go-arch-guard/rules/types"
)

func EventPipeline() core.Architecture {
	arch := core.Architecture{
		Layers: core.LayerModel{
			Sublayers: []string{
				"command", "aggregate", "event", "projection",
				"eventstore", "readstore", "model",
			},
			Direction: map[string][]string{
				"command":    {"aggregate", "eventstore", "model"},
				"aggregate":  {"event", "model"},
				"event":      {"model"},
				"projection": {"event", "readstore", "model"},
				"eventstore": {"event", "model"},
				"readstore":  {"model"},
				"model":      {},
			},
			PkgRestricted: map[string]bool{"model": true, "event": true},
			InternalTopLevel: map[string]bool{
				"command": true, "aggregate": true, "event": true,
				"projection": true, "eventstore": true, "readstore": true,
				"model": true, "pkg": true,
			},
			LayerDirNames: map[string]bool{
				"command": true, "aggregate": true, "event": true,
				"projection": true, "eventstore": true, "readstore": true,
				"model": true,
			},
		},
		Layout: core.LayoutModel{SharedDir: "pkg"},
		Naming: core.NamingPolicy{
			BannedPkgNames: defaultBannedPkgNames(),
			LegacyPkgNames: defaultLegacyPkgNames(),
		},
		Structure: core.StructurePolicy{
			ModelPath:        "model",
			DTOAllowedLayers: []string{"command", "projection"},
			TypePatterns: []core.TypePattern{
				{Dir: "command", FilePrefix: "command", TypeSuffix: "Command", RequireMethod: "Execute"},
				{Dir: "aggregate", FilePrefix: "aggregate", TypeSuffix: "Aggregate", RequireMethod: "Apply"},
			},
			InterfacePatternExclude: map[string]bool{
				"model": true, "event": true, "command": true, "aggregate": true,
			},
		},
	}
	if err := arch.Validate(); err != nil {
		panic("preset EventPipeline: " + err.Error())
	}
	return arch
}

func RecommendedEventPipeline() core.RuleSet {
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
