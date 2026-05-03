package interfaces

import (
	"fmt"
	"strings"

	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/rules/internal/rulemeta"
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
	return rulemeta.LayoutNotSupported(
		fmt.Sprintf("%s requires an internal/-based layout; no internal/ packages found in module %q", ruleID, projectModule),
		"use this rule with internal/-based presets (DDD, CleanArch, Hexagonal, ModularMonolith), or remove it from your ruleset for flat layouts")
}

// metaRuleDisabledByConfig signals that a rule is registered in the RuleSet
// but the supplied core.Architecture configuration prevents it from running
// (whole rule) or makes a sub-check inert (partial). Severity defaults to
// Warning via the runner's meta.* prefix handling.
func metaRuleDisabledByConfig(ruleID, reason, fix string) core.Violation {
	return rulemeta.RuleDisabledByConfig(ruleID, reason, fix)
}
