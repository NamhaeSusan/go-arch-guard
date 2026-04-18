package rules

import (
	"go/ast"
	"go/types"
	"slices"
	"strings"

	"golang.org/x/tools/go/packages"
)

// forbiddenCallRule defines a call-site gating rule: any CallExpr whose
// resolved callee matches Symbols and whose containing package layer is
// NOT in AllowedLayers produces a Violation with RuleName/Message/Fix.
type forbiddenCallRule struct {
	Symbols       []string
	AllowedLayers []string
	RuleName      string
	// Message and Fix receive the offending layer and the allowed-layer list.
	// Typed callbacks replace former fmt-template strings so misuse fails at
	// compile time instead of producing %!(EXTRA...) output at runtime.
	Message func(layer string, allowed []string) string
	Fix     func(layer string, allowed []string) string
}

// checkForbiddenCallsByLayer walks all CallExprs in internal packages and
// emits a violation for each call whose callee ID matches one of the rules
// and whose package layer is outside that rule's AllowedLayers.
func checkForbiddenCallsByLayer(
	pkgs []*packages.Package,
	projectModule, projectRoot string,
	m Model,
	cfg Config,
	rules []forbiddenCallRule,
) []Violation {
	if len(rules) == 0 {
		return nil
	}
	projectModule = resolveModule(pkgs, projectModule)
	projectRoot = resolveRoot(pkgs, projectRoot)
	internalPrefix := projectModule + "/internal/"

	type compiledRule struct {
		allowed      map[string]bool
		ruleName     string
		message      func(string, []string) string
		fix          func(string, []string) string
		allowedSlice []string
	}
	byName := map[string][]compiledRule{}
	for _, r := range rules {
		cr := compiledRule{
			allowed:      map[string]bool{},
			ruleName:     r.RuleName,
			message:      r.Message,
			fix:          r.Fix,
			allowedSlice: r.AllowedLayers,
		}
		for _, l := range r.AllowedLayers {
			cr.allowed[l] = true
		}
		for _, s := range r.Symbols {
			byName[s] = append(byName[s], cr)
		}
	}

	var violations []Violation
	for _, pkg := range pkgs {
		if isExcludedPackage(cfg, pkg.PkgPath, projectModule) {
			continue
		}
		if !strings.HasPrefix(pkg.PkgPath, internalPrefix) {
			continue
		}
		layer := layerOfPackage(m, pkg.PkgPath, internalPrefix)
		if layer == "" {
			continue
		}
		if pkg.TypesInfo == nil {
			continue
		}
		for _, file := range pkg.Syntax {
			ast.Inspect(file, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}
				id := resolveCalleeID(pkg.TypesInfo, call)
				if id == "" {
					return true
				}
				crs, ok := byName[id]
				if !ok {
					return true
				}
				pos := pkg.Fset.Position(call.Pos())
				relFile := relPathFromRoot(projectRoot, pos.Filename)
				for _, cr := range crs {
					if cr.allowed[layer] {
						continue
					}
					violations = append(violations, Violation{
						File:     relFile,
						Line:     pos.Line,
						Rule:     cr.ruleName,
						Message:  cr.message(layer, cr.allowedSlice),
						Fix:      cr.fix(layer, cr.allowedSlice),
						Severity: cfg.Sev,
					})
				}
				return true
			})
		}
	}
	return violations
}

// layerOfPackage returns the layer/sublayer for the given package path,
// using domain or flat identification depending on the model.
// Returns "" if the package is not under any known layer.
func layerOfPackage(m Model, pkgPath, internalPrefix string) string {
	if !strings.HasPrefix(pkgPath, internalPrefix) {
		return ""
	}
	if m.DomainDir != "" {
		domain := identifyDomainWith(m, pkgPath, internalPrefix)
		if domain == "" {
			return ""
		}
		sub := identifySublayerWith(m, pkgPath, internalPrefix, domain)
		if sub == "" {
			return ""
		}
		if !slices.Contains(m.Sublayers, sub) {
			return ""
		}
		return sub
	}
	return identifyFlatSublayer(m, pkgPath, internalPrefix)
}

// resolveCalleeID returns the fully-qualified identifier of a CallExpr's
// callee, in one of these shapes:
//   - "<pkg-path>.<FuncName>"               for package functions
//   - "<pkg-path>.(*<Recv>).<Method>"       for pointer-receiver methods
//   - "<pkg-path>.<Recv>.<Method>"          for value-receiver methods
//
// Returns "" when the callee cannot be resolved.
func resolveCalleeID(info *types.Info, call *ast.CallExpr) string {
	switch fun := call.Fun.(type) {
	case *ast.SelectorExpr:
		if sel, ok := info.Selections[fun]; ok && sel != nil {
			if fn, ok := sel.Obj().(*types.Func); ok {
				return funcQualifiedName(fn)
			}
		}
		if obj := info.Uses[fun.Sel]; obj != nil {
			if fn, ok := obj.(*types.Func); ok {
				return funcQualifiedName(fn)
			}
		}
	case *ast.Ident:
		if obj := info.Uses[fun]; obj != nil {
			if fn, ok := obj.(*types.Func); ok {
				return funcQualifiedName(fn)
			}
		}
	}
	return ""
}

func funcQualifiedName(fn *types.Func) string {
	sig, ok := fn.Type().(*types.Signature)
	if !ok {
		return ""
	}
	pkg := fn.Pkg()
	if pkg == nil {
		return ""
	}
	if sig.Recv() == nil {
		return pkg.Path() + "." + fn.Name()
	}
	recv := sig.Recv().Type()
	if ptr, ok := recv.(*types.Pointer); ok {
		if named, ok := ptr.Elem().(*types.Named); ok {
			return pkg.Path() + ".(*" + named.Obj().Name() + ")." + fn.Name()
		}
	}
	if named, ok := recv.(*types.Named); ok {
		return pkg.Path() + "." + named.Obj().Name() + "." + fn.Name()
	}
	return ""
}
