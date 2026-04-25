package interfaces

import (
	"fmt"
	"strings"

	"github.com/NamhaeSusan/go-arch-guard/core"
	"golang.org/x/tools/go/packages"
)

// hasInternalPackages reports whether any loaded package lives under
// <module>/internal/. Pattern and CrossDomainAnonymous classify packages by
// their internal/<DomainDir>/<sublayer> position, so flat-layout projects
// produce no useful violations and should be signaled via metaLayoutNotSupported.
func hasInternalPackages(pkgs []*packages.Package, projectModule string) bool {
	if projectModule == "" {
		return false
	}
	prefix := projectModule + "/internal/"
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
