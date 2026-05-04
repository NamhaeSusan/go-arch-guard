package orchestration

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"slices"
	"strconv"
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

type returnContext struct {
	info        *types.Info
	resultTypes []types.Type
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
			metrics := r.measure(fd, pkg.TypesInfo)
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

func (r *LogicBudget) measure(fd *ast.FuncDecl, info *types.Info) functionMetrics {
	metrics := functionMetrics{cyclomatic: 1}
	r.measureBlock(fd.Body, &metrics, newReturnContextFromFuncType(fd.Type, info))
	return metrics
}

func (r *LogicBudget) measureBlock(block *ast.BlockStmt, metrics *functionMetrics, retCtx returnContext) {
	if block == nil {
		return
	}
	for _, stmt := range block.List {
		r.measureStmt(stmt, metrics, retCtx)
	}
}

func (r *LogicBudget) measureStmt(stmt ast.Stmt, metrics *functionMetrics, retCtx returnContext) {
	if stmt == nil {
		return
	}
	if !r.cfg.countErrorBranches {
		if ifStmt, ok := stmt.(*ast.IfStmt); ok && isSimpleErrorReturn(ifStmt, retCtx) {
			r.measureStmt(ifStmt.Init, metrics, retCtx)
			return
		}
	}
	switch s := stmt.(type) {
	case *ast.BlockStmt:
		r.measureBlock(s, metrics, retCtx)
	case *ast.IfStmt:
		metrics.statements++
		metrics.branches++
		metrics.cyclomatic++
		metrics.cyclomatic += booleanComplexity(s.Cond)
		r.measureExpr(s.Cond, metrics, retCtx)
		r.measureStmt(s.Init, metrics, retCtx)
		r.measureBlock(s.Body, metrics, retCtx)
		r.measureStmt(s.Else, metrics, retCtx)
	case *ast.ForStmt:
		metrics.statements++
		metrics.branches++
		metrics.cyclomatic++
		metrics.cyclomatic += booleanComplexity(s.Cond)
		r.measureExpr(s.Cond, metrics, retCtx)
		r.measureStmt(s.Init, metrics, retCtx)
		r.measureStmt(s.Post, metrics, retCtx)
		r.measureBlock(s.Body, metrics, retCtx)
	case *ast.RangeStmt:
		metrics.statements++
		metrics.branches++
		metrics.cyclomatic++
		r.measureExpr(s.X, metrics, retCtx)
		r.measureBlock(s.Body, metrics, retCtx)
	case *ast.SwitchStmt:
		metrics.statements++
		metrics.branches++
		metrics.cyclomatic++
		r.measureExpr(s.Tag, metrics, retCtx)
		r.measureStmt(s.Init, metrics, retCtx)
		for _, stmt := range s.Body.List {
			cc, ok := stmt.(*ast.CaseClause)
			if !ok {
				continue
			}
			if len(cc.List) > 0 {
				metrics.branches++
				metrics.cyclomatic++
			}
			for _, expr := range cc.List {
				r.measureExpr(expr, metrics, retCtx)
			}
			for _, child := range cc.Body {
				r.measureStmt(child, metrics, retCtx)
			}
		}
	case *ast.TypeSwitchStmt:
		metrics.statements++
		metrics.branches++
		metrics.cyclomatic++
		r.measureStmt(s.Init, metrics, retCtx)
		r.measureFuncLits(s.Assign, metrics, retCtx)
		for _, stmt := range s.Body.List {
			cc, ok := stmt.(*ast.CaseClause)
			if !ok {
				continue
			}
			if len(cc.List) > 0 {
				metrics.branches++
				metrics.cyclomatic++
			}
			for _, expr := range cc.List {
				r.measureExpr(expr, metrics, retCtx)
			}
			for _, child := range cc.Body {
				r.measureStmt(child, metrics, retCtx)
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
				r.measureFuncLits(cc.Comm, metrics, retCtx)
			}
			for _, child := range cc.Body {
				r.measureStmt(child, metrics, retCtx)
			}
		}
	default:
		metrics.statements++
		r.measureFuncLits(stmt, metrics, retCtx)
	}
}

func (r *LogicBudget) measureExpr(expr ast.Expr, metrics *functionMetrics, retCtx returnContext) {
	r.measureFuncLits(expr, metrics, retCtx)
}

func (r *LogicBudget) measureFuncLits(node ast.Node, metrics *functionMetrics, retCtx returnContext) {
	if node == nil {
		return
	}
	ast.Inspect(node, func(n ast.Node) bool {
		lit, ok := n.(*ast.FuncLit)
		if !ok {
			return true
		}
		r.measureBlock(lit.Body, metrics, newReturnContextFromFuncType(lit.Type, retCtx.info))
		return false
	})
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

func newReturnContextFromFuncType(funcType *ast.FuncType, info *types.Info) returnContext {
	ctx := returnContext{info: info}
	if funcType == nil || funcType.Results == nil || info == nil {
		return ctx
	}
	for _, field := range funcType.Results.List {
		t := info.TypeOf(field.Type)
		count := len(field.Names)
		if count == 0 {
			count = 1
		}
		for range count {
			ctx.resultTypes = append(ctx.resultTypes, t)
		}
	}
	return ctx
}

func isSimpleErrorReturn(stmt *ast.IfStmt, retCtx returnContext) bool {
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
	return isErrorPropagationReturn(ret, errName, retCtx)
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

func isErrorPropagationReturn(ret *ast.ReturnStmt, errName string, retCtx returnContext) bool {
	if retCtx.info == nil || len(retCtx.resultTypes) == 0 || len(ret.Results) != len(retCtx.resultTypes) {
		return false
	}
	var propagated bool
	for i, expr := range ret.Results {
		if isErrorType(retCtx.resultTypes[i]) {
			if !isErrorPropagationExpr(expr, errName, retCtx.info) {
				return false
			}
			propagated = true
			continue
		}
		if !isZeroReturnExpr(expr) {
			return false
		}
	}
	return propagated
}

func isErrorType(t types.Type) bool {
	if t == nil {
		return false
	}
	errorObj := types.Universe.Lookup("error")
	if errorObj == nil {
		return false
	}
	iface, ok := errorObj.Type().Underlying().(*types.Interface)
	return ok && types.Implements(types.Unalias(t), iface)
}

func isErrorPropagationExpr(expr ast.Expr, errName string, info *types.Info) bool {
	if ident, ok := expr.(*ast.Ident); ok {
		return ident.Name == errName
	}
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return false
	}
	switch qualifiedCallName(call, info) {
	case "fmt.Errorf":
		return fmtErrorfWraps(call, errName)
	case "errors.Join":
		return errorsJoinPropagatesOnlyCheckedErr(call, errName)
	default:
		return false
	}
}

func qualifiedCallName(call *ast.CallExpr, info *types.Info) string {
	if call == nil || info == nil {
		return ""
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return ""
	}
	fn, ok := info.Uses[sel.Sel].(*types.Func)
	if !ok || fn.Pkg() == nil {
		return ""
	}
	return fn.Pkg().Path() + "." + fn.Name()
}

func fmtErrorfWraps(call *ast.CallExpr, errName string) bool {
	if len(call.Args) == 0 {
		return false
	}
	format, ok := call.Args[0].(*ast.BasicLit)
	if !ok || format.Kind != token.STRING {
		return false
	}
	formatValue, err := strconv.Unquote(format.Value)
	if err != nil {
		return false
	}
	for _, idx := range fmtWrapArgIndexes(formatValue) {
		argPos := idx + 1
		if argPos >= len(call.Args) {
			continue
		}
		if ident, ok := call.Args[argPos].(*ast.Ident); ok && ident.Name == errName {
			return true
		}
	}
	return false
}

func fmtWrapArgIndexes(format string) []int {
	var indexes []int
	nextArg := 0
	for i := 0; i < len(format); i++ {
		if format[i] != '%' {
			continue
		}
		i++
		if i >= len(format) {
			break
		}
		if format[i] == '%' {
			continue
		}

		verbArg := -1
		if idx, next, ok := parseFmtIndex(format, i); ok {
			verbArg = idx
			i = next
		}
		for i < len(format) && strings.ContainsRune("#0+- ", rune(format[i])) {
			i++
		}
		i, nextArg = consumeFmtWidthOrPrecisionArg(format, i, nextArg)
		if i < len(format) && format[i] == '.' {
			i, nextArg = consumeFmtWidthOrPrecisionArg(format, i+1, nextArg)
		}
		if idx, next, ok := parseFmtIndex(format, i); ok {
			verbArg = idx
			i = next
		}
		if i >= len(format) {
			break
		}

		arg := verbArg
		if arg < 0 {
			arg = nextArg
		}
		if format[i] == 'w' {
			indexes = append(indexes, arg)
		}
		if verbConsumesArg(format[i]) {
			nextArg = arg + 1
		}
	}
	return indexes
}

func consumeFmtWidthOrPrecisionArg(format string, i, nextArg int) (int, int) {
	if idx, next, ok := parseFmtIndex(format, i); ok {
		if next < len(format) && format[next] == '*' {
			return next + 1, idx + 1
		}
		return i, nextArg
	}
	if i < len(format) && format[i] == '*' {
		return i + 1, nextArg + 1
	}
	for i < len(format) && format[i] >= '0' && format[i] <= '9' {
		i++
	}
	return i, nextArg
}

func parseFmtIndex(format string, i int) (int, int, bool) {
	if i >= len(format) || format[i] != '[' {
		return 0, i, false
	}
	j := i + 1
	for j < len(format) && format[j] >= '0' && format[j] <= '9' {
		j++
	}
	if j == i+1 || j >= len(format) || format[j] != ']' {
		return 0, i, false
	}
	n, err := strconv.Atoi(format[i+1 : j])
	if err != nil || n <= 0 {
		return 0, i, false
	}
	return n - 1, j + 1, true
}

func verbConsumesArg(verb byte) bool {
	return verb != '%'
}

func errorsJoinPropagatesOnlyCheckedErr(call *ast.CallExpr, errName string) bool {
	var hasErr bool
	for _, arg := range call.Args {
		ident, ok := arg.(*ast.Ident)
		if !ok {
			return false
		}
		switch ident.Name {
		case errName:
			hasErr = true
		case "nil":
		default:
			return false
		}
	}
	return hasErr
}

func isZeroReturnExpr(expr ast.Expr) bool {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name == "nil" || e.Name == "false"
	case *ast.BasicLit:
		return e.Value == "0" || e.Value == `""`
	case *ast.CompositeLit:
		return len(e.Elts) == 0
	default:
		return false
	}
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
