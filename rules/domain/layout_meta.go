package domain

import (
	"fmt"

	"github.com/NamhaeSusan/go-arch-guard/core"
)

func metaLayoutNotSupported(ruleID, projectRoot, internalRoot string) core.Violation {
	return core.Violation{
		Rule:              "meta.layout-not-supported",
		Message:           fmt.Sprintf("%s requires an internal/-based layout; %s/%s was not found", ruleID, projectRoot, internalRoot),
		Fix:               "use this rule with internal/-based DDD layouts, or remove it from your ruleset for flat layouts",
		DefaultSeverity:   core.Warning,
		EffectiveSeverity: core.Warning,
	}
}

func metaRuleDisabledByConfig(ruleID, reason, fix string) core.Violation {
	return core.Violation{
		Rule:              "meta.rule-disabled-by-config",
		Message:           fmt.Sprintf("%s: %s", ruleID, reason),
		Fix:               fix,
		DefaultSeverity:   core.Warning,
		EffectiveSeverity: core.Warning,
	}
}
