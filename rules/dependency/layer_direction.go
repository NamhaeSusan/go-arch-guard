package dependency

import (
	"fmt"
	"slices"
	"strings"

	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/core/analysisutil"
	"golang.org/x/tools/go/packages"
)

type LayerDirection struct{ severity core.Severity }

func NewLayerDirection(opts ...Option) *LayerDirection {
	r := &LayerDirection{severity: core.Error}
	for _, opt := range opts {
		r.severity = opt.severity()
	}
	return r
}

func (r *LayerDirection) Spec() core.RuleSpec {
	return core.RuleSpec{
		ID:              "dependency.layer-direction",
		Description:     "domain sublayers must import only allowed inner dependencies",
		DefaultSeverity: r.severity,
		Violations: violationSpecs(r.severity,
			"layer.direction",
			"layer.inner-imports-pkg",
			"layer.unknown-sublayer",
		),
	}
}

func (r *LayerDirection) Check(ctx *core.Context) []core.Violation {
	projectModule := analysisutil.ResolveModuleFromContext(ctx, "")
	projectRoot := analysisutil.ResolveRootFromContext(ctx, "")
	if warns := validateModule(ctx.Pkgs(), projectModule); len(warns) > 0 {
		return warns
	}
	internalPrefix := projectModule + "/internal/"

	if ctx.Arch().Layout.DomainDir == "" {
		return r.checkFlat(ctx, projectModule, projectRoot, internalPrefix)
	}
	return r.checkDomain(ctx, projectModule, projectRoot, internalPrefix)
}

func (r *LayerDirection) checkDomain(ctx *core.Context, projectModule, projectRoot, internalPrefix string) []core.Violation {
	arch := ctx.Arch()
	var violations []core.Violation
	for _, pkg := range ctx.Pkgs() {
		if isExcludedPackage(ctx, pkg.PkgPath, projectModule) || !strings.HasPrefix(pkg.PkgPath, internalPrefix) {
			continue
		}

		src := classifyInternalPackage(arch, pkg.PkgPath, internalPrefix)
		if src.Kind != kindDomain {
			continue
		}
		if src.Sublayer != "" && !analysisutil.IsKnownSublayer(arch.Layers, src.Sublayer) {
			violations = append(violations, r.violation(relativePackageFile(pkg), 0,
				"layer.unknown-sublayer",
				fmt.Sprintf("unknown sublayer %q in domain %q", src.Sublayer, src.Domain),
				fmt.Sprintf("use one of the supported sublayers: %v", arch.Layers.Sublayers),
			))
			continue
		}

		for impPath := range pkg.Imports {
			if !strings.HasPrefix(impPath, internalPrefix) {
				continue
			}

			imp := classifyInternalPackage(arch, impPath, internalPrefix)
			switch imp.Kind {
			case kindShared:
				if arch.Layers.PkgRestricted[src.Sublayer] {
					file, line := analysisutil.FindImportPosition(pkg, impPath, projectRoot)
					violations = append(violations, r.violation(file, line,
						"layer.inner-imports-pkg",
						fmt.Sprintf("inner sublayer %q must not import internal/%s in domain %q", src.Sublayer, arch.Layout.SharedDir, src.Domain),
						"keep core and event layers self-contained; move shared concerns outward to app, handler, or infra",
					))
				}
				continue
			case kindOrchestration, kindUnclassified:
				continue
			}

			if imp.Domain != src.Domain {
				continue
			}
			if imp.Sublayer != "" && !analysisutil.IsKnownSublayer(arch.Layers, imp.Sublayer) {
				file, line := analysisutil.FindImportPosition(pkg, impPath, projectRoot)
				violations = append(violations, r.violation(file, line,
					"layer.unknown-sublayer",
					fmt.Sprintf("unknown sublayer %q in domain %q", imp.Sublayer, src.Domain),
					fmt.Sprintf("use one of the supported sublayers: %v", arch.Layers.Sublayers),
				))
				continue
			}
			violations = append(violations, r.checkDirection(pkg, projectRoot, impPath, arch.Layers.Direction, src.Domain, src.Sublayer, imp.Sublayer)...)
		}
	}
	return violations
}

func (r *LayerDirection) checkFlat(ctx *core.Context, projectModule, projectRoot, internalPrefix string) []core.Violation {
	arch := ctx.Arch()
	var violations []core.Violation
	for _, pkg := range ctx.Pkgs() {
		if isExcludedPackage(ctx, pkg.PkgPath, projectModule) || !strings.HasPrefix(pkg.PkgPath, internalPrefix) {
			continue
		}

		src := classifyInternalPackage(arch, pkg.PkgPath, internalPrefix)
		if src.Kind != kindDomain {
			continue
		}

		for impPath := range pkg.Imports {
			if !strings.HasPrefix(impPath, internalPrefix) {
				continue
			}
			imp := classifyInternalPackage(arch, impPath, internalPrefix)
			switch imp.Kind {
			case kindShared:
				if arch.Layers.PkgRestricted[src.Sublayer] {
					file, line := analysisutil.FindImportPosition(pkg, impPath, projectRoot)
					violations = append(violations, r.violation(file, line,
						"layer.inner-imports-pkg",
						fmt.Sprintf("inner sublayer %q must not import internal/%s", src.Sublayer, arch.Layout.SharedDir),
						"keep inner layers self-contained; move shared concerns to an outer layer",
					))
				}
				continue
			case kindOrchestration, kindUnclassified:
				continue
			}
			violations = append(violations, r.checkDirection(pkg, projectRoot, impPath, arch.Layers.Direction, "", src.Sublayer, imp.Sublayer)...)
		}
	}
	return violations
}

func (r *LayerDirection) checkDirection(pkg *packages.Package, projectRoot, impPath string, direction map[string][]string, domain, srcSublayer, impSublayer string) []core.Violation {
	if impSublayer == "" || srcSublayer == impSublayer {
		return nil
	}
	allowed, known := direction[srcSublayer]
	if !known || slices.Contains(allowed, impSublayer) {
		return nil
	}

	file, line := analysisutil.FindImportPosition(pkg, impPath, projectRoot)
	message := fmt.Sprintf("sublayer %q must not import sublayer %q", srcSublayer, impSublayer)
	if domain != "" {
		message = fmt.Sprintf("%s in domain %q", message, domain)
	}
	return []core.Violation{r.violation(file, line,
		"layer.direction",
		message,
		fmt.Sprintf("allowed imports for %q: %v", srcSublayer, allowed),
	)}
}

func (r *LayerDirection) violation(file string, line int, rule, message, fix string) core.Violation {
	return core.Violation{
		File:              file,
		Line:              line,
		Rule:              rule,
		Message:           message,
		Fix:               fix,
		DefaultSeverity:   r.severity,
		EffectiveSeverity: r.severity,
	}
}

func relativePackageFile(pkg *packages.Package) string {
	if len(pkg.GoFiles) == 0 {
		return pkg.PkgPath
	}
	return analysisutil.RelativePathForPackage(pkg, pkg.GoFiles[0])
}
