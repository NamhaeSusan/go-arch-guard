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
	ruleID                 = "tx.boundary"
	startOutsideLayerID    = "tx.start-outside-allowed-layer"
	typeInSignatureID      = "tx.type-in-signature"
	defaultAllowedTxLayer  = "app"
	defaultViolationDetail = "transaction boundary rule"
)

type Config struct {
	StartSymbols  []string
	Types         []string
	AllowedLayers []string
}

type Boundary struct {
	cfg      Config
	severity core.Severity
}

type Option func(*Boundary)

func New(cfg Config, opts ...Option) *Boundary {
	r := &Boundary{
		cfg: Config{
			StartSymbols:  slices.Clone(cfg.StartSymbols),
			Types:         slices.Clone(cfg.Types),
			AllowedLayers: slices.Clone(cfg.AllowedLayers),
		},
		severity: core.Error,
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func WithSeverity(s core.Severity) Option {
	return func(r *Boundary) {
		r.severity = s
	}
}

func (r *Boundary) Spec() core.RuleSpec {
	return core.RuleSpec{
		ID:              ruleID,
		Description:     defaultViolationDetail,
		DefaultSeverity: r.severity,
		Violations: []core.ViolationSpec{
			{
				ID:              startOutsideLayerID,
				Description:     "transaction start call outside allowed layer",
				DefaultSeverity: r.severity,
			},
			{
				ID:              typeInSignatureID,
				Description:     "transaction type in function signature outside allowed layer",
				DefaultSeverity: r.severity,
			},
		},
	}
}

func (r *Boundary) Check(ctx *core.Context) []core.Violation {
	if ctx == nil || (len(r.cfg.StartSymbols) == 0 && len(r.cfg.Types) == 0) {
		return nil
	}

	allowed := r.allowedLayers()
	var violations []core.Violation
	violations = append(violations, r.checkStartCalls(ctx, allowed)...)
	violations = append(violations, r.checkSignatureTypes(ctx, allowed)...)
	return violations
}

func (r *Boundary) allowedLayers() []string {
	if len(r.cfg.AllowedLayers) > 0 {
		return r.cfg.AllowedLayers
	}
	return []string{defaultAllowedTxLayer}
}

func (r *Boundary) checkStartCalls(ctx *core.Context, allowed []string) []core.Violation {
	if len(r.cfg.StartSymbols) == 0 {
		return nil
	}
	wanted := stringSet(r.cfg.StartSymbols)
	allowedSet := stringSet(allowed)

	var violations []core.Violation
	r.walkInternalPackages(ctx, func(pkg *packages.Package, layer string) {
		if allowedSet[layer] || pkg.TypesInfo == nil {
			return
		}
		for _, file := range pkg.Syntax {
			ast.Inspect(file, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}
				if !wanted[analysisutil.ResolveCalleeID(pkg.TypesInfo, call)] {
					return true
				}
				pos := pkg.Fset.Position(call.Pos())
				violations = append(violations, r.violation(
					startOutsideLayerID,
					analysisutil.RelPathFromRoot(ctx.Root(), pos.Filename),
					pos.Line,
					fmt.Sprintf("transaction must not start in layer %q; allowed layers: %v", layer, allowed),
					fmt.Sprintf("move the transaction-starting call out of %q into an allowed layer: %v", layer, allowed),
				))
				return true
			})
		}
	})
	return violations
}

func (r *Boundary) checkSignatureTypes(ctx *core.Context, allowed []string) []core.Violation {
	if len(r.cfg.Types) == 0 {
		return nil
	}
	wanted := stringSet(r.cfg.Types)
	allowedSet := stringSet(allowed)

	var violations []core.Violation
	r.walkInternalPackages(ctx, func(pkg *packages.Package, layer string) {
		if allowedSet[layer] || pkg.TypesInfo == nil {
			return
		}
		for _, file := range pkg.Syntax {
			for _, decl := range file.Decls {
				fd, ok := decl.(*ast.FuncDecl)
				if !ok || fd.Type == nil {
					continue
				}
				r.checkFields(ctx, pkg, fd.Type.Params, wanted, allowed, &violations)
				r.checkFields(ctx, pkg, fd.Type.Results, wanted, allowed, &violations)
			}
		}
	})
	return violations
}

