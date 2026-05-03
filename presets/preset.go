package presets

import (
	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/rules/dependency"
	"github.com/NamhaeSusan/go-arch-guard/rules/interfaces"
	"github.com/NamhaeSusan/go-arch-guard/rules/naming"
	"github.com/NamhaeSusan/go-arch-guard/rules/structural"
	"github.com/NamhaeSusan/go-arch-guard/rules/testpolicy"
	"github.com/NamhaeSusan/go-arch-guard/rules/types"
)

type Preset string

const (
	PresetDDD             Preset = "ddd"
	PresetCleanArch       Preset = "cleanarch"
	PresetLayered         Preset = "layered"
	PresetHexagonal       Preset = "hexagonal"
	PresetModularMonolith Preset = "modular_monolith"
	PresetConsumerWorker  Preset = "consumer_worker"
	PresetBatch           Preset = "batch"
	PresetEventPipeline   Preset = "event_pipeline"
)

func defaultBannedPkgNames() []string {
	return []string{"util", "common", "misc", "helper", "shared", "services"}
}

func defaultLegacyPkgNames() []string {
	return []string{"router", "bootstrap"}
}

func domainTopLevel() map[string]bool {
	return map[string]bool{
		"domain":        true,
		"orchestration": true,
		"pkg":           true,
	}
}

func domainLayout() (string, string, string) {
	return "domain", "orchestration", "pkg"
}

func mustValidatePreset(name string, arch core.Architecture) core.Architecture {
	if err := arch.Validate(); err != nil {
		panic("preset " + name + ": " + err.Error())
	}
	return arch
}

func recommendedRules(includeIsolation, includeAlias, includeModelRequired, includeCrossDomain bool) core.RuleSet {
	rules := []core.Rule{}
	if includeIsolation {
		rules = append(rules, dependency.NewIsolation())
	}
	rules = append(rules,
		dependency.NewLayerDirection(),
		dependency.NewBlastRadius(),
		naming.NewNoStutter(),
		naming.NewImplSuffix(),
		naming.NewSnakeCaseFiles(),
		naming.NewNoLayerSuffix(),
		testpolicy.NewNoHandMock(),
		structural.NewRepoFileInterface(),
	)
	if includeAlias {
		rules = append(rules, structural.NewAlias())
	}
	rules = append(rules,
		structural.NewLayerPlacement(),
		structural.NewBannedPackage(),
	)
	if includeModelRequired {
		rules = append(rules, structural.NewModelRequired())
	}
	rules = append(rules,
		structural.NewInternalTopLevel(),
		interfaces.NewPattern(),
		interfaces.NewTooManyMethods(),
		interfaces.NewContainer(),
	)
	if includeCrossDomain {
		rules = append(rules, interfaces.NewCrossDomainAnonymous())
	}
	rules = append(rules,
		naming.NewTypePattern(),
		types.NewNoSetter(),
	)
	return core.NewRuleSet(rules...)
}
