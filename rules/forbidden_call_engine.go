package rules

import (
	"go/ast"
	"go/types"
	"slices"
	"strings"

	"github.com/NamhaeSusan/go-arch-guard/core/analysisutil"
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

var _ = funcQualifiedName

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
						File:              relFile,
						Line:              pos.Line,
						Rule:              cr.ruleName,
						Message:           cr.message(layer, cr.allowedSlice),
						Fix:               cr.fix(layer, cr.allowedSlice),
						DefaultSeverity:   cfg.Sev,
						EffectiveSeverity: cfg.Sev,
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
// Returns "" when the callee cannot be resolved.
func resolveCalleeID(info *types.Info, call *ast.CallExpr) string {
	return analysisutil.ResolveCalleeID(info, call)
}

func funcQualifiedName(fn *types.Func) string {
	return analysisutil.FuncQualifiedName(fn)
}
