package orchestration

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/core/analysisutil"
	"golang.org/x/tools/go/packages"
)

const ruleAliasSignatures = "orchestration.alias-signatures"

type AliasSignatures struct {
	cfg ruleConfig
}

func NewAliasSignatures(opts ...Option) *AliasSignatures {
	return &AliasSignatures{cfg: newConfig(opts, core.Warning)}
}

func (r *AliasSignatures) Spec() core.RuleSpec {
	return core.RuleSpec{
		ID:              ruleAliasSignatures,
		Description:     "orchestration public APIs must not expose domain sub-package types",
		DefaultSeverity: r.cfg.severity,
		Violations: []core.ViolationSpec{
			{ID: ruleAliasSignatures, DefaultSeverity: r.cfg.severity},
		},
	}
}

func (r *AliasSignatures) Check(ctx *core.Context) []core.Violation {
	if ctx == nil {
		return nil
	}
	arch := ctx.Arch()
	if arch.Layout.DomainDir == "" {
		return []core.Violation{metaRuleDisabledByConfig(ruleAliasSignatures,
			"Layout.DomainDir is empty (flat layout); orchestration signature checks require a domain directory",
			"set Layout.DomainDir to your domain root, or remove orchestration.NewAliasSignatures() from your RuleSet")}
	}
	if arch.Layout.OrchestrationDir == "" {
		return []core.Violation{metaRuleDisabledByConfig(ruleAliasSignatures,
			"Layout.OrchestrationDir is empty; orchestration signature checks require an orchestration directory",
			"set Layout.OrchestrationDir, or remove orchestration.NewAliasSignatures() from your RuleSet")}
	}

	projectModule := analysisutil.ResolveModuleFromContext(ctx, "")
	if !hasInternalPackages(ctx.Pkgs(), projectModule, arch.Layout.InternalRoot) {
		return []core.Violation{metaLayoutNotSupported(ruleAliasSignatures, projectModule)}
	}

	var violations []core.Violation
	for _, pkg := range ctx.Pkgs() {
		if !isOrchestrationPackage(pkg.PkgPath, projectModule, arch.Layout) {
			continue
		}
		violations = append(violations, r.checkPackage(ctx, pkg, projectModule, arch)...)
	}
	return violations
}

func (r *AliasSignatures) checkPackage(ctx *core.Context, pkg *packages.Package, projectModule string, arch core.Architecture) []core.Violation {
	var violations []core.Violation
	for _, file := range pkg.Syntax {
		if analysisutil.IsTestFile(file, pkg.Fset) {
			continue
		}
		analysisutil.WalkFuncDecls(file, func(fd *ast.FuncDecl) {
			if fd.Type == nil || !fd.Name.IsExported() {
				return
			}
			violations = append(violations, r.checkFuncDecl(ctx, pkg, file, fd, projectModule, arch)...)
		})
		violations = append(violations, r.checkExportedInterfaceMethods(ctx, pkg, file, projectModule, arch)...)
	}
	return violations
}

func (r *AliasSignatures) checkFuncDecl(ctx *core.Context, pkg *packages.Package, file *ast.File, fd *ast.FuncDecl, projectModule string, arch core.Architecture) []core.Violation {
	skipConstructorServiceAlias := func(expr ast.Expr) bool {
		return r.cfg.allowConstructorServiceAliases && isConstructorServiceAlias(fd, expr, file, projectModule, arch)
	}
	var violations []core.Violation
	violations = append(violations, r.checkFieldList(ctx, pkg, fd.Name.Name, fd.Type.Params, projectModule, arch, skipConstructorServiceAlias)...)
	violations = append(violations, r.checkFieldList(ctx, pkg, fd.Name.Name, fd.Type.Results, projectModule, arch, nil)...)
	return violations
}

func (r *AliasSignatures) checkExportedInterfaceMethods(ctx *core.Context, pkg *packages.Package, file *ast.File, projectModule string, arch core.Architecture) []core.Violation {
	var violations []core.Violation
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.TYPE {
			continue
		}
		for _, spec := range gen.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok || !ts.Name.IsExported() {
				continue
			}
			iface, ok := ts.Type.(*ast.InterfaceType)
			if !ok || iface.Methods == nil {
				continue
			}
			for _, method := range iface.Methods.List {
				fn, ok := method.Type.(*ast.FuncType)
				if !ok {
					continue
				}
				for _, name := range method.Names {
					if !name.IsExported() {
						continue
					}
					apiName := ts.Name.Name + "." + name.Name
					violations = append(violations, r.checkFieldList(ctx, pkg, apiName, fn.Params, projectModule, arch, nil)...)
					violations = append(violations, r.checkFieldList(ctx, pkg, apiName, fn.Results, projectModule, arch, nil)...)
				}
			}
		}
	}
	return violations
}

