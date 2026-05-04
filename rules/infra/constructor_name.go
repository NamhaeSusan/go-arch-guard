package infra

import (
	"fmt"
	"go/ast"
	"go/types"
	"slices"
	"strings"

	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/core/analysisutil"
	"golang.org/x/tools/go/packages"
)

type ConstructorName struct {
	cfg ruleConfig
}

func NewConstructorName(opts ...Option) *ConstructorName {
	return &ConstructorName{cfg: newConfig(opts, core.Warning)}
}

func (r *ConstructorName) Spec() core.RuleSpec {
	return core.RuleSpec{
		ID:              "infra.constructor-name",
		Description:     "infra adapter constructors should use configured constructor names",
		DefaultSeverity: r.cfg.severity,
	}
}

func (r *ConstructorName) Check(ctx *core.Context) []core.Violation {
	if ctx == nil {
		return nil
	}
	arch := ctx.Arch()
	module := analysisutil.ResolveModuleFromContext(ctx, "")
	infraSublayers := r.infraSublayers(arch)

	var violations []core.Violation
	for _, pkg := range ctx.Pkgs() {
		if !isInfraSublayerPackage(pkg, module, arch.Layout, infraSublayers) {
			continue
		}
		violations = append(violations, r.checkPackage(ctx, pkg)...)
	}
	return violations
}

func (r *ConstructorName) checkPackage(ctx *core.Context, pkg *packages.Package) []core.Violation {
	var violations []core.Violation
	for _, file := range pkg.Syntax {
		if analysisutil.IsTestFile(file, pkg.Fset) {
			continue
		}
		filePath := analysisutil.RelativePathForPackage(pkg, pkg.Fset.Position(file.Pos()).Filename)
		if ctx.IsExcluded(filePath) {
			continue
		}
		analysisutil.WalkFuncDecls(file, func(fd *ast.FuncDecl) {
			if !looksLikeConstructor(fd) || !returnsSamePackageNamedType(pkg, fd) || r.isAllowed(fd.Name.Name) {
				return
			}
			pos := pkg.Fset.Position(fd.Name.Pos())
			violations = append(violations, core.Violation{
				File:              analysisutil.RelativePathForPackage(pkg, pos.Filename),
				Line:              pos.Line,
				Rule:              "infra.constructor-name",
				Message:           fmt.Sprintf("infra constructor %q must use one of: %s", fd.Name.Name, strings.Join(r.cfg.allowedConstructorNames, ", ")),
				Fix:               "rename the constructor or configure infra.WithAllowedConstructorNames",
				DefaultSeverity:   r.cfg.severity,
				EffectiveSeverity: r.cfg.severity,
			})
		})
	}
	return violations
}

func looksLikeConstructor(fd *ast.FuncDecl) bool {
	return fd != nil &&
		fd.Recv == nil &&
		fd.Name != nil &&
		fd.Name.IsExported() &&
		strings.HasPrefix(fd.Name.Name, "New")
}

func returnsSamePackageNamedType(pkg *packages.Package, fd *ast.FuncDecl) bool {
	if pkg == nil || pkg.TypesInfo == nil || fd == nil || fd.Type == nil || fd.Type.Results == nil {
		return false
	}
	for _, result := range fd.Type.Results.List {
		if namedTypePackagePath(pkg.TypesInfo.TypeOf(result.Type)) == pkg.PkgPath {
			return true
		}
	}
	return false
}

func namedTypePackagePath(t types.Type) string {
	for {
		if ptr, ok := types.Unalias(t).(*types.Pointer); ok {
			t = ptr.Elem()
			continue
		}
		break
	}
	named, ok := types.Unalias(t).(*types.Named)
	if !ok || named.Obj() == nil || named.Obj().Pkg() == nil {
		return ""
	}
	return named.Obj().Pkg().Path()
}

func (r *ConstructorName) isAllowed(name string) bool {
	return slices.Contains(r.cfg.allowedConstructorNames, name)
}

func (r *ConstructorName) infraSublayers(arch core.Architecture) []string {
	if len(r.cfg.infraSublayers) > 0 {
		return r.cfg.infraSublayers
	}
	var out []string
	for _, sl := range arch.Layers.Sublayers {
		if isDefaultInfraSublayer(sl) {
			out = append(out, sl)
		}
	}
	if len(out) == 0 {
		return []string{"infra"}
	}
	return out
}

func isDefaultInfraSublayer(sublayer string) bool {
	base := sublayer
	if i := strings.LastIndex(sublayer, "/"); i >= 0 {
		base = sublayer[i+1:]
	}
	return base == "infra" || base == "adapter" || base == "repository"
}

func isInfraSublayerPackage(pkg *packages.Package, module string, layout core.LayoutModel, infraSublayers []string) bool {
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
		for _, sublayer := range infraSublayers {
			if matchesInfraSublayer(afterInternal, layout.DomainDir, sublayer) {
				return true
			}
		}
	}
	return false
}

func matchesInfraSublayer(afterInternal []string, domainDir, sublayer string) bool {
	if domainDir != "" {
		if len(afterInternal) < 2 || afterInternal[0] != domainDir {
			return false
		}
		return matchesPathPrefix(afterInternal[2:], strings.Split(sublayer, "/"))
	}
	return matchesPathPrefix(afterInternal, strings.Split(sublayer, "/"))
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

var _ core.Rule = (*ConstructorName)(nil)
