package naming

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"

	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/core/analysisutil"
)

type NoHandMock struct {
	severity core.Severity
}

func NewNoHandMock(opts ...Option) *NoHandMock {
	cfg := newConfig(opts, core.Warning)
	return &NoHandMock{severity: cfg.severity}
}

func (r *NoHandMock) Spec() core.RuleSpec {
	return core.RuleSpec{
		ID:              "testing.no-handmock",
		Description:     "test files must not define hand-rolled mocks with methods",
		DefaultSeverity: r.severity,
	}
}

func (r *NoHandMock) Check(ctx *core.Context) []core.Violation {
	var violations []core.Violation
	seenPkgDirs := make(map[string]bool)
	for _, pkg := range ctx.Pkgs() {
		if len(pkg.GoFiles) == 0 {
			continue
		}
		pkgDir := filepath.Dir(pkg.GoFiles[0])
		if seenPkgDirs[pkgDir] {
			continue
		}
		seenPkgDirs[pkgDir] = true
		testFiles, err := filepath.Glob(filepath.Join(pkgDir, "*_test.go"))
		if err != nil {
			continue
		}
		for _, file := range testFiles {
			relPath := analysisutil.RelativePathForPackage(pkg, file)
			if ctx.IsExcluded(relPath) {
				continue
			}
			violations = append(violations, r.checkFile(file, relPath)...)
		}
	}
	return violations
}

func (r *NoHandMock) checkFile(path, relPath string) []core.Violation {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		return nil
	}
	structs := collectMockStructs(fset, file)
	if len(structs) == 0 {
		return nil
	}
	var violations []core.Violation
	base := filepath.Base(path)
	for _, decl := range file.Decls {
		fd, ok := decl.(*ast.FuncDecl)
		if !ok || fd.Recv == nil || len(fd.Recv.List) == 0 {
			continue
		}
		recvName := analysisutil.ReceiverTypeName(fd.Recv.List[0].Type)
		line, ok := structs[recvName]
		if !ok {
			continue
		}
		violations = append(violations, core.Violation{
			File:              relPath,
			Line:              line,
			Rule:              "testing.no-handmock",
			Message:           `test file "` + base + `" defines hand-rolled mock "` + recvName + `" with methods - use mockery instead`,
			Fix:               "generate mock with mockery and import from mocks/ package",
			DefaultSeverity:   r.severity,
			EffectiveSeverity: r.severity,
		})
		delete(structs, recvName)
	}
	return violations
}

func collectMockStructs(fset *token.FileSet, file *ast.File) map[string]int {
	result := make(map[string]int)
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
			if _, ok := ts.Type.(*ast.StructType); !ok {
				continue
			}
			lower := strings.ToLower(ts.Name.Name)
			if strings.HasPrefix(lower, "mock") || strings.HasPrefix(lower, "fake") || strings.HasPrefix(lower, "stub") {
				result[ts.Name.Name] = fset.Position(ts.Name.Pos()).Line
			}
		}
	}
	return result
}

var _ core.Rule = (*NoHandMock)(nil)
