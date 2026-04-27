package interfaces

import (
	"fmt"
	"strings"

	"github.com/NamhaeSusan/go-arch-guard/core"
	"golang.org/x/tools/go/packages"
)

// hasInternalPackages reports whether any loaded package lives under
// <module>/<internalRoot>/. Pattern and CrossDomainAnonymous classify
// packages by their <internalRoot>/<DomainDir>/<sublayer> position, so
// flat-layout projects produce no useful violations and should be
// signaled via metaLayoutNotSupported.
func hasInternalPackages(pkgs []*packages.Package, projectModule, internalRoot string) bool {
	if projectModule == "" {
		return false
	}
	prefix := projectModule + "/" + internalRoot + "/"
	for _, pkg := range pkgs {
		if strings.HasPrefix(pkg.PkgPath, prefix) {
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

// metaRuleDisabledByConfig signals that a rule is registered in the RuleSet
// but the supplied core.Architecture configuration prevents it from running
// (whole rule) or makes a sub-check inert (partial). Severity defaults to
// Warning via the runner's meta.* prefix handling.
func metaRuleDisabledByConfig(ruleID, reason, fix string) core.Violation {
	return core.Violation{
		Rule:              "meta.rule-disabled-by-config",
		Message:           fmt.Sprintf("%s: %s", ruleID, reason),
		Fix:               fix,
		DefaultSeverity:   core.Warning,
		EffectiveSeverity: core.Warning,
	}
}
