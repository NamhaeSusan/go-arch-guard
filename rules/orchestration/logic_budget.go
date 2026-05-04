package orchestration

import (
	"fmt"
	"go/ast"
	"go/token"
	"slices"
	"strings"

	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/core/analysisutil"
	"golang.org/x/tools/go/packages"
)

type LogicBudget struct {
	cfg ruleConfig
}

type functionMetrics struct {
	branches   int
	statements int
	cyclomatic int
}

func NewLogicBudget(opts ...Option) *LogicBudget {
	return &LogicBudget{cfg: newConfig(opts, core.Warning)}
}

func (r *LogicBudget) Spec() core.RuleSpec {
	return core.RuleSpec{
		ID:              "orchestration.logic-budget",
		Description:     "orchestration functions should stay within configurable control-flow budgets",
		DefaultSeverity: r.cfg.severity,
	}
}

func (r *LogicBudget) Check(ctx *core.Context) []core.Violation {
	if ctx == nil {
		return nil
	}
	arch := ctx.Arch()
	module := analysisutil.ResolveModuleFromContext(ctx, "")
	dirs := r.orchestrationDirs(arch.Layout)
	if len(dirs) == 0 {
		return nil
	}

	var violations []core.Violation
	for _, pkg := range ctx.Pkgs() {
		if !isOrchestrationPackage(pkg, module, dirs) {
			continue
		}
		violations = append(violations, r.checkPackage(ctx, pkg)...)
	}
	return violations
}

func (r *LogicBudget) checkPackage(ctx *core.Context, pkg *packages.Package) []core.Violation {
	var violations []core.Violation
	for _, file := range pkg.Syntax {
		if analysisutil.IsTestFile(file, pkg.Fset) {
			continue
		}
		filePath := analysisutil.RelativePathForPackage(pkg, pkg.Fset.Position(file.Pos()).Filename)
		if ctx.IsExcluded(filePath) || r.isIgnoredPath(filePath) {
			continue
		}
		analysisutil.WalkFuncDecls(file, func(fd *ast.FuncDecl) {
			if fd.Body == nil || r.isIgnored(fd.Name.Name) {
				return
			}
			metrics := r.measure(fd.Body)
			reasons := r.exceededReasons(metrics)
			if len(reasons) == 0 {
				return
			}
			pos := pkg.Fset.Position(fd.Name.Pos())
			violations = append(violations, core.Violation{
				File:              analysisutil.RelativePathForPackage(pkg, pos.Filename),
				Line:              pos.Line,
				Rule:              "orchestration.logic-budget",
				Message:           fmt.Sprintf("orchestration function %q exceeds logic budget: %s", fd.Name.Name, strings.Join(reasons, ", ")),
				Fix:               "move business decisions into domain/app services or raise the configured budget",
				DefaultSeverity:   r.cfg.severity,
				EffectiveSeverity: r.cfg.severity,
			})
		})
	}
	return violations
}

func (r *LogicBudget) measure(body *ast.BlockStmt) functionMetrics {
	metrics := functionMetrics{cyclomatic: 1}
	r.measureBlock(body, &metrics)
	return metrics
}

func (r *LogicBudget) measureBlock(block *ast.BlockStmt, metrics *functionMetrics) {
	if block == nil {
		return
	}
	for _, stmt := range block.List {
		r.measureStmt(stmt, metrics)
	}
}

func (r *LogicBudget) measureStmt(stmt ast.Stmt, metrics *functionMetrics) {
	if stmt == nil {
		return
	}
	if !r.cfg.countErrorBranches {
		if ifStmt, ok := stmt.(*ast.IfStmt); ok && isSimpleErrorReturn(ifStmt) {
			r.measureStmt(ifStmt.Init, metrics)
			return
		}
	}
	switch s := stmt.(type) {
	case *ast.BlockStmt:
		r.measureBlock(s, metrics)
	case *ast.IfStmt:
		metrics.statements++
		metrics.branches++
		metrics.cyclomatic++
		metrics.cyclomatic += booleanComplexity(s.Cond)
		r.measureStmt(s.Init, metrics)
		r.measureBlock(s.Body, metrics)
		r.measureStmt(s.Else, metrics)
	case *ast.ForStmt:
		metrics.statements++
		metrics.branches++
		metrics.cyclomatic++
		metrics.cyclomatic += booleanComplexity(s.Cond)
		r.measureStmt(s.Init, metrics)
		r.measureStmt(s.Post, metrics)
		r.measureBlock(s.Body, metrics)
	case *ast.RangeStmt:
		metrics.statements++
		metrics.branches++
		metrics.cyclomatic++
		r.measureBlock(s.Body, metrics)
	case *ast.SwitchStmt:
		metrics.statements++
		metrics.branches++
		metrics.cyclomatic++
		r.measureStmt(s.Init, metrics)
		for _, stmt := range s.Body.List {
			cc, ok := stmt.(*ast.CaseClause)
			if !ok {
				continue
			}
			if len(cc.List) > 0 {
				metrics.branches++
				metrics.cyclomatic++
			}
			for _, child := range cc.Body {
				r.measureStmt(child, metrics)
			}
		}
	case *ast.TypeSwitchStmt:
		metrics.statements++
		metrics.branches++
		metrics.cyclomatic++
		r.measureStmt(s.Init, metrics)
		for _, stmt := range s.Body.List {
			cc, ok := stmt.(*ast.CaseClause)
			if !ok {
				continue
			}
			if len(cc.List) > 0 {
				metrics.branches++
				metrics.cyclomatic++
			}
			for _, child := range cc.Body {
				r.measureStmt(child, metrics)
			}
		}
	case *ast.SelectStmt:
		metrics.statements++
		metrics.branches++
		metrics.cyclomatic++
		for _, stmt := range s.Body.List {
			cc, ok := stmt.(*ast.CommClause)
			if !ok {
				continue
			}
			if cc.Comm != nil {
				metrics.branches++
				metrics.cyclomatic++
			}
			for _, child := range cc.Body {
				r.measureStmt(child, metrics)
			}
		}
	default:
		metrics.statements++
	}
}

