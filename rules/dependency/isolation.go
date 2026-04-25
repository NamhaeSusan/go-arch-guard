package dependency

import (
	"fmt"
	"slices"
	"strings"

	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/core/analysisutil"
	"golang.org/x/tools/go/packages"
)

type Isolation struct{ severity core.Severity }

func NewIsolation(opts ...Option) *Isolation {
	cfg := newConfig(opts, core.Error)
	return &Isolation{severity: cfg.severity}
}

func (r *Isolation) Spec() core.RuleSpec {
	return core.RuleSpec{
		ID:              "dependency.isolation",
		Description:     "internal packages must respect domain isolation boundaries",
		DefaultSeverity: r.severity,
		Violations: violationSpecs(r.severity,
			"isolation.cross-domain",
			"isolation.cmd-deep-import",
			"isolation.orchestration-deep-import",
			"isolation.pkg-imports-domain",
			"isolation.pkg-imports-orchestration",
			"isolation.domain-imports-orchestration",
			"isolation.stray-imports-orchestration",
			"isolation.stray-imports-domain",
			"isolation.transport-imports-domain",
			"isolation.transport-imports-orchestration",
			"isolation.transport-imports-unclassified",
		),
	}
}

func (r *Isolation) Check(ctx *core.Context) []core.Violation {
	arch := ctx.Arch()
	layout := arch.Layout
	if layout.DomainDir == "" {
		return nil
	}

	projectModule := analysisutil.ResolveModuleFromContext(ctx, "")
	projectRoot := analysisutil.ResolveRootFromContext(ctx, "")
	pkgs := ctx.Pkgs()
	if warns := validateModule(pkgs, projectModule); len(warns) > 0 {
		return warns
	}
	if !hasInternalPackages(pkgs, projectModule, layout.InternalRoot) {
		return []core.Violation{metaLayoutNotSupported("dependency.isolation", projectModule)}
	}

	internalPrefix := projectModule + "/" + layout.InternalRoot + "/"
	cmdPrefix := projectModule + "/cmd"
	var violations []core.Violation

	for _, pkg := range pkgs {
		if isExcludedPackage(ctx, pkg.PkgPath, projectModule) {
			continue
		}
		if pkg.PkgPath == cmdPrefix || strings.HasPrefix(pkg.PkgPath, cmdPrefix+"/") {
			violations = append(violations, r.checkCmdPackage(pkg, arch, projectRoot, internalPrefix)...)
			continue
		}
		if !strings.HasPrefix(pkg.PkgPath, internalPrefix) {
			continue
		}
		src := classifyInternalPackage(arch, pkg.PkgPath, internalPrefix)
		if src.Kind == kindApp {
			continue
		}
		for impPath := range pkg.Imports {
			if !strings.HasPrefix(impPath, internalPrefix) {
				continue
			}
			imp := classifyInternalPackage(arch, impPath, internalPrefix)
			violations = append(violations, r.checkInternalImport(pkg, src, imp, impPath, arch, projectRoot, internalPrefix)...)
		}
	}

	return violations
}

func (r *Isolation) checkCmdPackage(pkg *packages.Package, arch core.Architecture, projectRoot, internalPrefix string) []core.Violation {
	if !arch.Structure.RequireAlias {
		return nil
	}
	var violations []core.Violation
	layout := arch.Layout
	for impPath := range pkg.Imports {
		if !strings.HasPrefix(impPath, internalPrefix) {
			continue
		}
		imp := classifyInternalPackage(arch, impPath, internalPrefix)
		if imp.Kind != kindDomain {
			continue
		}
		file, line := analysisutil.FindImportPosition(pkg, impPath, projectRoot)
		violations = append(violations, r.violation(file, line,
			"isolation.cmd-deep-import",
			fmt.Sprintf("cmd/ must only import domain alias, not sub-package %q", impPath),
			fmt.Sprintf("import the domain alias package instead: %s%s/%s", internalPrefix, layout.DomainDir, imp.Domain),
		))
	}
	return violations
}

