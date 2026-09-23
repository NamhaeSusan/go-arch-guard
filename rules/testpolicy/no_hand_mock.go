package testpolicy

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"sort"
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
		ID:              "testpolicy.no-handmock",
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
		// Test packages may split a receiver declaration and its methods over
		// multiple files. Keep internal and external test packages separate.
		fset := token.NewFileSet()
		type mockDecl struct {
			path string
			line int
		}
		structs := make(map[string]mockDecl)
		methods := make(map[string]bool)
		for _, path := range testFiles {
			rel := analysisutil.RelativePathForPackage(pkg, path)
			if ctx.IsExcluded(rel) {
				continue
			}
			file, err := parser.ParseFile(fset, path, nil, 0)
			if err != nil {
				continue
			}
			prefix := file.Name.Name + "."
			for name, line := range collectMockStructs(fset, file) {
				structs[prefix+name] = mockDecl{rel, line}
			}
			analysisutil.WalkFuncDecls(file, func(fd *ast.FuncDecl) {
				if fd.Recv != nil && len(fd.Recv.List) > 0 {
					methods[prefix+analysisutil.ReceiverTypeName(fd.Recv.List[0].Type)] = true
				}
			})
		}
		keys := make([]string, 0, len(structs))
		for key := range structs {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			decl := structs[key]
			if !methods[key] {
				continue
			}
			name := key[strings.IndexByte(key, '.')+1:]
			violations = append(violations, core.Violation{
				File: decl.path, Line: decl.line, Rule: r.Spec().ID,
				Message:         `test file "` + filepath.Base(decl.path) + `" defines hand-rolled mock "` + name + `" with methods - use a mock generator instead`,
				Fix:             "generate the mock with your project's mock generator and import from the dedicated mocks package",
				DefaultSeverity: r.severity, EffectiveSeverity: r.severity,
			})
		}
	}
	return violations
}

func collectMockStructs(fset *token.FileSet, file *ast.File) map[string]int {
	result := make(map[string]int)
	analysisutil.WalkTypeSpecs(file, fset, func(ts *ast.TypeSpec, pos token.Position) {
		if _, ok := ts.Type.(*ast.StructType); !ok {
			return
		}
		lower := strings.ToLower(ts.Name.Name)
		if strings.HasPrefix(lower, "mock") || strings.HasPrefix(lower, "fake") || strings.HasPrefix(lower, "stub") {
			result[ts.Name.Name] = pos.Line
		}
	})
	return result
}

var _ core.Rule = (*NoHandMock)(nil)