func booleanComplexity(expr ast.Expr) int {
	var count int
	ast.Inspect(expr, func(node ast.Node) bool {
		bin, ok := node.(*ast.BinaryExpr)
		if !ok {
			return true
		}
		if bin.Op == token.LAND || bin.Op == token.LOR {
			count++
		}
		return true
	})
	return count
}

func isSimpleErrorReturn(stmt *ast.IfStmt) bool {
	if stmt == nil || stmt.Else != nil || stmt.Body == nil || len(stmt.Body.List) != 1 {
		return false
	}
	errName, ok := errorNilCheckIdent(stmt.Cond)
	if !ok {
		return false
	}
	ret, ok := stmt.Body.List[0].(*ast.ReturnStmt)
	if !ok {
		return false
	}
	return returnIncludesIdent(ret, errName)
}

func errorNilCheckIdent(expr ast.Expr) (string, bool) {
	bin, ok := expr.(*ast.BinaryExpr)
	if !ok || bin.Op != token.NEQ {
		return "", false
	}
	if ident, ok := bin.X.(*ast.Ident); ok && ident.Name != "" && isNilIdent(bin.Y) {
		return ident.Name, true
	}
	if ident, ok := bin.Y.(*ast.Ident); ok && ident.Name != "" && isNilIdent(bin.X) {
		return ident.Name, true
	}
	return "", false
}

func isNilIdent(expr ast.Expr) bool {
	ident, ok := expr.(*ast.Ident)
	return ok && ident.Name == "nil"
}

func returnIncludesIdent(ret *ast.ReturnStmt, name string) bool {
	for _, expr := range ret.Results {
		if exprIncludesIdent(expr, name) {
			return true
		}
	}
	return false
}

func exprIncludesIdent(expr ast.Expr, name string) bool {
	var found bool
	ast.Inspect(expr, func(node ast.Node) bool {
		if found {
			return false
		}
		ident, ok := node.(*ast.Ident)
		if ok && ident.Name == name {
			found = true
			return false
		}
		return true
	})
	return found
}

func (r *LogicBudget) exceededReasons(metrics functionMetrics) []string {
	var reasons []string
	if metrics.branches > r.cfg.maxBranches {
		reasons = append(reasons, fmt.Sprintf("branches %d > %d", metrics.branches, r.cfg.maxBranches))
	}
	if metrics.statements > r.cfg.maxStatements {
		reasons = append(reasons, fmt.Sprintf("statements %d > %d", metrics.statements, r.cfg.maxStatements))
	}
	if metrics.cyclomatic > r.cfg.maxCyclomatic {
		reasons = append(reasons, fmt.Sprintf("cyclomatic %d > %d", metrics.cyclomatic, r.cfg.maxCyclomatic))
	}
	return reasons
}

func (r *LogicBudget) isIgnored(name string) bool {
	return slices.Contains(r.cfg.ignoredFunctions, name)
}

func (r *LogicBudget) isIgnoredPath(path string) bool {
	path = analysisutil.NormalizeMatchPath(path)
	for _, pattern := range r.cfg.ignoredPaths {
		if matchPathPattern(analysisutil.NormalizeMatchPath(pattern), path) {
			return true
		}
	}
	return false
}

func matchPathPattern(pattern, path string) bool {
	if strings.HasSuffix(pattern, "...") {
		prefix := strings.TrimRight(strings.TrimSuffix(pattern, "..."), "/")
		return path == prefix || strings.HasPrefix(path, prefix+"/")
	}
	return pattern == path
}

func (r *LogicBudget) orchestrationDirs(layout core.LayoutModel) []string {
	if len(r.cfg.orchestrationDirs) > 0 {
		return normalizeOrchestrationDirs(r.cfg.orchestrationDirs, layout.InternalRoot)
	}
	if layout.OrchestrationDir == "" {
		return nil
	}
	return normalizeOrchestrationDirs([]string{layout.OrchestrationDir}, layout.InternalRoot)
}

func normalizeOrchestrationDirs(dirs []string, internalRoot string) []string {
	if internalRoot == "" {
		internalRoot = "internal"
	}
	out := make([]string, 0, len(dirs))
	for _, dir := range dirs {
		dir = analysisutil.NormalizeMatchPath(dir)
		if dir == "" {
			continue
		}
		if strings.Contains(dir, "/") {
			out = append(out, dir)
			continue
		}
		out = append(out, internalRoot+"/"+dir)
	}
	return out
}

func isOrchestrationPackage(pkg *packages.Package, module string, dirs []string) bool {
	if pkg == nil || module == "" {
		return false
	}
	rel := analysisutil.ProjectRelativePackagePath(pkg.PkgPath, module)
	if rel == "" || rel == "." {
		return false
	}
	for _, dir := range dirs {
		if rel == dir || strings.HasPrefix(rel, dir+"/") {
			return true
		}
	}
	return false
}

var _ core.Rule = (*LogicBudget)(nil)