func (r *Boundary) checkFields(ctx *core.Context, pkg *packages.Package, fields *ast.FieldList, wanted map[string]bool, allowed []string, out *[]core.Violation) {
	if fields == nil {
		return
	}
	for _, field := range fields.List {
		typ := pkg.TypesInfo.TypeOf(field.Type)
		id := analysisutil.NamedQualifiedName(analysisutil.StripWrappers(typ))
		if id == "" || !wanted[id] {
			continue
		}
		pos := pkg.Fset.Position(field.Pos())
		*out = append(*out, r.violation(
			typeInSignatureID,
			analysisutil.RelPathFromRoot(ctx.Root(), pos.Filename),
			pos.Line,
			fmt.Sprintf("tx type %q must not appear in function signature outside allowed layers: %v", id, allowed),
			fmt.Sprintf("keep %q confined to allowed layers: %v", id, allowed),
		))
	}
}

func (r *Boundary) walkInternalPackages(ctx *core.Context, visit func(*packages.Package, string)) {
	module := analysisutil.ResolveModuleFromContext(ctx, "")
	if module == "" {
		return
	}
	internalPrefix := module + "/internal/"
	for _, pkg := range ctx.Pkgs() {
		if !strings.HasPrefix(pkg.PkgPath, internalPrefix) {
			continue
		}
		if ctx.IsExcluded(analysisutil.ProjectRelativePackagePath(pkg.PkgPath, module)) {
			continue
		}
		layer := packageLayer(ctx.Arch(), pkg.PkgPath, internalPrefix)
		if layer == "" {
			continue
		}
		visit(pkg, layer)
	}
}

func (r *Boundary) violation(rule, file string, line int, message, fix string) core.Violation {
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

func packageLayer(arch core.Architecture, pkgPath, internalPrefix string) string {
	layers := arch.Layers.Sublayers
	if arch.Layout.DomainDir == "" {
		return flatLayer(layers, strings.TrimPrefix(pkgPath, internalPrefix))
	}
	domainPrefix := internalPrefix + arch.Layout.DomainDir + "/"
	if !strings.HasPrefix(pkgPath, domainPrefix) {
		return ""
	}
	afterDomainDir := strings.TrimPrefix(pkgPath, domainPrefix)
	parts := strings.SplitN(afterDomainDir, "/", 2)
	if len(parts) < 2 {
		return ""
	}
	return nestedLayer(layers, parts[1])
}

func flatLayer(layers []string, rel string) string {
	parts := strings.SplitN(rel, "/", 2)
	if len(parts) == 0 {
		return ""
	}
	if slices.Contains(layers, parts[0]) {
		return parts[0]
	}
	return ""
}

func nestedLayer(layers []string, rel string) string {
	parts := strings.SplitN(rel, "/", 3)
	if len(parts) >= 2 {
		nested := parts[0] + "/" + parts[1]
		if slices.Contains(layers, nested) {
			return nested
		}
		if hasNestedLayer(layers, parts[0]) {
			return ""
		}
	}
	if len(parts) > 0 && slices.Contains(layers, parts[0]) {
		return parts[0]
	}
	return ""
}

func hasNestedLayer(layers []string, root string) bool {
	prefix := root + "/"
	for _, layer := range layers {
		if strings.HasPrefix(layer, prefix) {
			return true
		}
	}
	return false
}

func stringSet(values []string) map[string]bool {
	out := make(map[string]bool, len(values))
	for _, value := range values {
		out[value] = true
	}
	return out
}