func (r *AliasSignatures) checkFieldList(ctx *core.Context, pkg *packages.Package, apiName string, fields *ast.FieldList, projectModule string, arch core.Architecture, skip func(ast.Expr) bool) []core.Violation {
	if fields == nil {
		return nil
	}
	var violations []core.Violation
	for _, field := range fields.List {
		relPath := analysisutil.RelativePathForPackage(pkg, pkg.Fset.Position(field.Pos()).Filename)
		if ctx.IsExcluded(relPath) {
			continue
		}
		if skip != nil && skip(field.Type) {
			continue
		}
		for _, leak := range leakedDomainTypes(pkg.TypesInfo, field.Type, projectModule, arch) {
			pos := pkg.Fset.Position(field.Pos())
			violations = append(violations, core.Violation{
				File:              relPath,
				Line:              pos.Line,
				Rule:              ruleAliasSignatures,
				Message:           fmt.Sprintf("orchestration API %q exposes domain %q sublayer %q type %q", apiName, leak.domain, leak.sublayer, leak.name),
				Fix:               "return an orchestration-local DTO or keep domain sub-package types behind app/domain APIs",
				DefaultSeverity:   r.cfg.severity,
				EffectiveSeverity: r.cfg.severity,
			})
		}
	}
	return violations
}

type leakedType struct {
	domain   string
	sublayer string
	name     string
}

func leakedDomainTypes(info *types.Info, expr ast.Expr, projectModule string, arch core.Architecture) []leakedType {
	if info == nil || expr == nil {
		return nil
	}
	t := info.TypeOf(expr)
	if t == nil {
		return nil
	}
	var leaks []leakedType
	visitType(t, func(named *types.Named) {
		obj := named.Obj()
		if obj == nil || obj.Pkg() == nil {
			return
		}
		domain, sublayer := classifyDomainSublayer(obj.Pkg().Path(), projectModule, arch.Layout)
		if domain == "" || sublayer == "" {
			return
		}
		leaks = append(leaks, leakedType{
			domain:   domain,
			sublayer: sublayer,
			name:     obj.Name(),
		})
	})
	return leaks
}

func visitType(t types.Type, visit func(*types.Named)) {
	if t == nil {
		return
	}
	t = types.Unalias(t)
	switch x := t.(type) {
	case *types.Named:
		visit(x)
	case *types.Pointer:
		visitType(x.Elem(), visit)
	case *types.Slice:
		visitType(x.Elem(), visit)
	case *types.Array:
		visitType(x.Elem(), visit)
	case *types.Map:
		visitType(x.Key(), visit)
		visitType(x.Elem(), visit)
	case *types.Chan:
		visitType(x.Elem(), visit)
	case *types.Signature:
		visitTuple(x.Params(), visit)
		visitTuple(x.Results(), visit)
	case *types.Tuple:
		visitTuple(x, visit)
	}
}

func visitTuple(tuple *types.Tuple, visit func(*types.Named)) {
	if tuple == nil {
		return
	}
	for i := 0; i < tuple.Len(); i++ {
		visitType(tuple.At(i).Type(), visit)
	}
}

func isConstructorServiceAlias(fd *ast.FuncDecl, expr ast.Expr, file *ast.File, projectModule string, arch core.Architecture) bool {
	if fd == nil || fd.Recv != nil || !strings.HasPrefix(fd.Name.Name, "New") {
		return false
	}
	sel, ok := selectorAfterPointer(expr)
	if !ok || !strings.HasSuffix(sel.Sel.Name, "Service") {
		return false
	}
	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}
	importPath := analysisutil.ResolveIdentImportPath(file, ident.Name)
	domain, sublayer := classifyDomainSublayer(importPath, projectModule, arch.Layout)
	return domain != "" && sublayer == ""
}

func selectorAfterPointer(expr ast.Expr) (*ast.SelectorExpr, bool) {
	if star, ok := expr.(*ast.StarExpr); ok {
		expr = star.X
	}
	sel, ok := expr.(*ast.SelectorExpr)
	return sel, ok
}

func classifyDomainSublayer(pkgPath, projectModule string, layout core.LayoutModel) (string, string) {
	if pkgPath == "" || projectModule == "" {
		return "", ""
	}
	prefix := strings.TrimSuffix(projectModule, "/") + "/" + layout.InternalRoot + "/" + layout.DomainDir + "/"
	rel, ok := strings.CutPrefix(pkgPath, prefix)
	if !ok {
		return "", ""
	}
	parts := strings.Split(rel, "/")
	if len(parts) == 0 || parts[0] == "" {
		return "", ""
	}
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], strings.Join(parts[1:], "/")
}

func isOrchestrationPackage(pkgPath, projectModule string, layout core.LayoutModel) bool {
	if pkgPath == "" || projectModule == "" || layout.OrchestrationDir == "" {
		return false
	}
	prefix := strings.TrimSuffix(projectModule, "/") + "/" + layout.InternalRoot + "/" + layout.OrchestrationDir
	return pkgPath == prefix || strings.HasPrefix(pkgPath, prefix+"/")
}

func hasInternalPackages(pkgs []*packages.Package, projectModule, internalRoot string) bool {
	if projectModule == "" {
		return false
	}
	prefix := strings.TrimSuffix(projectModule, "/") + "/" + internalRoot + "/"
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
		Fix:               "use this rule with internal/-based domain presets, or remove it from your ruleset for flat layouts",
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

var _ core.Rule = (*AliasSignatures)(nil)
