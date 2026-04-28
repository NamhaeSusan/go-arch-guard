package interfaces

import (
	"fmt"
	"sort"

	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/core/analysisutil"
)

const defaultMaxMethods = 10

// TooManyMethods flags interfaces whose method count exceeds the configured
// cap (default 10). Override the cap with interfaces.WithMaxMethods(n).
//
// Scope mirrors interfaces.Pattern: the rule iterates every loaded package,
// applies the same Architecture-driven exclusion (isExcludedInterfacePatternPkg),
// and only inspects exported interfaces declared in those packages.
type TooManyMethods struct {
	cfg ruleConfig
}

func NewTooManyMethods(opts ...Option) *TooManyMethods {
	return &TooManyMethods{cfg: newConfig(opts, core.Error)}
}

func (r *TooManyMethods) Spec() core.RuleSpec {
	return core.RuleSpec{
		ID:              "interfaces.too-many-methods",
		Description:     "exported interfaces must not exceed the configured method count",
		DefaultSeverity: r.cfg.severity,
	}
}

func (r *TooManyMethods) maxMethods() int {
	if r.cfg.maxMethods <= 0 {
		return defaultMaxMethods
	}
	return r.cfg.maxMethods
}

func (r *TooManyMethods) Check(ctx *core.Context) []core.Violation {
	if ctx == nil {
		return nil
	}
	pkgs := ctx.Pkgs()
	projectModule := analysisutil.ResolveModuleFromContext(ctx, "")
	arch := ctx.Arch()
	if !hasInternalPackages(pkgs, projectModule, arch.Layout.InternalRoot) {
		return []core.Violation{metaLayoutNotSupported("interfaces.too-many-methods", projectModule)}
	}

	cap := r.maxMethods()
	var violations []core.Violation
	for _, pkg := range pkgs {
		if isExcludedInterfacePatternPkg(arch, pkg) {
			continue
		}
		ifaces := collectExportedInterfacesFromPkg(pkg)
		if len(ifaces) == 0 {
			continue
		}
		names := make([]string, 0, len(ifaces))
		for name := range ifaces {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			iface := ifaces[name]
			count := iface.Methods.NumFields()
			if count <= cap {
				continue
			}
			pos := pkg.Fset.Position(iface.Pos())
			violations = append(violations, core.Violation{
				File:              analysisutil.RelativePathForPackage(pkg, pos.Filename),
				Line:              pos.Line,
				Rule:              "interfaces.too-many-methods",
				Message:           fmt.Sprintf("interface %q has %d methods, expected at most %d", name, count, cap),
				Fix:               "split the interface by consumer needs",
				DefaultSeverity:   r.cfg.severity,
				EffectiveSeverity: r.cfg.severity,
			})
		}
	}
	return violations
}

var _ core.Rule = (*TooManyMethods)(nil)
