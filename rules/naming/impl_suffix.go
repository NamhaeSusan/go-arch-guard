package naming

import (
	"go/ast"
	"go/token"
	"strings"

	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/core/analysisutil"
)

type ImplSuffix struct {
	severity core.Severity
}

func NewImplSuffix(opts ...Option) *ImplSuffix {
	cfg := newConfig(opts, core.Warning)
	return &ImplSuffix{severity: cfg.severity}
}

func (r *ImplSuffix) Spec() core.RuleSpec {
	return core.RuleSpec{
		ID:              "naming.no-impl-suffix",
		Description:     `exported type names must not end with "Impl"`,
		DefaultSeverity: r.severity,
	}
}

func (r *ImplSuffix) Check(ctx *core.Context) []core.Violation {
	var violations []core.Violation
	for _, pkg := range ctx.Pkgs() {
		for _, file := range pkg.Syntax {
			filePath := analysisutil.RelativePathForPackage(pkg, pkg.Fset.Position(file.Pos()).Filename)
			if ctx.IsExcluded(filePath) {
				continue
			}
			analysisutil.WalkTypeSpecs(file, pkg.Fset, func(ts *ast.TypeSpec, pos token.Position) {
				if !ts.Name.IsExported() || !strings.HasSuffix(ts.Name.Name, "Impl") {
					return
				}
				violations = append(violations, core.Violation{
					File:              analysisutil.RelativePathForPackage(pkg, pos.Filename),
					Line:              pos.Line,
					Rule:              "naming.no-impl-suffix",
					Message:           `type "` + ts.Name.Name + `" uses banned suffix "Impl"`,
					Fix:               "rename without Impl suffix",
					DefaultSeverity:   r.severity,
					EffectiveSeverity: r.severity,
				})
			})
		}
	}
	return violations
}

var _ core.Rule = (*ImplSuffix)(nil)
