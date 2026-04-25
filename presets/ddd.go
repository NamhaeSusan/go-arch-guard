package presets

import (
	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/rules/dependency"
	"github.com/NamhaeSusan/go-arch-guard/rules/interfaces"
	"github.com/NamhaeSusan/go-arch-guard/rules/naming"
	"github.com/NamhaeSusan/go-arch-guard/rules/structural"
	"github.com/NamhaeSusan/go-arch-guard/rules/types"
)

func DDD() core.Architecture {
	domainDir, orchestrationDir, sharedDir := domainLayout()
	arch := core.Architecture{
		Layers: core.LayerModel{
			Sublayers: []string{
				"handler", "app", "core", "core/model",
				"core/repo", "core/svc", "event", "infra",
			},
			Direction: map[string][]string{
				"handler":    {"app"},
				"app":        {"core/model", "core/repo", "core/svc", "event"},
				"core":       {"core/model"},
				"core/model": {},
				"core/repo":  {"core/model"},
				"core/svc":   {"core/model"},
				"event":      {"core/model"},
				"infra":      {"core/repo", "core/model", "event"},
			},
			PkgRestricted: map[string]bool{
				"core": true, "core/model": true,
				"core/repo": true, "core/svc": true, "event": true,
			},
			InternalTopLevel: map[string]bool{
				"domain": true, "orchestration": true, "pkg": true,
				"app": true, "server": true,
			},
			LayerDirNames: map[string]bool{
				"handler": true, "app": true, "core": true,
				"model": true, "repo": true, "svc": true,
				"event": true, "infra": true,
				"service": true, "controller": true,
				"entity": true, "store": true, "persistence": true,
				"domain": true,
			},
			PortLayers:     []string{"core/repo"},
			ContractLayers: []string{"core/repo", "core/svc"},
		},
		Layout: core.LayoutModel{
			DomainDir:        domainDir,
			OrchestrationDir: orchestrationDir,
			SharedDir:        sharedDir,
			AppDir:           "app",
			ServerDir:        "server",
		},
		Naming: core.NamingPolicy{
			BannedPkgNames: defaultBannedPkgNames(),
			LegacyPkgNames: defaultLegacyPkgNames(),
			AliasFileName:  "alias.go",
		},
		Structure: core.StructurePolicy{
			RequireAlias:     true,
			RequireModel:     true,
			ModelPath:        "core/model",
			DTOAllowedLayers: []string{"handler", "app"},
			InterfacePatternExclude: map[string]bool{
				"handler": true, "app": true, "core/model": true, "core/repo": true, "event": true,
			},
		},
	}
	if err := arch.Validate(); err != nil {
		panic("preset DDD: " + err.Error())
	}
	return arch
}

func RecommendedDDD() core.RuleSet {
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
		structural.NewAlias(),
		structural.NewPlacement(),
		structural.NewBannedPackage(),
		structural.NewModelRequired(),
		structural.NewInternalTopLevel(),
		interfaces.NewPattern(interfaces.WithMaxMethods(10)),
		interfaces.NewContainer(),
		interfaces.NewCrossDomainAnonymous(),
		types.NewTypePattern(),
		types.NewNoSetter(),
	)
}
