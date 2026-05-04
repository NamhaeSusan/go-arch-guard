package types

import (
	"fmt"
	"go/ast"
	gotypes "go/types"
	"slices"
	"strings"

	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/core/analysisutil"
	"golang.org/x/tools/go/packages"
)

const ruleNoPanicInDomain = "errors.no-panic-in-domain"

type NoPanicInDomain struct {
	cfg ruleConfig
}

func NewNoPanicInDomain(opts ...Option) *NoPanicInDomain {
	return &NoPanicInDomain{cfg: newConfig(opts, core.Warning)}
}

func (r *NoPanicInDomain) Spec() core.RuleSpec {
	return core.RuleSpec{
		ID:              ruleNoPanicInDomain,
		Description:     "domain and application layers should not panic or terminate the process",
		DefaultSeverity: r.cfg.severity,
		Violations: []core.ViolationSpec{
			{
				ID:              ruleNoPanicInDomain,
				Description:     "domain/application code panics or exits instead of returning an error",
				DefaultSeverity: r.cfg.severity,
			},
		},
	}
}

func (r *NoPanicInDomain) Check(ctx *core.Context) []core.Violation {
	if ctx == nil {
		return nil
	}
	module := analysisutil.ResolveModuleFromContext(ctx, "")
	if warns := validateModule(ctx.Pkgs(), module); len(warns) > 0 {
		return warns
	}
	layers := r.inspectedLayers(ctx.Arch())
	if len(layers) == 0 {
		return nil
	}

	var violations []core.Violation
	for _, pkg := range ctx.Pkgs() {
		if !isInspectedLayerPackage(pkg, module, ctx.Arch().Layout, layers) {
			continue
		}
		violations = append(violations, r.checkPackage(ctx, pkg)...)
	}
	return violations
}

func (r *NoPanicInDomain) checkPackage(ctx *core.Context, pkg *packages.Package) []core.Violation {
	var violations []core.Violation
	for _, file := range pkg.Syntax {
		filename := pkg.Fset.Position(file.Pos()).Filename
		relPath := analysisutil.RelativePathForPackage(pkg, filename)
		if analysisutil.IsTestFile(file, pkg.Fset) || ast.IsGenerated(file) ||
			ctx.IsExcluded(relPath) || r.isAllowedPath(relPath) {
			continue
		}
		violations = append(violations, r.checkFile(pkg, file, relPath)...)
	}
	return violations
}

func (r *NoPanicInDomain) checkFile(pkg *packages.Package, file *ast.File, relPath string) []core.Violation {
	var violations []core.Violation
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if d.Body == nil || r.isAllowedFunction(d.Name.Name) {
				continue
			}
			violations = append(violations, r.checkNode(pkg, d.Body, relPath)...)
		case *ast.GenDecl:
			violations = append(violations, r.checkNode(pkg, d, relPath)...)
		}
	}
	return violations
}

func (r *NoPanicInDomain) checkNode(pkg *packages.Package, node ast.Node, relPath string) []core.Violation {
	if node == nil {
		return nil
	}
	var violations []core.Violation
	ast.Inspect(node, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		callID := failureCallID(pkg.TypesInfo, call)
		if !isDeniedFailureCall(callID) {
			return true
		}
		pos := pkg.Fset.Position(call.Pos())
		violations = append(violations, core.Violation{
			File:              relPath,
			Line:              pos.Line,
			Rule:              ruleNoPanicInDomain,
			Message:           fmt.Sprintf("domain/application layer calls %q instead of returning an error", callID),
			Fix:               "return an error and let outer layers decide whether to retry, recover, roll back, or exit",
			DefaultSeverity:   r.cfg.severity,
			EffectiveSeverity: r.cfg.severity,
		})
		return true
	})
	return violations
}

func failureCallID(info *gotypes.Info, call *ast.CallExpr) string {
	if ident, ok := call.Fun.(*ast.Ident); ok && ident.Name == "panic" {
		if info == nil {
			return "panic"
		}
		obj := info.Uses[ident]
		if obj == nil {
			return "panic"
		}
		if _, ok := obj.(*gotypes.Builtin); ok {
			return "panic"
		}
	}
	return analysisutil.ResolveCalleeID(info, call)
}

func isDeniedFailureCall(callID string) bool {
	switch callID {
	case "panic", "log.Fatal", "log.Fatalf", "log.Fatalln", "os.Exit":
		return true
	default:
		return false
	}
}

