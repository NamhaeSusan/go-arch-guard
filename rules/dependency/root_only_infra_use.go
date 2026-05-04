package dependency

import (
	"fmt"
	"go/ast"
	"path"
	"strings"

	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/core/analysisutil"
	"golang.org/x/tools/go/packages"
)

const ruleRootOnlyInfraUse = "composition.root-only-infra-use"

type RootOnlyInfraUse struct {
	cfg ruleConfig
}

func NewRootOnlyInfraUse(opts ...Option) *RootOnlyInfraUse {
	return &RootOnlyInfraUse{cfg: newConfig(opts, core.Warning)}
}

func (r *RootOnlyInfraUse) Spec() core.RuleSpec {
	return core.RuleSpec{
		ID:              ruleRootOnlyInfraUse,
		Description:     "domain infra adapters may be imported only from composition roots",
		DefaultSeverity: r.cfg.severity,
		Violations: []core.ViolationSpec{
			{
				ID:              ruleRootOnlyInfraUse,
				Description:     "non-composition package imports a domain infra adapter directly",
				DefaultSeverity: r.cfg.severity,
			},
		},
	}
}

func (r *RootOnlyInfraUse) Check(ctx *core.Context) []core.Violation {
	if ctx == nil {
		return nil
	}
	pkgs := ctx.Pkgs()
	arch := ctx.Arch()
	projectModule := analysisutil.ResolveModuleFromContext(ctx, "")
	if warns := validateModule(pkgs, projectModule); len(warns) > 0 {
		return warns
	}
	if !hasInternalPackages(pkgs, projectModule, arch.Layout.InternalRoot) {
		return []core.Violation{metaLayoutNotSupported(ruleRootOnlyInfraUse, projectModule)}
	}
	if arch.Layout.DomainDir == "" {
		return []core.Violation{metaRuleDisabledByConfig(ruleRootOnlyInfraUse,
			"Layout.DomainDir is empty (flat layout); infra adapter import detection requires a domain directory",
			"set Layout.DomainDir to your domain root, or remove dependency.NewRootOnlyInfraUse() from your RuleSet")}
	}

	checker := infraImportChecker{
		rule:          r,
		ctx:           ctx,
		arch:          arch,
		projectModule: projectModule,
		roots:         r.compositionRoots(arch),
	}
	var violations []core.Violation
	seen := make(map[string]struct{})
	for _, pkg := range pkgs {
		if pkg == nil {
			continue
		}
		for _, v := range checker.checkPackage(pkg) {
			key := rootOnlyInfraUseViolationKey(v)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			violations = append(violations, v)
		}
	}
	return violations
}

func rootOnlyInfraUseViolationKey(v core.Violation) string {
	return fmt.Sprintf("%s:%d:%s:%s", v.File, v.Line, v.Rule, v.Message)
}

func (r *RootOnlyInfraUse) compositionRoots(arch core.Architecture) []string {
	roots := []string{"cmd/..."}
	if arch.Layout.AppDir != "" {
		roots = append(roots, arch.Layout.InternalRoot+"/"+arch.Layout.AppDir+"/...")
	}
	for _, root := range r.cfg.compositionRoots {
		roots = append(roots, core.NormalizeMatchPath(root))
	}
	return roots
}

type infraImportChecker struct {
	rule          *RootOnlyInfraUse
	ctx           *core.Context
	arch          core.Architecture
	projectModule string
	roots         []string
}

func (c infraImportChecker) checkPackage(pkg *packages.Package) []core.Violation {
	var violations []core.Violation
	for _, file := range pkg.Syntax {
		if !c.rule.cfg.includeTestFiles && analysisutil.IsTestFile(file, pkg.Fset) {
			continue
		}
		filePath := c.filePath(pkg, file)
		if c.ctx.IsExcluded(filePath) {
			continue
		}
		for _, imp := range file.Imports {
			importPath := strings.Trim(imp.Path.Value, `"`)
			targetDomain, ok := c.domainInfraTarget(importPath)
			if !ok || c.isAllowedImporter(pkg.PkgPath, filePath, targetDomain) {
				continue
			}
			pos := pkg.Fset.Position(imp.Pos())
			violations = append(violations, c.violation(filePath, pos.Line, pkg.PkgPath, importPath, targetDomain))
		}
	}
	return violations
}

func (c infraImportChecker) domainInfraTarget(pkgPath string) (string, bool) {
	domain, layer, ok := c.domainLayer(pkgPath)
	if !ok {
		return "", false
	}
	return domain, layer == "infra" || strings.HasPrefix(layer, "infra/")
}

func (c infraImportChecker) isAllowedImporter(pkgPath, filePath, targetDomain string) bool {
	rel := analysisutil.ProjectRelativePackagePath(pkgPath, c.projectModule)
	if rel != "" && matchesAnyRoot(rel, c.roots) {
		return true
	}
	if domain, ok := c.domainRoot(rel); ok && domain == targetDomain && path.Base(filePath) == aliasFacadeFileName(c.arch) {
		return true
	}
	domain, layer, ok := c.domainLayer(pkgPath)
	return ok && domain == targetDomain && (layer == "infra" || strings.HasPrefix(layer, "infra/"))
}

func aliasFacadeFileName(arch core.Architecture) string {
	if arch.Naming.AliasFileName != "" {
		return arch.Naming.AliasFileName
	}
	return "alias.go"
}

func (c infraImportChecker) domainRoot(rel string) (string, bool) {
	if rel == "" {
		return "", false
	}
	parts := strings.Split(rel, "/")
	if len(parts) != 3 || parts[0] != c.arch.Layout.InternalRoot || parts[1] != c.arch.Layout.DomainDir {
		return "", false
	}
	return parts[2], true
}

func (c infraImportChecker) domainLayer(pkgPath string) (string, string, bool) {
	rel := analysisutil.ProjectRelativePackagePath(pkgPath, c.projectModule)
	if rel == "" {
		return "", "", false
	}
	parts := strings.Split(rel, "/")
	if len(parts) < 4 || parts[0] != c.arch.Layout.InternalRoot || parts[1] != c.arch.Layout.DomainDir {
		return "", "", false
	}
	return parts[2], strings.Join(parts[3:], "/"), true
}

func matchesAnyRoot(path string, roots []string) bool {
	path = core.NormalizeMatchPath(path)
	for _, root := range roots {
		root = core.NormalizeMatchPath(root)
		if strings.HasSuffix(root, "/...") {
			prefix := strings.TrimSuffix(root, "/...")
			if path == prefix || strings.HasPrefix(path, prefix+"/") {
				return true
			}
			continue
		}
		if path == root {
			return true
		}
	}
	return false
}

func (c infraImportChecker) filePath(pkg *packages.Package, file *ast.File) string {
	if pkg == nil || pkg.Fset == nil || file == nil {
		return ""
	}
	return analysisutil.RelativePathForPackage(pkg, pkg.Fset.Position(file.Pos()).Filename)
}

func (c infraImportChecker) violation(file string, line int, importer, importPath, targetDomain string) core.Violation {
	return core.Violation{
		File:              file,
		Line:              line,
		Rule:              ruleRootOnlyInfraUse,
		Message:           fmt.Sprintf("package %q imports domain %q infra adapter %q outside a composition root", importer, targetDomain, importPath),
		Fix:               "wire infra adapters from internal/" + c.arch.Layout.AppDir + "/ or cmd/... and pass dependencies through app services or ports",
		DefaultSeverity:   c.rule.cfg.severity,
		EffectiveSeverity: c.rule.cfg.severity,
	}
}

var _ core.Rule = (*RootOnlyInfraUse)(nil)
