package rulemeta

import (
	"fmt"

	"github.com/NamhaeSusan/go-arch-guard/core"
)

func Warning(rule, message, fix string) core.Violation {
	return core.Violation{
		Rule:              rule,
		Message:           message,
		Fix:               fix,
		DefaultSeverity:   core.Warning,
		EffectiveSeverity: core.Warning,
	}
}

func RuleDisabledByConfig(ruleID, reason, fix string) core.Violation {
	return Warning("meta.rule-disabled-by-config", fmt.Sprintf("%s: %s", ruleID, reason), fix)
}

func LayoutNotSupported(message, fix string) core.Violation {
	return Warning("meta.layout-not-supported", message, fix)
}

func NoMatchingPackages(message string) core.Violation {
	return Warning("meta.no-matching-packages", message, "verify the module argument matches go.mod")
}