func (r *NoPanicInDomain) inspectedLayers(arch core.Architecture) []string {
	if len(r.cfg.inspectedLayers) > 0 {
		return r.cfg.inspectedLayers
	}
	candidates := []string{"app", "core/model", "core/svc", "event", "entity", "usecase", "domain", "application"}
	layers := make([]string, 0, len(candidates)+1)
	for _, candidate := range candidates {
		if hasSublayer(arch, candidate) {
			layers = append(layers, candidate)
		}
	}
	if hasSublayer(arch, "core") && !hasSublayerPrefix(arch, "core/") {
		layers = append(layers, "core")
	}
	if len(layers) > 0 {
		return layers
	}
	return []string{"app", "core/model", "core/svc", "event", "entity", "usecase", "domain", "application", "core"}
}

func hasSublayer(arch core.Architecture, layer string) bool {
	return slices.Contains(arch.Layers.Sublayers, layer)
}

func hasSublayerPrefix(arch core.Architecture, prefix string) bool {
	return slices.ContainsFunc(arch.Layers.Sublayers, func(layer string) bool {
		return strings.HasPrefix(layer, prefix)
	})
}

func (r *NoPanicInDomain) isAllowedPath(path string) bool {
	return matchesAnyPath(r.cfg.allowedPaths, path)
}

func (r *NoPanicInDomain) isAllowedFunction(name string) bool {
	return matchesAnyPattern(r.cfg.allowedFunctions, name)
}

func matchesAnyPath(patterns []string, path string) bool {
	path = core.NormalizeMatchPath(path)
	for _, pattern := range patterns {
		pattern = core.NormalizeMatchPath(pattern)
		if strings.HasSuffix(pattern, "/...") {
			prefix := strings.TrimSuffix(pattern, "/...")
			if path == prefix || strings.HasPrefix(path, prefix+"/") {
				return true
			}
			continue
		}
		if path == pattern {
			return true
		}
	}
	return false
}

func matchesAnyPattern(patterns []string, value string) bool {
	return slices.ContainsFunc(patterns, func(pattern string) bool {
		if strings.HasSuffix(pattern, "*") {
			return strings.HasPrefix(value, strings.TrimSuffix(pattern, "*"))
		}
		return pattern == value
	})
}

func isInspectedLayerPackage(pkg *packages.Package, module string, layout core.LayoutModel, layers []string) bool {
	if pkg == nil || module == "" {
		return false
	}
	rel := analysisutil.ProjectRelativePackagePath(pkg.PkgPath, module)
	if rel == "" || rel == "." {
		return false
	}
	parts := strings.Split(rel, "/")
	internalRoot := layout.InternalRoot
	if internalRoot == "" {
		internalRoot = "internal"
	}
	rootParts := strings.Split(strings.Trim(internalRoot, "/"), "/")
	if !hasPathPrefix(parts, rootParts) {
		return false
	}
	afterInternal := parts[len(rootParts):]
	for _, layer := range layers {
		if matchesLayer(afterInternal, layout.DomainDir, layer) {
			return true
		}
	}
	return false
}

func hasPathPrefix(parts, prefix []string) bool {
	if len(parts) < len(prefix) {
		return false
	}
	for i := range prefix {
		if parts[i] != prefix[i] {
			return false
		}
	}
	return true
}

func matchesLayer(afterInternal []string, domainDir, layer string) bool {
	if domainDir != "" {
		if len(afterInternal) < 2 || afterInternal[0] != domainDir {
			return false
		}
		return matchesPathPrefix(afterInternal[2:], strings.Split(layer, "/"))
	}
	return matchesPathPrefix(afterInternal, strings.Split(layer, "/"))
}

func matchesPathPrefix(parts, want []string) bool {
	if len(parts) < len(want) {
		return false
	}
	for i := range want {
		if parts[i] != want[i] {
			return false
		}
	}
	return true
}

func validateModule(pkgs []*packages.Package, module string) []core.Violation {
	if module == "" {
		return []core.Violation{metaNoMatchingPackages("project module could not be determined - panic checks will be skipped")}
	}
	prefix := module + "/"
	for _, pkg := range pkgs {
		if pkg != nil && (pkg.PkgPath == module || strings.HasPrefix(pkg.PkgPath, prefix)) {
			return nil
		}
	}
	return []core.Violation{metaNoMatchingPackages(fmt.Sprintf("module %q does not match any loaded package - panic checks will be skipped", module))}
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

var _ core.Rule = (*NoPanicInDomain)(nil)
