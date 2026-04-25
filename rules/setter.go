package rules

import (
	"fmt"
	"go/ast"
	"strings"
	"unicode"

	"golang.org/x/tools/go/packages"
)

// CheckNoSetters flags exported setter methods (Set-prefixed, pointer receiver,
// with at least one parameter). Fluent builders — methods returning the
// receiver type — are exempt. Test files and packages under testdata/ or
// mocks/ are auto-excluded.
//
// Default severity is Warning; use WithSeverity(Error) for strict enforcement.
// Use WithExclude(...) to exempt specific paths.
func CheckNoSetters(pkgs []*packages.Package, opts ...Option) []Violation {
	cfg := NewConfig(opts...)
	sev := Warning
	if cfg.SeverityExplicit() {
		sev = cfg.Sev
	}

	var violations []Violation
	for _, pkg := range pkgs {
		if isAutoExcludedPkg(pkg) {
			continue
		}
		pkgRelPath := projectRelativePackagePath(pkg.PkgPath, resolveModule(pkgs, ""))
		if cfg.IsExcluded(pkgRelPath) {
			continue
		}
		for _, file := range pkg.Syntax {
			filename := pkg.Fset.Position(file.Pos()).Filename
			if isAutoExcludedFile(filename) {
				continue
			}
			for _, decl := range file.Decls {
				fd, ok := decl.(*ast.FuncDecl)
				if !ok {
					continue
				}
				if !isExportedSetter(fd.Name.Name) {
					continue
				}
				if !hasPointerReceiver(fd) {
					continue
				}
				if !hasParams(fd) {
					continue
				}
				if isFluentBuilder(fd) {
					continue
				}
				pos := pkg.Fset.Position(fd.Pos())
				recvType := receiverTypeString(fd)
				violations = append(violations, Violation{
					Rule:              "setter.forbidden",
					File:              pos.Filename,
					Line:              pos.Line,
					Message:           fmt.Sprintf("exported setter %s on *%s — prefer explicit constructor parameter; use With-pattern only if truly optional", fd.Name.Name, recvType),
					Fix:               fmt.Sprintf("add %s as an explicit parameter to New%s(...). Use With%s option only when %s is genuinely optional with multiple combinations.", strings.TrimPrefix(fd.Name.Name, "Set"), recvType, strings.TrimPrefix(fd.Name.Name, "Set"), strings.TrimPrefix(fd.Name.Name, "Set")),
					DefaultSeverity:   sev,
					EffectiveSeverity: sev,
				})
			}
		}
	}
	return violations
}

// isExportedSetter returns true if name is "Set" or starts with "Set" followed
// by an uppercase letter (i.e. exported setter convention).
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

// hasPointerReceiver returns true if the method has a pointer (*T) receiver.
func hasPointerReceiver(fd *ast.FuncDecl) bool {
	if fd.Recv == nil || len(fd.Recv.List) == 0 {
		return false
	}
	_, ok := fd.Recv.List[0].Type.(*ast.StarExpr)
	return ok
}

// hasParams returns true if the function has at least one parameter.
func hasParams(fd *ast.FuncDecl) bool {
	return fd.Type.Params != nil && len(fd.Type.Params.List) > 0
}

// receiverTypeString extracts the base type name from a pointer receiver (e.g. "Order" from "*Order").
func receiverTypeString(fd *ast.FuncDecl) string {
	if fd.Recv == nil || len(fd.Recv.List) == 0 {
		return ""
	}
	expr := fd.Recv.List[0].Type
	if star, ok := expr.(*ast.StarExpr); ok {
		expr = star.X
	}
	if ident, ok := expr.(*ast.Ident); ok {
		return ident.Name
	}
	return ""
}

// isFluentBuilder returns true if the method has exactly one result whose type
// matches the pointer receiver type (e.g. func (b *Builder) SetX(...) *Builder).
func isFluentBuilder(fd *ast.FuncDecl) bool {
	if fd.Type.Results == nil || len(fd.Type.Results.List) != 1 {
		return false
	}
	result := fd.Type.Results.List[0].Type
	star, ok := result.(*ast.StarExpr)
	if !ok {
		return false
	}
	resultIdent, ok := star.X.(*ast.Ident)
	if !ok {
		return false
	}
	return resultIdent.Name == receiverTypeString(fd)
}

// isAutoExcludedFile returns true for test files.
func isAutoExcludedFile(filename string) bool {
	return strings.HasSuffix(filename, "_test.go")
}

// isAutoExcludedPkg returns true for packages under testdata/ or mocks/ directories.
func isAutoExcludedPkg(pkg *packages.Package) bool {
	return containsPathSegment(pkg.PkgPath, "testdata") ||
		containsPathSegment(pkg.PkgPath, "mocks")
}

func containsPathSegment(path, segment string) bool {
	for _, part := range strings.Split(path, "/") {
		if part == segment {
			return true
		}
	}
	return false
}
