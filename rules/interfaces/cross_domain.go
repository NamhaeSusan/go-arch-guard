package interfaces

import (
	"fmt"
	"go/ast"
	"sort"
	"strings"

	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/core/analysisutil"
	"golang.org/x/tools/go/packages"
)

type CrossDomainAnonymous struct {
	cfg ruleConfig
}

func NewCrossDomainAnonymous(opts ...Option) *CrossDomainAnonymous {
	return &CrossDomainAnonymous{cfg: newConfig(opts, core.Error)}
}

func (r *CrossDomainAnonymous) Spec() core.RuleSpec {
	return core.RuleSpec{
		ID:              "interfaces.cross-domain-anonymous",
		Description:     "detect inline anonymous interfaces over foreign domain types",
		DefaultSeverity: r.cfg.severity,
		Violations: []core.ViolationSpec{
			{ID: "interface.cross-domain-anonymous", DefaultSeverity: r.cfg.severity},
		},
	}
}

func (r *CrossDomainAnonymous) Check(ctx *core.Context) []core.Violation {
	if ctx == nil {
		return nil
	}
	pkgs := ctx.Pkgs()
	projectModule := analysisutil.ResolveModuleFromContext(ctx, "")
	arch := ctx.Arch()
	if !hasInternalPackages(pkgs, projectModule, arch.Layout.InternalRoot) {
		return []core.Violation{metaLayoutNotSupported("interfaces.cross-domain-anonymous", projectModule)}
	}
	if arch.Layout.DomainDir == "" {
		return nil
	}

	var violations []core.Violation
	for _, pkg := range pkgs {
		violations = append(violations, r.checkPackage(pkg, arch)...)
	}
	return violations
}

func (r *CrossDomainAnonymous) checkPackage(pkg *packages.Package, arch core.Architecture) []core.Violation {
	if isOrchestrationPath(pkg.PkgPath, arch) {
		return nil
	}
	currentDomain := owningDomainForPath(pkg.PkgPath, arch)

	var violations []core.Violation
	for _, file := range pkg.Syntax {
		if analysisutil.IsTestFile(file, pkg.Fset) {
			continue
		}
		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.GenDecl:
				for _, spec := range d.Specs {
					ts, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}
					if _, isIface := ts.Type.(*ast.InterfaceType); isIface {
						continue
					}
					violations = append(violations, r.inspectExpr(ts.Type, file, pkg, currentDomain, arch)...)
				}
			case *ast.FuncDecl:
				if d.Type == nil {
					continue
				}
				violations = append(violations, r.inspectFieldList(d.Type.Params, file, pkg, currentDomain, arch)...)
				violations = append(violations, r.inspectFieldList(d.Type.Results, file, pkg, currentDomain, arch)...)
			}
		}
	}
	return violations
}

func (r *CrossDomainAnonymous) inspectFieldList(fields *ast.FieldList, file *ast.File, pkg *packages.Package, currentDomain string, arch core.Architecture) []core.Violation {
	if fields == nil {
		return nil
	}
	var violations []core.Violation
	for _, f := range fields.List {
		violations = append(violations, r.inspectExpr(f.Type, file, pkg, currentDomain, arch)...)
	}
	return violations
}

func (r *CrossDomainAnonymous) inspectExpr(expr ast.Expr, file *ast.File, pkg *packages.Package, currentDomain string, arch core.Architecture) []core.Violation {
	var violations []core.Violation
	ast.Inspect(expr, func(n ast.Node) bool {
		iface, ok := n.(*ast.InterfaceType)
		if !ok {
			return true
		}
		violations = append(violations, r.checkAnonymousInterface(iface, file, pkg, currentDomain, arch)...)
		return false
	})
	return violations
}

func (r *CrossDomainAnonymous) checkAnonymousInterface(iface *ast.InterfaceType, file *ast.File, pkg *packages.Package, currentDomain string, arch core.Architecture) []core.Violation {
	if iface.Methods == nil || len(iface.Methods.List) == 0 {
		return nil
	}

	hasMethodDecl := false
	crossDomains := make(map[string]bool)
	for _, method := range iface.Methods.List {
		if len(method.Names) == 0 {
			continue
		}
		funcType, ok := method.Type.(*ast.FuncType)
		if !ok {
			continue
		}
		hasMethodDecl = true
		collectCrossDomainRefs(funcType, file, currentDomain, arch, crossDomains)
	}
	if !hasMethodDecl || len(crossDomains) == 0 {
		return nil
	}

	domains := make([]string, 0, len(crossDomains))
	for d := range crossDomains {
		domains = append(domains, d)
	}
	sort.Strings(domains)

	pos := pkg.Fset.Position(iface.Pos())
	return []core.Violation{{
		File:              analysisutil.RelativePathForPackage(pkg, pos.Filename),
		Line:              pos.Line,
		Rule:              "interface.cross-domain-anonymous",
		Message:           fmt.Sprintf("anonymous interface declared in package %q references types from domain(s) %v", pkg.PkgPath, domains),
		Fix:               "move this adapter/abstraction into " + arch.Layout.InternalRoot + "/" + arch.Layout.OrchestrationDir + "/",
		DefaultSeverity:   r.cfg.severity,
		EffectiveSeverity: r.cfg.severity,
	}}
}

func collectCrossDomainRefs(funcType *ast.FuncType, file *ast.File, currentDomain string, arch core.Architecture, out map[string]bool) {
	visit := func(expr ast.Expr) {
		walkTypeExprForDomainRefs(expr, file, currentDomain, arch, out)
	}
	if funcType.Params != nil {
		for _, f := range funcType.Params.List {
			visit(f.Type)
		}
	}
	if funcType.Results != nil {
		for _, f := range funcType.Results.List {
			visit(f.Type)
		}
	}
}

func walkTypeExprForDomainRefs(expr ast.Expr, file *ast.File, currentDomain string, arch core.Architecture, out map[string]bool) {
	ast.Inspect(expr, func(n ast.Node) bool {
		sel, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		ident, ok := sel.X.(*ast.Ident)
		if !ok {
			return true
		}
		importPath := analysisutil.ResolveIdentImportPath(file, ident.Name)
		if importPath == "" {
			return true
		}
		refDomain := owningDomainForPath(importPath, arch)
		if refDomain != "" && refDomain != currentDomain {
			out[refDomain] = true
		}
		return true
	})
}

func owningDomainForPath(pkgPath string, arch core.Architecture) string {
	if arch.Layout.DomainDir == "" {
		return ""
	}
	parts := strings.Split(pkgPath, "/")
	for i := 0; i+2 < len(parts); i++ {
		if parts[i] == arch.Layout.InternalRoot && parts[i+1] == arch.Layout.DomainDir {
			return parts[i+2]
		}
	}
	return ""
}

func isOrchestrationPath(pkgPath string, arch core.Architecture) bool {
	if arch.Layout.OrchestrationDir == "" {
		return false
	}
	parts := strings.Split(pkgPath, "/")
	for i := 0; i+1 < len(parts); i++ {
		if parts[i] == arch.Layout.InternalRoot && parts[i+1] == arch.Layout.OrchestrationDir {
			return true
		}
	}
	return false
}

var _ core.Rule = (*CrossDomainAnonymous)(nil)