func (r *Isolation) checkInternalImport(pkg *packages.Package, src, imp classified, impPath string, arch core.Architecture, projectRoot, internalPrefix string) []core.Violation {
	if imp.Kind == kindShared && src.Kind != kindShared {
		return nil
	}
	if src.Kind == kindTransport {
		return r.checkTransportImport(pkg, projectRoot, impPath, arch.Layout, imp)
	}
	if isSameDomain(src, imp) {
		return nil
	}
	switch src.Kind {
	case kindOrchestration:
		return r.checkOrchestrationDeepImport(pkg, imp, impPath, arch, projectRoot, internalPrefix)
	case kindShared:
		if v := r.checkSharedImport(pkg, imp, impPath, arch.Layout, projectRoot); v != nil {
			return v
		}
	}
	if imp.Kind == kindOrchestration {
		return r.checkOrchestrationDependent(pkg, src, impPath, arch.Layout, projectRoot)
	}
	if isCrossDomain(src, imp) {
		return r.crossDomainViolation(pkg, src, imp, impPath, arch.Layout, projectRoot)
	}
	if src.Kind == kindUnclassified && isDomainKind(imp.Kind) {
		return r.strayDomainViolation(pkg, imp, impPath, arch.Layout, projectRoot)
	}
	return nil
}

func (r *Isolation) checkOrchestrationDeepImport(pkg *packages.Package, imp classified, impPath string, arch core.Architecture, projectRoot, internalPrefix string) []core.Violation {
	if imp.Kind != kindDomain || !arch.Structure.RequireAlias {
		return nil
	}
	layout := arch.Layout
	label := layout.OrchestrationDir
	if isOrchestrationHandlerWith(arch, pkg.PkgPath, internalPrefix) {
		label = layout.OrchestrationDir + " handler"
	}
	file, line := analysisutil.FindImportPosition(pkg, impPath, projectRoot)
	return []core.Violation{r.violation(file, line,
		"isolation.orchestration-deep-import",
		fmt.Sprintf("%s must only import domain alias, not sub-package %q", label, impPath),
		fmt.Sprintf("import the domain alias package instead: %s%s/%s", internalPrefix, layout.DomainDir, imp.Domain),
	)}
}

func (r *Isolation) checkSharedImport(pkg *packages.Package, imp classified, impPath string, layout core.LayoutModel, projectRoot string) []core.Violation {
	switch {
	case isDomainKind(imp.Kind):
		file, line := analysisutil.FindImportPosition(pkg, impPath, projectRoot)
		return []core.Violation{r.violation(file, line,
			"isolation.pkg-imports-domain",
			fmt.Sprintf("%s/ must not import domain %q", layout.SharedDir, imp.Domain),
			fmt.Sprintf("%s/ should only contain shared utilities with no domain or orchestration dependencies", layout.SharedDir),
		)}
	case imp.Kind == kindOrchestration:
		file, line := analysisutil.FindImportPosition(pkg, impPath, projectRoot)
		return []core.Violation{r.violation(file, line,
			"isolation.pkg-imports-orchestration",
			fmt.Sprintf("%s/ must not import %s", layout.SharedDir, layout.OrchestrationDir),
			fmt.Sprintf("move %s-aware code to %s/%s or cmd/", layout.OrchestrationDir, layout.InternalRoot, layout.OrchestrationDir),
		)}
	}
	return nil
}

func (r *Isolation) checkOrchestrationDependent(pkg *packages.Package, src classified, impPath string, layout core.LayoutModel, projectRoot string) []core.Violation {
	file, line := analysisutil.FindImportPosition(pkg, impPath, projectRoot)
	if isDomainKind(src.Kind) {
		return []core.Violation{r.violation(file, line,
			"isolation.domain-imports-orchestration",
			fmt.Sprintf("domain %q must not import %s", src.Domain, layout.OrchestrationDir),
			fmt.Sprintf("move cross-domain coordination to %s/%s callers instead of domain internals", layout.InternalRoot, layout.OrchestrationDir),
		)}
	}
	return []core.Violation{r.violation(file, line,
		"isolation.stray-imports-orchestration",
		fmt.Sprintf("package %q must not import %s", pkg.PkgPath, layout.OrchestrationDir),
		fmt.Sprintf("only cmd/ and %s/%s may depend on %s", layout.InternalRoot, layout.OrchestrationDir, layout.OrchestrationDir),
	)}
}

func (r *Isolation) crossDomainViolation(pkg *packages.Package, src, imp classified, impPath string, layout core.LayoutModel, projectRoot string) []core.Violation {
	file, line := analysisutil.FindImportPosition(pkg, impPath, projectRoot)
	return []core.Violation{r.violation(file, line,
		"isolation.cross-domain",
		fmt.Sprintf("domain %q must not import domain %q", src.Domain, imp.Domain),
		fmt.Sprintf("use %s/ for cross-domain orchestration or move shared types to %s/", layout.OrchestrationDir, layout.SharedDir),
	)}
}

