package composition

import (
	"fmt"
	"strings"

	"github.com/NamhaeSusan/go-arch-guard/core"
	"golang.org/x/tools/go/packages"
)

func hasInternalPackages(pkgs []*packages.Package, projectModule, internalRoot string) bool {
	if projectModule == "" {
		return false
	}
	prefix := projectModule + "/" + internalRoot + "/"
	for _, pkg := range pkgs {
		if pkg != nil && strings.HasPrefix(pkg.PkgPath, prefix) {
			return true
		}
	}
	return false
}

func metaLayoutNotSupported(ruleID, projectModule string) core.Violation {
	return core.Violation{
		Rule:              "meta.layout-not-supported",
		Message:           fmt.Sprintf("%s requires an internal/-based layout; no internal/ packages found in module %q", ruleID, projectModule),
		Fix:               "use this rule with internal/-based presets (DDD, CleanArch, Hexagonal, ModularMonolith), or remove it from your ruleset for flat layouts",
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
