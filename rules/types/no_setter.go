package types

import (
	"fmt"
	"go/ast"
	"strings"
	"unicode"

	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/core/analysisutil"
	"golang.org/x/tools/go/packages"
)

const setterForbiddenID = "setter.forbidden"

var NoSetterDefaultSpec = core.RuleSpec{
	ID:              "types.no-setter",
	Description:     "exported setter methods on pointer receivers are forbidden",
	DefaultSeverity: core.Warning,
	Violations: []core.ViolationSpec{
		{ID: setterForbiddenID, Description: "exported setter method on pointer receiver", DefaultSeverity: core.Warning},
	},
}

type NoSetter struct {
	severity core.Severity
}

func NewNoSetter(opts ...Option) *NoSetter {
	r := &NoSetter{severity: core.Warning}
	for _, opt := range opts {
		opt.applyNoSetter(r)
	}
	return r
}

func (r *NoSetter) Spec() core.RuleSpec {
	return specWithSeverity(NoSetterDefaultSpec, r.severity)
}

func (r *NoSetter) Check(ctx *core.Context) []core.Violation {
	if ctx == nil {
		return nil
	}

	projectModule := analysisutil.ResolveModuleFromContext(ctx, "")
	var violations []core.Violation
	for _, pkg := range ctx.Pkgs() {
		if autoExcludedPackage(pkg) {
			continue
		}
		pkgRelPath := analysisutil.ProjectRelativePackagePath(pkg.PkgPath, projectModule)
		if ctx.IsExcluded(pkgRelPath) {
			continue
		}
		for _, file := range pkg.Syntax {
			filename := pkg.Fset.Position(file.Pos()).Filename
			relPath := analysisutil.RelativePathForPackage(pkg, filename)
			if autoExcludedFile(relPath) || ctx.IsExcluded(relPath) {
				continue
			}
			violations = append(violations, r.checkFile(pkg, file, relPath)...)
		}
	}
	return violations
}

func (r *NoSetter) checkFile(pkg *packages.Package, file *ast.File, relPath string) []core.Violation {
	var violations []core.Violation
	for _, decl := range file.Decls {
		fd, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if !isExportedSetter(fd.Name.Name) || !hasPointerReceiver(fd) || !hasParams(fd) || isFluentBuilder(fd) {
			continue
		}
		recvType := receiverTypeString(fd)
		pos := pkg.Fset.Position(fd.Pos())
		field := strings.TrimPrefix(fd.Name.Name, "Set")
		violations = append(violations, core.Violation{
			Rule:              setterForbiddenID,
			File:              relPath,
			Line:              pos.Line,
			Message:           fmt.Sprintf("exported setter %s on *%s is forbidden", fd.Name.Name, recvType),
			Fix:               fmt.Sprintf("add %s as an explicit New%s parameter, or use With%s only when it is genuinely optional", field, recvType, field),
			DefaultSeverity:   r.severity,
			EffectiveSeverity: r.severity,
		})
	}
	return violations
}

func isExportedSetter(name string) bool {
	if name == "Set" {
		return true
	}
	if !strings.HasPrefix(name, "Set") {
		return false
	}
	rest := name[3:]
	if rest == "" {
		return false
	}
	return unicode.IsUpper(rune(rest[0]))
}

func hasPointerReceiver(fd *ast.FuncDecl) bool {
	if fd.Recv == nil || len(fd.Recv.List) == 0 {
		return false
	}
	_, ok := fd.Recv.List[0].Type.(*ast.StarExpr)
	return ok
}

func hasParams(fd *ast.FuncDecl) bool {
	return fd.Type.Params != nil && len(fd.Type.Params.List) > 0
}

func receiverTypeString(fd *ast.FuncDecl) string {
	if fd.Recv == nil || len(fd.Recv.List) == 0 {
		return ""
	}
	return receiverTypeName(fd.Recv.List[0].Type)
}

func isFluentBuilder(fd *ast.FuncDecl) bool {
	if fd.Type.Results == nil || len(fd.Type.Results.List) != 1 {
		return false
	}
	result := fd.Type.Results.List[0].Type
	if star, ok := result.(*ast.StarExpr); ok {
		result = star.X
	}
	ident, ok := result.(*ast.Ident)
	return ok && ident.Name == receiverTypeString(fd)
}

func autoExcludedFile(filename string) bool {
	return strings.HasSuffix(filename, "_test.go") || hasPathSegment(filename, "testdata") || hasPathSegment(filename, "mocks")
}

func autoExcludedPackage(pkg *packages.Package) bool {
	if pkg == nil {
		return true
	}
	return hasPathSegment(pkg.PkgPath, "testdata") || hasPathSegment(pkg.PkgPath, "mocks")
}

func hasPathSegment(path, segment string) bool {
	path = analysisutil.NormalizeMatchPath(path)
	for _, part := range strings.Split(path, "/") {
		if part == segment {
			return true
		}
	}
	return false
}
