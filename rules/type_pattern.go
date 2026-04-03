package rules

import (
	"fmt"
	"go/ast"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
)

// CheckTypePatterns enforces that files matching a TypePattern in the model
// define the expected exported type with the required method.
func CheckTypePatterns(pkgs []*packages.Package, opts ...Option) []Violation {
	cfg := NewConfig(opts...)
	m := cfg.model()
	if len(m.TypePatterns) == 0 {
		return nil
	}

	var violations []Violation
	for _, pkg := range pkgs {
		for _, tp := range m.TypePatterns {
			violations = append(violations, checkTypePattern(pkg, tp, cfg)...)
		}
	}
	return violations
}

// checkTypePattern validates naming conventions for types in flat-layout
// directories directly under internal/ (e.g. internal/worker/, internal/aggregate/).
func checkTypePattern(pkg *packages.Package, tp TypePattern, cfg Config) []Violation {
	// TypePattern applies only to flat-layout packages directly under internal/ (e.g. internal/worker).
	if !strings.HasSuffix(pkg.PkgPath, "/internal/"+tp.Dir) {
		return nil
	}

	var violations []Violation
	methods := collectMethods(pkg)

	for _, file := range pkg.Syntax {
		filename := pkg.Fset.Position(file.Pos()).Filename
		base := filepath.Base(filename)
		if strings.HasSuffix(base, "_test.go") {
			continue
		}
		relPath := relativePathForPackage(pkg, filename)
		if cfg.IsExcluded(relPath) {
			continue
		}

		name := strings.TrimSuffix(base, ".go")
		suffix, ok := strings.CutPrefix(name, tp.FilePrefix+"_")
		if !ok {
			continue
		}

		expectedType := snakeToPascal(suffix) + tp.TypeSuffix

		if !hasExportedType(file, expectedType) {
			violations = append(violations, Violation{
				File:     relPath,
				Rule:     "naming.type-pattern-mismatch",
				Message:  fmt.Sprintf("file %q must define type %q", base, expectedType),
				Fix:      fmt.Sprintf("add \"type %s struct { ... }\"", expectedType),
				Severity: cfg.Sev,
			})
			continue
		}

		if !methods[expectedType+"."+tp.RequireMethod] {
			violations = append(violations, Violation{
				File:     relPath,
				Rule:     "naming.type-pattern-missing-method",
				Message:  fmt.Sprintf("type %q must have a %s method", expectedType, tp.RequireMethod),
				Fix:      fmt.Sprintf("add \"func (w *%s) %s(ctx context.Context) error { ... }\"", expectedType, tp.RequireMethod),
				Severity: cfg.Sev,
			})
		}
	}
	return violations
}

func hasExportedType(file *ast.File, name string) bool {
	for _, decl := range file.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range gd.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			if ts.Name.Name == name {
				return true
			}
		}
	}
	return false
}
