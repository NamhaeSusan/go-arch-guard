package types

import (
	"fmt"
	"go/ast"
	"path/filepath"
	"strings"

	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/core/analysisutil"
	"golang.org/x/tools/go/packages"
)

const (
	typePatternMismatchID      = "naming.type-pattern-mismatch"
	typePatternMissingMethodID = "naming.type-pattern-missing-method"
)

// TypePatternDefaultSpec returns a fresh copy of the rule's static metadata
// (with the hard-coded default severity, before any construction-time
// override). The returned value is independent — mutating it has no effect
// on subsequent calls.
func TypePatternDefaultSpec() core.RuleSpec {
	return core.RuleSpec{
		ID:              "types.type-pattern",
		Description:     "files matching configured type patterns must define expected types and methods",
		DefaultSeverity: core.Error,
		Violations: []core.ViolationSpec{
			{ID: typePatternMismatchID, Description: "file does not define the expected exported type", DefaultSeverity: core.Error},
			{ID: typePatternMissingMethodID, Description: "type does not define the required method", DefaultSeverity: core.Error},
		},
	}
}

type TypePattern struct {
	severity core.Severity
}

func NewTypePattern(opts ...Option) *TypePattern {
	cfg := newConfig(opts, core.Error)
	return &TypePattern{severity: cfg.severity}
}

func (r *TypePattern) Spec() core.RuleSpec {
	return specWithSeverity(TypePatternDefaultSpec(), r.severity)
}

func (r *TypePattern) Check(ctx *core.Context) []core.Violation {
	if ctx == nil {
		return nil
	}
	patterns := ctx.Arch().Structure.TypePatterns
	if len(patterns) == 0 {
		return nil
	}

	var violations []core.Violation
	for _, pkg := range ctx.Pkgs() {
		for _, pattern := range patterns {
			violations = append(violations, r.checkPackage(ctx, pkg, pattern)...)
		}
	}
	return violations
}

func (r *TypePattern) checkPackage(ctx *core.Context, pkg *packages.Package, pattern core.TypePattern) []core.Violation {
	if pkg == nil || !strings.HasSuffix(pkg.PkgPath, "/"+pattern.Dir) {
		return nil
	}

	methods := collectMethods(pkg)
	var violations []core.Violation
	for _, file := range pkg.Syntax {
		filename := pkg.Fset.Position(file.Pos()).Filename
		base := filepath.Base(filename)
		if strings.HasSuffix(base, "_test.go") {
			continue
		}
		relPath := analysisutil.RelativePathForPackage(pkg, filename)
		if ctx.IsExcluded(relPath) {
			continue
		}
		name := strings.TrimSuffix(base, ".go")
		suffix, ok := strings.CutPrefix(name, pattern.FilePrefix+"_")
		if !ok {
			continue
		}

		expectedType := analysisutil.SnakeToPascal(suffix) + pattern.TypeSuffix
		if !hasExportedType(file, expectedType) {
			violations = append(violations, core.Violation{
				File:              relPath,
				Rule:              typePatternMismatchID,
				Message:           fmt.Sprintf("file %q must define type %q", base, expectedType),
				Fix:               fmt.Sprintf("add type %s struct { ... }", expectedType),
				DefaultSeverity:   r.severity,
				EffectiveSeverity: r.severity,
			})
			continue
		}

		if pattern.RequireMethod != "" && !methods[expectedType+"."+pattern.RequireMethod] {
			violations = append(violations, core.Violation{
				File:              relPath,
				Rule:              typePatternMissingMethodID,
				Message:           fmt.Sprintf("type %q must have a %s method", expectedType, pattern.RequireMethod),
				Fix:               fmt.Sprintf("add func (w *%s) %s(...) { ... }", expectedType, pattern.RequireMethod),
				DefaultSeverity:   r.severity,
				EffectiveSeverity: r.severity,
			})
		}
	}
	return violations
}

func specWithSeverity(spec core.RuleSpec, severity core.Severity) core.RuleSpec {
	spec.DefaultSeverity = severity
	spec.Violations = append([]core.ViolationSpec(nil), spec.Violations...)
	for i := range spec.Violations {
		spec.Violations[i].DefaultSeverity = severity
	}
	return spec
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
			if ts.Name.Name == name && ts.Name.IsExported() {
				return true
			}
		}
	}
	return false
}

func collectMethods(pkg *packages.Package) map[string]bool {
	result := make(map[string]bool)
	for _, file := range pkg.Syntax {
		for _, decl := range file.Decls {
			fd, ok := decl.(*ast.FuncDecl)
			if !ok || fd.Recv == nil || len(fd.Recv.List) == 0 {
				continue
			}
			typeName := analysisutil.ReceiverTypeName(fd.Recv.List[0].Type)
			if typeName != "" {
				result[typeName+"."+fd.Name.Name] = true
			}
		}
	}
	return result
}
