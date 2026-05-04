package handler

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/core/analysisutil"
	"golang.org/x/tools/go/packages"
)

const ruleNoModelResponse = "handler.no-model-response"

type NoModelResponse struct {
	cfg ruleConfig
}

func NewNoModelResponse(opts ...Option) *NoModelResponse {
	return &NoModelResponse{cfg: newConfig(opts, core.Warning)}
}

func (r *NoModelResponse) Spec() core.RuleSpec {
	return core.RuleSpec{
		ID:              ruleNoModelResponse,
		Description:     "handler and transport responses must not expose domain model types",
		DefaultSeverity: r.cfg.severity,
		Violations: []core.ViolationSpec{
			{
				ID:              ruleNoModelResponse,
				Description:     "handler or transport response exposes a configured domain model type",
				DefaultSeverity: r.cfg.severity,
			},
		},
	}
}

func (r *NoModelResponse) Check(ctx *core.Context) []core.Violation {
	if ctx == nil {
		return nil
	}

	arch := ctx.Arch()
	projectModule := analysisutil.ResolveModuleFromContext(ctx, "")
	if !hasInternalPackages(ctx.Pkgs(), projectModule, arch.Layout.InternalRoot) {
		return []core.Violation{metaLayoutNotSupported(ruleNoModelResponse, projectModule)}
	}
	if arch.Layout.DomainDir == "" {
		return []core.Violation{metaRuleDisabledByConfig(ruleNoModelResponse,
			"Layout.DomainDir is empty (flat layout); handler response model detection requires a domain directory",
			"set Layout.DomainDir to your domain root, or remove handler.NewNoModelResponse() from your RuleSet")}
	}
	if strings.Trim(arch.Structure.ModelPath, "/") == "" {
		return []core.Violation{metaRuleDisabledByConfig(ruleNoModelResponse,
			"Structure.ModelPath is empty; handler response model detection needs a configured model layer",
			"set Structure.ModelPath to your model layer, or remove handler.NewNoModelResponse() from your RuleSet")}
	}

	checker := responseChecker{
		rule:          r,
		ctx:           ctx,
		arch:          arch,
		projectModule: projectModule,
	}
	var violations []core.Violation
	for _, pkg := range ctx.Pkgs() {
		if pkg == nil || !checker.isTargetPackage(pkg.PkgPath) {
			continue
		}
		violations = append(violations, checker.checkPackage(pkg)...)
	}
	return violations
}

type responseChecker struct {
	rule          *NoModelResponse
	ctx           *core.Context
	arch          core.Architecture
	projectModule string
}

func (c responseChecker) checkPackage(pkg *packages.Package) []core.Violation {
	var violations []core.Violation
	for _, file := range pkg.Syntax {
		if analysisutil.IsTestFile(file, pkg.Fset) {
			continue
		}
		filePath := c.filePath(pkg, file.Pos())
		if c.ctx.IsExcluded(filePath) {
			continue
		}
		analysisutil.WalkTypeSpecs(file, pkg.Fset, func(ts *ast.TypeSpec, pos token.Position) {
			if !isResponseTypeName(ts.Name.Name) {
				return
			}
			typ := pkg.TypesInfo.TypeOf(ts.Type)
			if typ == nil {
				return
			}
			if ref, ok := c.findModelRef(typ, pkg.PkgPath, true); ok {
				violations = append(violations, c.violation(filePath, pos.Line, "response type "+quote(ts.Name.Name), ref))
			}
		})
		analysisutil.WalkFuncDecls(file, func(fd *ast.FuncDecl) {
			if fd.Type == nil || fd.Type.Results == nil || !fd.Name.IsExported() {
				return
			}
			for _, field := range fd.Type.Results.List {
				typ := pkg.TypesInfo.TypeOf(field.Type)
				if typ == nil {
					continue
				}
				if ref, ok := c.findModelRef(typ, pkg.PkgPath, false); ok {
					pos := pkg.Fset.Position(fd.Name.Pos())
					violations = append(violations, c.violation(filePath, pos.Line, "handler result "+quote(funcName(fd)), ref))
					break
				}
			}
		})
	}
	return violations
}

func (c responseChecker) findModelRef(t types.Type, currentPkgPath string, expandLocalNamed bool) (modelRef, bool) {
	return c.findModelRefSeen(t, currentPkgPath, expandLocalNamed, make(map[string]bool))
}

