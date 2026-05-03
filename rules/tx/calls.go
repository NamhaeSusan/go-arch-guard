package tx

import (
	"fmt"
	"go/ast"
	"slices"
	"strings"

	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/core/analysisutil"
	"golang.org/x/tools/go/packages"
)

const (
	forbiddenCallRuleID    = "tx.forbidden-call"
	mandatoryWrapperRuleID = "tx.mandatory-wrapper"
)

type ForbiddenCall struct {
	Symbols       []string
	AllowedLayers []string
}

type MandatoryWrapper struct {
	Symbols       []string
	AllowedLayers []string
	ReplaceWith   string
}

type CallOption func(*callRule)

func WithCallSeverity(s core.Severity) CallOption {
	return func(r *callRule) {
		r.severity = s
	}
}

type callPolicy struct {
	symbols       []string
	allowedLayers []string
	replaceWith   string
}

type symbolCall struct {
	pkg    *packages.Package
	layer  string
	relPkg string
	symbol string
	call   *ast.CallExpr
}

type callRule struct {
	id          string
	description string
	policies    []callPolicy
	severity    core.Severity
	message     func(symbol, layer string, allowed []string, replacement string) (string, string)
}

func NewForbiddenCalls(calls []ForbiddenCall, opts ...CallOption) core.Rule {
	policies := make([]callPolicy, 0, len(calls))
	for _, call := range calls {
		policies = append(policies, callPolicy{
			symbols:       slices.Clone(call.Symbols),
			allowedLayers: slices.Clone(call.AllowedLayers),
		})
	}
	return newCallRule(forbiddenCallRuleID, "forbid configured calls outside allowed layers", policies, forbiddenCallMessage, opts...)
}

func NewMandatoryWrapper(wrappers []MandatoryWrapper, opts ...CallOption) core.Rule {
	policies := make([]callPolicy, 0, len(wrappers))
	for _, wrapper := range wrappers {
		policies = append(policies, callPolicy{
			symbols:       slices.Clone(wrapper.Symbols),
			allowedLayers: slices.Clone(wrapper.AllowedLayers),
			replaceWith:   wrapper.ReplaceWith,
		})
	}
	return newCallRule(mandatoryWrapperRuleID, "require configured calls to go through project wrappers", policies, mandatoryWrapperMessage, opts...)
}

func newCallRule(id, description string, policies []callPolicy, message func(string, string, []string, string) (string, string), opts ...CallOption) *callRule {
	r := &callRule{
		id:          id,
		description: description,
		policies:    policies,
		severity:    core.Error,
		message:     message,
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (r *callRule) Spec() core.RuleSpec {
	return core.RuleSpec{
		ID:              r.id,
		Description:     r.description,
		DefaultSeverity: r.severity,
		Violations: []core.ViolationSpec{{
			ID:              r.id,
			Description:     r.description,
			DefaultSeverity: r.severity,
		}},
	}
}

func (r *callRule) Check(ctx *core.Context) []core.Violation {
	if ctx == nil {
		return nil
	}
	if len(r.policies) == 0 || !hasConfiguredSymbols(r.policies) {
		return []core.Violation{metaRuleDisabledByConfig(r.id,
			"no call symbols configured; call enforcement skipped",
			"configure at least one symbol, or remove this tx call rule from your RuleSet")}
	}

	module := analysisutil.ResolveModuleFromContext(ctx, "")
	if module == "" {
		return nil
	}
	var violations []core.Violation
	checkSymbolCalls(ctx, false, func(hit symbolCall) {
		for _, policy := range r.policies {
			if !slices.Contains(policy.symbols, hit.symbol) || isAllowedCallSite(ctx.Arch().Layout.InternalRoot, policy.allowedLayers, hit.layer, hit.relPkg) {
				continue
			}
			pos := hit.pkg.Fset.Position(hit.call.Pos())
			message, fix := r.message(hit.symbol, hit.layer, policy.allowedLayers, policy.replaceWith)
			violations = append(violations, core.Violation{
				File:              analysisutil.RelPathFromRoot(ctx.Root(), pos.Filename),
				Line:              pos.Line,
				Rule:              r.id,
				Message:           message,
				Fix:               fix,
				DefaultSeverity:   r.severity,
				EffectiveSeverity: r.severity,
			})
		}
	})
	return violations
}

func hasConfiguredSymbols(policies []callPolicy) bool {
	for _, policy := range policies {
		if len(policy.symbols) > 0 {
			return true
		}
	}
	return false
}

func isAllowedCallSite(internalRoot string, allowed []string, layer, relPkg string) bool {
	if internalRoot == "" {
		internalRoot = "internal"
	}
	relWithoutRoot := strings.TrimPrefix(relPkg, internalRoot+"/")
	for _, item := range allowed {
		item = strings.Trim(strings.TrimSpace(item), "/")
		if item == "" {
			continue
		}
		if item == layer {
			return true
		}
		if strings.Contains(item, "/") &&
			(relPkg == item || strings.HasPrefix(relPkg, item+"/") ||
				relWithoutRoot == item || strings.HasPrefix(relWithoutRoot, item+"/")) {
			return true
		}
	}
	return false
}

func forbiddenCallMessage(symbol, layer string, allowed []string, _ string) (string, string) {
	return fmt.Sprintf("call to %q is forbidden in layer %q; allowed layers/packages: %v", symbol, layer, allowed),
		fmt.Sprintf("move this call into an allowed layer/package: %v", allowed)
}

func mandatoryWrapperMessage(symbol, layer string, allowed []string, replacement string) (string, string) {
	if replacement == "" {
		replacement = "a project wrapper"
	}
	return fmt.Sprintf("call to %q in layer %q must go through %s", symbol, layer, replacement),
		fmt.Sprintf("replace the direct call with %s; direct calls are only allowed in: %v", replacement, allowed)
}

func walkInternalPackages(ctx *core.Context, visit func(*packages.Package, string)) {
	walkInternalPackagesByMode(ctx, true, visit)
}

func walkInternalPackagesByMode(ctx *core.Context, layeredOnly bool, visit func(*packages.Package, string)) {
	module := analysisutil.ResolveModuleFromContext(ctx, "")
	if module == "" {
		return
	}
	internalPrefix := module + "/" + ctx.Arch().Layout.InternalRoot + "/"
	for _, pkg := range ctx.Pkgs() {
		if !strings.HasPrefix(pkg.PkgPath, internalPrefix) {
			continue
		}
		if ctx.IsExcluded(analysisutil.ProjectRelativePackagePath(pkg.PkgPath, module)) {
			continue
		}
		layer := packageLayer(ctx.Arch(), pkg.PkgPath, internalPrefix)
		if layer == "" && layeredOnly {
			continue
		}
		visit(pkg, layer)
	}
}

func checkSymbolCalls(ctx *core.Context, layeredOnly bool, visit func(symbolCall)) {
	module := analysisutil.ResolveModuleFromContext(ctx, "")
	walkInternalPackagesByMode(ctx, layeredOnly, func(pkg *packages.Package, layer string) {
		if pkg.TypesInfo == nil {
			return
		}
		relPkg := analysisutil.ProjectRelativePackagePath(pkg.PkgPath, module)
		for _, file := range pkg.Syntax {
			ast.Inspect(file, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}
				symbol := analysisutil.ResolveCalleeID(pkg.TypesInfo, call)
				if symbol != "" {
					visit(symbolCall{pkg: pkg, layer: layer, relPkg: relPkg, symbol: symbol, call: call})
				}
				return true
			})
		}
	})
}