func (r *Isolation) strayDomainViolation(pkg *packages.Package, imp classified, impPath string, layout core.LayoutModel, projectRoot string) []core.Violation {
	file, line := analysisutil.FindImportPosition(pkg, impPath, projectRoot)
	return []core.Violation{r.violation(file, line,
		"isolation.stray-imports-domain",
		fmt.Sprintf("package %q must not import domain %q", pkg.PkgPath, imp.Domain),
		fmt.Sprintf("move domain orchestration to %s/%s or app wiring to cmd/", layout.InternalRoot, layout.OrchestrationDir),
	)}
}

func isDomainKind(k internalKind) bool {
	return k == kindDomain || k == kindDomainRoot
}

func isSameDomain(a, b classified) bool {
	return isDomainKind(a.Kind) && isDomainKind(b.Kind) && a.Domain != "" && a.Domain == b.Domain
}

func isCrossDomain(a, b classified) bool {
	return isDomainKind(a.Kind) && isDomainKind(b.Kind) &&
		a.Domain != "" && b.Domain != "" && a.Domain != b.Domain
}

func (r *Isolation) checkTransportImport(pkg *packages.Package, projectRoot, impPath string, layout core.LayoutModel, imp classified) []core.Violation {
	switch imp.Kind {
	case kindApp, kindShared, kindTransport, kindCmd:
		return nil
	case kindDomain, kindDomainRoot:
		file, line := analysisutil.FindImportPosition(pkg, impPath, projectRoot)
		return []core.Violation{r.violation(file, line,
			"isolation.transport-imports-domain",
			fmt.Sprintf("transport package %q must not import domain %q directly", pkg.PkgPath, imp.Domain),
			fmt.Sprintf("import %s/ (the app/composition root) instead of domain sub-packages directly", layout.AppDir),
		)}
	case kindOrchestration:
		file, line := analysisutil.FindImportPosition(pkg, impPath, projectRoot)
		return []core.Violation{r.violation(file, line,
			"isolation.transport-imports-orchestration",
			fmt.Sprintf("transport package %q must not import %s directly", pkg.PkgPath, layout.OrchestrationDir),
			fmt.Sprintf("transport layers should only import %s/ (composition root) and %s/ (shared utilities)", layout.AppDir, layout.SharedDir),
		)}
	case kindUnclassified:
		file, line := analysisutil.FindImportPosition(pkg, impPath, projectRoot)
		return []core.Violation{r.violation(file, line,
			"isolation.transport-imports-unclassified",
			fmt.Sprintf("transport package %q must not import unclassified internal package %q", pkg.PkgPath, impPath),
			fmt.Sprintf("move the dependency into %s/%s (expose via Container), %s/%s, or another transport package", layout.InternalRoot, layout.AppDir, layout.InternalRoot, layout.SharedDir),
		)}
	default:
		return nil
	}
}

