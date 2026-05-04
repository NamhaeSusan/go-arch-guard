package dependency

import (
	"fmt"
	"go/ast"
	"slices"
	"sort"
	"strings"

	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/core/analysisutil"
	"golang.org/x/tools/go/packages"
)

type NoSideEffectCallInCore struct {
	cfg ruleConfig
}

func NewNoSideEffectCallInCore(opts ...Option) *NoSideEffectCallInCore {
	return &NoSideEffectCallInCore{cfg: newConfig(opts, core.Warning)}
}

func (r *NoSideEffectCallInCore) Spec() core.RuleSpec {
	return core.RuleSpec{
		ID:              "purity.no-side-effect-call-in-core",
		Description:     "domain core layers should not call side-effectful runtime APIs directly",
		DefaultSeverity: r.cfg.severity,
		Violations: []core.ViolationSpec{
			{
				ID:              "purity.no-side-effect-call-in-core",
				Description:     "domain core calls a side-effectful runtime API directly",
				DefaultSeverity: r.cfg.severity,
			},
		},
	}
}

func (r *NoSideEffectCallInCore) Check(ctx *core.Context) []core.Violation {
	if ctx == nil {
		return nil
	}
	arch := ctx.Arch()
	module := analysisutil.ResolveModuleFromContext(ctx, "")
	if warns := validateSideEffectModule(ctx.Pkgs(), module); len(warns) > 0 {
		return warns
	}
	layers := r.inspectedLayers(arch)
	if module == "" || len(layers) == 0 || len(r.cfg.deniedCalls) == 0 {
		return nil
	}

	var violations []core.Violation
	for _, pkg := range ctx.Pkgs() {
		if !isInspectedLayerPackage(pkg, module, arch.Layout, layers) {
			continue
		}
		violations = append(violations, r.checkPackage(ctx, pkg)...)
	}
	return violations
}

func (r *NoSideEffectCallInCore) checkPackage(ctx *core.Context, pkg *packages.Package) []core.Violation {
	var violations []core.Violation
	for _, file := range pkg.Syntax {
		filename := pkg.Fset.Position(file.Pos()).Filename
		relPath := analysisutil.RelativePathForPackage(pkg, filename)
		if analysisutil.IsTestFile(file, pkg.Fset) || isGeneratedFile(file) || ctx.IsExcluded(relPath) {
			continue
		}
		ast.Inspect(file, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			callID := analysisutil.ResolveCalleeID(pkg.TypesInfo, call)
			if callID == "" || !r.isDenied(callID) || r.isAllowed(callID) {
				return true
			}
			pos := pkg.Fset.Position(call.Pos())
			violations = append(violations, core.Violation{
				File:              analysisutil.RelativePathForPackage(pkg, pos.Filename),
				Line:              pos.Line,
				Rule:              "purity.no-side-effect-call-in-core",
				Message:           fmt.Sprintf("domain core calls side-effectful API %q directly", callID),
				Fix:               "pass runtime values in from an outer layer or configure dependency.WithAllowedCalls for intentional exceptions",
				DefaultSeverity:   r.cfg.severity,
				EffectiveSeverity: r.cfg.severity,
			})
			return true
		})
	}
	return violations
}

func (r *NoSideEffectCallInCore) inspectedLayers(arch core.Architecture) []string {
	if len(r.cfg.inspectedLayers) > 0 {
		return r.cfg.inspectedLayers
	}
	if len(arch.Layers.PkgRestricted) > 0 {
		layers := make([]string, 0, len(arch.Layers.PkgRestricted))
		for layer, restricted := range arch.Layers.PkgRestricted {
			if restricted {
				layers = append(layers, layer)
			}
		}
		sort.Strings(layers)
		return layers
	}
	return []string{"core/model", "entity", "domain", "event"}
}

func (r *NoSideEffectCallInCore) isDenied(callID string) bool {
	return matchesAnyCallPattern(r.cfg.deniedCalls, callID)
}

func (r *NoSideEffectCallInCore) isAllowed(callID string) bool {
	return matchesAnyCallPattern(r.cfg.allowedCalls, callID)
}

func matchesAnyCallPattern(patterns []string, callID string) bool {
	return slices.ContainsFunc(patterns, func(pattern string) bool {
		return matchCallPattern(pattern, callID)
	})
}

func matchCallPattern(pattern, callID string) bool {
	if strings.HasSuffix(pattern, "*") {
		return strings.HasPrefix(callID, strings.TrimSuffix(pattern, "*"))
	}
	return pattern == callID
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
	for i := 0; i < len(parts); i++ {
		if parts[i] != internalRoot {
			continue
		}
		afterInternal := parts[i+1:]
		for _, layer := range layers {
			if matchesLayer(afterInternal, layout.DomainDir, layer) {
				return true
			}
		}
	}
	return false
}

func validateSideEffectModule(pkgs []*packages.Package, module string) []core.Violation {
	if module == "" {
		return []core.Violation{metaNoMatchingPackagesForSideEffect("project module could not be determined - purity checks will be skipped")}
	}
	prefix := module + "/"
	for _, pkg := range pkgs {
		if pkg != nil && (pkg.PkgPath == module || strings.HasPrefix(pkg.PkgPath, prefix)) {
			return nil
		}
	}
	return []core.Violation{metaNoMatchingPackagesForSideEffect(fmt.Sprintf("module %q does not match any loaded package - purity checks will be skipped", module))}
}

func metaNoMatchingPackagesForSideEffect(message string) core.Violation {
	return core.Violation{
		Rule:              "meta.no-matching-packages",
		Message:           message,
		Fix:               "verify the module argument matches go.mod",
		DefaultSeverity:   core.Warning,
		EffectiveSeverity: core.Warning,
	}
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

func isGeneratedFile(file *ast.File) bool {
	return ast.IsGenerated(file)
}

func defaultDeniedCalls() []string {
	return []string{
		"time.Now",
		"time.Since",
		"time.Until",
		"time.After",
		"os.Getenv",
		"os.LookupEnv",
		"os.ReadFile",
		"os.WriteFile",
		"os.Open",
		"os.OpenFile",
		"os.Create",
		"os.CreateTemp",
		"os.Mkdir",
		"os.MkdirAll",
		"os.MkdirTemp",
		"os.Remove",
		"os.RemoveAll",
		"os.Rename",
		"os.Truncate",
		"os.Chmod",
		"os.Chown",
		"os.Chtimes",
		"os.Link",
		"os.Symlink",
		"log.Print",
		"log.Printf",
		"log.Println",
		"log.Fatal",
		"log.Fatalf",
		"log.Fatalln",
		"log.Panic",
		"log.Panicf",
		"log.Panicln",
		"math/rand.*",
		"crypto/rand.Read",
		"net/http.Get",
		"net/http.Head",
		"net/http.Post",
		"net/http.PostForm",
		"net/http.(*Client).*",
	}
}

var _ core.Rule = (*NoSideEffectCallInCore)(nil)