func (c responseChecker) findModelRefSeen(t types.Type, currentPkgPath string, expandLocalNamed bool, seen map[string]bool) (modelRef, bool) {
	if t == nil {
		return modelRef{}, false
	}
	t = types.Unalias(t)
	key := types.TypeString(t, typeStringQualifier)
	if seen[key] {
		return modelRef{}, false
	}
	seen[key] = true

	switch typ := t.(type) {
	case *types.Named:
		if ref, ok := c.modelRefForNamed(typ); ok {
			return ref, true
		}
		if expandLocalNamed {
			obj := typ.Obj()
			if obj != nil && obj.Pkg() != nil && obj.Pkg().Path() == currentPkgPath {
				return c.findModelRefSeen(typ.Underlying(), currentPkgPath, expandLocalNamed, seen)
			}
		}
	case *types.Pointer:
		return c.findModelRefSeen(typ.Elem(), currentPkgPath, expandLocalNamed, seen)
	case *types.Slice:
		return c.findModelRefSeen(typ.Elem(), currentPkgPath, expandLocalNamed, seen)
	case *types.Array:
		return c.findModelRefSeen(typ.Elem(), currentPkgPath, expandLocalNamed, seen)
	case *types.Map:
		if ref, ok := c.findModelRefSeen(typ.Key(), currentPkgPath, expandLocalNamed, seen); ok {
			return ref, true
		}
		return c.findModelRefSeen(typ.Elem(), currentPkgPath, expandLocalNamed, seen)
	case *types.Chan:
		return c.findModelRefSeen(typ.Elem(), currentPkgPath, expandLocalNamed, seen)
	case *types.Struct:
		for i := range typ.NumFields() {
			if ref, ok := c.findModelRefSeen(typ.Field(i).Type(), currentPkgPath, expandLocalNamed, seen); ok {
				return ref, true
			}
		}
	case *types.Tuple:
		for i := range typ.Len() {
			if ref, ok := c.findModelRefSeen(typ.At(i).Type(), currentPkgPath, expandLocalNamed, seen); ok {
				return ref, true
			}
		}
	case *types.Signature:
		if ref, ok := c.findModelRefSeen(typ.Results(), currentPkgPath, expandLocalNamed, seen); ok {
			return ref, true
		}
	case *types.Interface:
		for i := range typ.NumExplicitMethods() {
			if ref, ok := c.findModelRefSeen(typ.ExplicitMethod(i).Type(), currentPkgPath, expandLocalNamed, seen); ok {
				return ref, true
			}
		}
	}
	return modelRef{}, false
}

type modelRef struct {
	typeName  string
	domain    string
	modelPath string
}

func (c responseChecker) modelRefForNamed(named *types.Named) (modelRef, bool) {
	if named == nil || named.Obj() == nil || named.Obj().Pkg() == nil {
		return modelRef{}, false
	}
	obj := named.Obj()
	qualifiedName := obj.Pkg().Path() + "." + obj.Name()
	if c.rule.cfg.allowedModelTypes[qualifiedName] {
		return modelRef{}, false
	}
	domain, modelPath, ok := c.domainModelPath(obj.Pkg().Path())
	if !ok {
		return modelRef{}, false
	}
	return modelRef{
		typeName:  obj.Name(),
		domain:    domain,
		modelPath: modelPath,
	}, true
}

func (c responseChecker) domainModelPath(pkgPath string) (string, string, bool) {
	rel := analysisutil.ProjectRelativePackagePath(pkgPath, c.projectModule)
	if rel == "" {
		return "", "", false
	}
	parts := strings.Split(rel, "/")
	if len(parts) < 4 || parts[0] != c.arch.Layout.InternalRoot || parts[1] != c.arch.Layout.DomainDir {
		return "", "", false
	}
	modelParts := strings.Split(strings.Trim(c.arch.Structure.ModelPath, "/"), "/")
	if len(parts) < 3+len(modelParts) {
		return "", "", false
	}
	for i, want := range modelParts {
		if parts[3+i] != want {
			return "", "", false
		}
	}
	return parts[2], strings.Join(modelParts, "/"), true
}

func (c responseChecker) isTargetPackage(pkgPath string) bool {
	rel := analysisutil.ProjectRelativePackagePath(pkgPath, c.projectModule)
	if rel == "" {
		return false
	}
	parts := strings.Split(rel, "/")
	if len(parts) < 2 || parts[0] != c.arch.Layout.InternalRoot {
		return false
	}
	if c.arch.Layout.ServerDir != "" && parts[1] == c.arch.Layout.ServerDir {
		return true
	}
	if c.arch.Layout.DomainDir != "" && len(parts) >= 4 && parts[1] == c.arch.Layout.DomainDir && parts[3] == "handler" {
		return true
	}
	if c.arch.Layout.OrchestrationDir != "" && len(parts) >= 3 && parts[1] == c.arch.Layout.OrchestrationDir && parts[2] == "handler" {
		return true
	}
	return false
}

func (c responseChecker) filePath(pkg *packages.Package, pos token.Pos) string {
	if pkg == nil || pkg.Fset == nil || pos == token.NoPos {
		return ""
	}
	return analysisutil.RelativePathForPackage(pkg, pkg.Fset.Position(pos).Filename)
}

func (c responseChecker) violation(file string, line int, subject string, ref modelRef) core.Violation {
	return core.Violation{
		File:              file,
		Line:              line,
		Rule:              ruleNoModelResponse,
		Message:           fmt.Sprintf("%s exposes domain %q model %q from %s", subject, ref.domain, ref.typeName, ref.modelPath),
		Fix:               "return a transport DTO or app-level output DTO instead of exposing a domain model type",
		DefaultSeverity:   c.rule.cfg.severity,
		EffectiveSeverity: c.rule.cfg.severity,
	}
}

func isResponseTypeName(name string) bool {
	if !ast.IsExported(name) || strings.Contains(name, "Request") {
		return false
	}
	return strings.HasSuffix(name, "Response") || strings.HasSuffix(name, "Result")
}

func funcName(fd *ast.FuncDecl) string {
	if fd == nil {
		return ""
	}
	if fd.Recv == nil || len(fd.Recv.List) == 0 {
		return fd.Name.Name
	}
	recv := analysisutil.ReceiverTypeName(fd.Recv.List[0].Type)
	if recv == "" {
		return fd.Name.Name
	}
	return recv + "." + fd.Name.Name
}

func typeStringQualifier(pkg *types.Package) string {
	if pkg == nil {
		return ""
	}
	return pkg.Path()
}

func quote(s string) string {
	return `"` + s + `"`
}

var _ core.Rule = (*NoModelResponse)(nil)