func (r *Isolation) violation(file string, line int, rule, message, fix string) core.Violation {
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

type internalKind int

const (
	kindDomain internalKind = iota
	kindOrchestration
	kindShared
	kindDomainRoot
	kindCmd
	kindUnclassified
	kindApp
	kindTransport
)

type classified struct {
	Kind     internalKind
	Domain   string
	Sublayer string
}

func classifyInternalPackage(arch core.Architecture, pkgPath, internalPrefix string) classified {
	if !strings.HasPrefix(pkgPath, internalPrefix) {
		return classified{Kind: kindUnclassified}
	}
	layout := arch.Layout
	if layout.SharedDir != "" && isUnderInternalDir(pkgPath, internalPrefix, layout.SharedDir) {
		return classified{Kind: kindShared}
	}
	if layout.AppDir != "" && isUnderInternalDir(pkgPath, internalPrefix, layout.AppDir) {
		return classified{Kind: kindApp}
	}
	if layout.ServerDir != "" {
		rel := strings.TrimPrefix(pkgPath, internalPrefix)
		serverPrefix := layout.ServerDir + "/"
		if strings.HasPrefix(rel, serverPrefix) && strings.TrimPrefix(rel, serverPrefix) != "" {
			return classified{Kind: kindTransport}
		}
	}
	if layout.OrchestrationDir != "" && isUnderInternalDir(pkgPath, internalPrefix, layout.OrchestrationDir) {
		return classified{Kind: kindOrchestration}
	}

	if layout.DomainDir != "" {
		domain := identifyDomainWith(arch, pkgPath, internalPrefix)
		if domain == "" {
			return classified{Kind: kindUnclassified}
		}
		sub := identifySublayerWith(arch, pkgPath, internalPrefix, domain)
		if sub == "" {
			return classified{Kind: kindDomainRoot, Domain: domain}
		}
		return classified{Kind: kindDomain, Domain: domain, Sublayer: sub}
	}

	rel := strings.TrimPrefix(pkgPath, internalPrefix)
	parts := strings.SplitN(rel, "/", 2)
	if parts[0] == "" {
		return classified{Kind: kindUnclassified}
	}
	if slices.Contains(arch.Layers.Sublayers, parts[0]) {
		return classified{Kind: kindDomain, Sublayer: parts[0]}
	}
	return classified{Kind: kindUnclassified}
}

func identifySublayerWith(arch core.Architecture, pkgPath, internalPrefix, domain string) string {
	domainPrefix := internalPrefix + arch.Layout.DomainDir + "/" + domain + "/"
	if !strings.HasPrefix(pkgPath, domainPrefix) {
		return ""
	}
	rel := strings.TrimPrefix(pkgPath, domainPrefix)
	parts := strings.SplitN(rel, "/", 3)
	if len(parts) >= 2 {
		nested := parts[0] + "/" + parts[1]
		if slices.Contains(arch.Layers.Sublayers, nested) {
			return nested
		}
		if hasNestedSublayers(arch, parts[0]) {
			return nested
		}
	}
	return parts[0]
}

func hasNestedSublayers(arch core.Architecture, root string) bool {
	prefix := root + "/"
	for _, sublayer := range arch.Layers.Sublayers {
		if strings.HasPrefix(sublayer, prefix) {
			return true
		}
	}
	return false
}

func identifyDomainWith(arch core.Architecture, pkgPath, internalPrefix string) string {
	rel := strings.TrimPrefix(pkgPath, internalPrefix)
	if !strings.HasPrefix(rel, arch.Layout.DomainDir+"/") {
		return ""
	}
	after := strings.TrimPrefix(rel, arch.Layout.DomainDir+"/")
	parts := strings.SplitN(after, "/", 2)
	return parts[0]
}

func isUnderInternalDir(pkgPath, internalPrefix, dir string) bool {
	if dir == "" {
		return false
	}
	rel := strings.TrimPrefix(pkgPath, internalPrefix)
	return rel == dir || strings.HasPrefix(rel, dir+"/")
}

func isOrchestrationHandlerWith(arch core.Architecture, pkgPath, internalPrefix string) bool {
	rel := strings.TrimPrefix(pkgPath, internalPrefix)
	return strings.HasPrefix(rel, arch.Layout.OrchestrationDir+"/handler")
}

func isExcludedPackage(ctx *core.Context, pkgPath, projectModule string) bool {
	return ctx.IsExcluded(analysisutil.ProjectRelativePackagePath(pkgPath, projectModule))
}

func validateModule(pkgs []*packages.Package, projectModule string) []core.Violation {
	if projectModule == "" {
		return []core.Violation{metaNoMatchingPackages("project module could not be determined - all import checks will be skipped")}
	}
	prefix := projectModule + "/"
	for _, pkg := range pkgs {
		if pkg.PkgPath == projectModule || strings.HasPrefix(pkg.PkgPath, prefix) {
			return nil
		}
	}
	return []core.Violation{metaNoMatchingPackages(fmt.Sprintf("module %q does not match any loaded package - all import checks will be skipped", projectModule))}
}

func metaNoMatchingPackages(message string) core.Violation {
	return core.Violation{
		Rule:              "meta.no-matching-packages",
		Message:           message,
		Fix:               "verify the module argument matches go.mod",
		DefaultSeverity:   core.Warning,
		EffectiveSeverity: core.Warning,
	}
}

// hasInternalPackages reports whether any loaded package lives under
// <module>/<internalRoot>/. The dependency rules in this package operate
// only on internal-root-based layouts; flat-layout projects are signaled
// via metaLayoutNotSupported instead of silently producing zero violations.
func hasInternalPackages(pkgs []*packages.Package, projectModule, internalRoot string) bool {
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

func violationSpecs(severity core.Severity, ids ...string) []core.ViolationSpec {
	specs := make([]core.ViolationSpec, len(ids))
	for i, id := range ids {
		specs[i] = core.ViolationSpec{ID: id, Description: id, DefaultSeverity: severity}
	}
	return specs
}
