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

// scanScope configures which packages the tx-boundary engines include and how
// special-cased packages (cmd/ composition roots, unclassified internal
// packages) are handled.
type scanScope struct {
	// enforceUnclassified: when true, internal packages whose layer can't be
	// determined are treated as non-allowed (layer "") and still checked.
	// When false (default), they are skipped to avoid noise on ad-hoc helpers.
	enforceUnclassified bool
	// enforceCmdRoot: when true, packages under <module>/cmd/... are scanned
	// and bypass the AllowedLayers list — a forbidden call in cmd/ produces
	// a violation. When false (default), cmd/ packages are skipped; this is
	// the backward-compatible behavior for projects that legitimately start
	// transactions from main. Modeled as a dedicated flag (not a synthetic
	// layer name) so user-defined sublayers named "cmd" can't accidentally
	// exempt the composition root.
	enforceCmdRoot bool
}

// cmdRootLayerToken is the reserved layer token used in violation messages
// for composition-root packages. Angle brackets never appear in Go package
// or import paths, so it cannot collide with any user-defined sublayer name.
const cmdRootLayerToken = "<cmd-root>"

// checkForbiddenCallsByLayer walks all CallExprs in packages under
// <module>/internal/ (and optionally <module>/cmd/ when scanScope.enforceCmdRoot
// is true) and emits a violation for each call whose callee ID matches one of
// the rules and whose package layer is outside that rule's AllowedLayers.
// Composition-root packages bypass AllowedLayers and are emitted with the
// reserved layer token cmdRootLayerToken for message clarity.
func checkForbiddenCallsByLayer(
	pkgs []*packages.Package,
	projectModule, projectRoot string,
	m Model,
	cfg Config,
	scope scanScope,
	rules []forbiddenCallRule,
) []Violation {
	if len(rules) == 0 {
		return nil
	}
	projectModule = resolveModule(pkgs, projectModule)
	projectRoot = resolveRoot(pkgs, projectRoot)
	internalPrefix := projectModule + "/internal/"
	cmdPrefix := projectModule + "/cmd/"

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
		decision := scanLayerFor(m, pkg.PkgPath, internalPrefix, cmdPrefix, scope)
		if !decision.scan {
			continue
		}
		if pkg.TypesInfo == nil {
			continue
		}
		layer := decision.layer
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
					// Composition-root packages bypass AllowedLayers: their
					// policy is controlled exclusively by EnforceCmdRoot.
					if !decision.isCmdRoot && cr.allowed[layer] {
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

// scanDecision carries the per-package decision from scanLayerFor.
type scanDecision struct {
	// scan is true if the package should be inspected by the engines.
	scan bool
	// layer is the sublayer name used for violation messages and the
	// AllowedLayers check. For composition-root packages it's cmdRootLayerToken
	// (reserved sentinel) purely for message clarity — the allowed-layers
	// check is skipped for those because isCmdRoot short-circuits it.
	layer string
	// isCmdRoot is true when the package sits under <module>/cmd/...
	// Engines must bypass AllowedLayers for these and always emit.
	isCmdRoot bool
}

// scanLayerFor classifies a package for the tx-boundary engines:
//
//   - Packages under <module>/cmd/... are scan targets with isCmdRoot=true
//     only when enforceCmdRoot is true; otherwise skipped (backward-compat
//     default).
//   - Packages under <module>/internal/... are scanned with their known
//     sublayer.
//   - Internal packages that don't map to any known sublayer (layer == "")
//     are scanned only when enforceUnclassified is true; otherwise skipped
//     to avoid noise on ad-hoc helper packages (testutil, codegen, etc.).
//   - Everything else is skipped.
func scanLayerFor(m Model, pkgPath, internalPrefix, cmdPrefix string, scope scanScope) scanDecision {
	if strings.HasPrefix(pkgPath, cmdPrefix) || pkgPath == strings.TrimSuffix(cmdPrefix, "/") {
		if !scope.enforceCmdRoot {
			return scanDecision{}
		}
		return scanDecision{scan: true, layer: cmdRootLayerToken, isCmdRoot: true}
	}
	if !strings.HasPrefix(pkgPath, internalPrefix) {
		return scanDecision{}
	}
	layer := layerOfPackage(m, pkgPath, internalPrefix)
	if layer == "" && !scope.enforceUnclassified {
		return scanDecision{}
	}
	return scanDecision{scan: true, layer: layer}
}

// layerOfPackage returns the layer/sublayer for the given package path,
// using domain or flat identification depending on the model.
// Returns "" if the package is not under any known layer.
func layerOfPackage(m Model, pkgPath, internalPrefix string) string {
	c := classifyInternalPackage(m, pkgPath, internalPrefix)
	if c.Kind != kindDomain {
		return ""
	}
	if !slices.Contains(m.Sublayers, c.Sublayer) {
		return ""
	}
	return c.Sublayer
}

// resolveCalleeID returns the fully-qualified identifier of a CallExpr's
// callee, in one of these shapes:
//   - "<pkg-path>.<FuncName>"               for package functions
//   - "<pkg-path>.(*<Recv>).<Method>"       for pointer-receiver methods
//   - "<pkg-path>.<Recv>.<Method>"          for value-receiver methods
//
// Generic instantiations (e.g. f[T](), pkg.F[T]()) arrive with Fun wrapped
// in *ast.IndexExpr or *ast.IndexListExpr; we unwrap to the base expression
// before resolving so these call sites are not silently skipped.
//
// Returns "" when the callee cannot be resolved.
func resolveCalleeID(info *types.Info, call *ast.CallExpr) string {
	fun := unwrapIndexExpr(call.Fun)
	switch fun := fun.(type) {
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

// unwrapIndexExpr strips *ast.IndexExpr and *ast.IndexListExpr wrappers that
// appear when a generic function or method is called with explicit type
// arguments (e.g. F[T]() or pkg.F[T1, T2]()). It returns the innermost
// non-index expression, which is either a *ast.SelectorExpr or *ast.Ident.
func unwrapIndexExpr(expr ast.Expr) ast.Expr {
	for {
		switch x := expr.(type) {
		case *ast.IndexExpr:
			expr = x.X
		case *ast.IndexListExpr:
			expr = x.X
		default:
			return expr
		}
	}
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
