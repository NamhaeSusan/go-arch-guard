package rules

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
	"unicode"

	"golang.org/x/tools/go/packages"
)

func CheckNaming(pkgs []*packages.Package, opts ...Option) []Violation {
	cfg := NewConfig(opts...)
	var violations []Violation
	for _, pkg := range pkgs {
		violations = append(violations, checkStutter(pkg, cfg)...)
		violations = append(violations, checkImplSuffix(pkg, cfg)...)
		violations = append(violations, checkSnakeCaseFiles(pkg, cfg)...)
		violations = append(violations, checkRepoFileInterface(pkg, cfg)...)
		violations = append(violations, checkNoLayerSuffix(pkg, cfg)...)
		violations = append(violations, checkHandlerNoInterface(pkg, cfg)...)
		violations = append(violations, checkAppNoInterface(pkg, cfg)...)
		violations = append(violations, checkNoHandMock(pkg, cfg)...)
	}
	return violations
}

func checkStutter(pkg *packages.Package, cfg Config) []Violation {
	var violations []Violation
	pkgName := pkg.Name
	for _, file := range pkg.Syntax {
		filePath := relativePathForPackage(pkg, pkg.Fset.Position(file.Pos()).Filename)
		if cfg.IsExcluded(filePath) {
			continue
		}
		for _, decl := range file.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			for _, spec := range gd.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok || !ts.Name.IsExported() {
					continue
				}
				name := ts.Name.Name
				if stutters(pkgName, name) {
					suggested := strings.TrimPrefix(strings.ToLower(name), strings.ToLower(pkgName))
					if len(suggested) > 0 {
						suggested = strings.ToUpper(suggested[:1]) + suggested[1:]
					}
					pos := pkg.Fset.Position(ts.Name.Pos())
					violations = append(violations, Violation{
						File:     relativePathForPackage(pkg, pos.Filename),
						Line:     pos.Line,
						Rule:     "naming.no-stutter",
						Message:  `type "` + name + `" stutters with package "` + pkgName + `"`,
						Fix:      `rename to "` + suggested + `"`,
						Severity: cfg.Sev,
					})
				}
			}
		}
	}
	return violations
}

func stutters(pkgName, typeName string) bool {
	if len(typeName) <= len(pkgName) {
		return false
	}
	prefix := strings.ToLower(typeName[:len(pkgName)])
	if prefix != strings.ToLower(pkgName) {
		return false
	}
	next := rune(typeName[len(pkgName)])
	return unicode.IsUpper(next)
}

func checkImplSuffix(pkg *packages.Package, cfg Config) []Violation {
	var violations []Violation
	for _, file := range pkg.Syntax {
		filePath := relativePathForPackage(pkg, pkg.Fset.Position(file.Pos()).Filename)
		if cfg.IsExcluded(filePath) {
			continue
		}
		for _, decl := range file.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			for _, spec := range gd.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok || !ts.Name.IsExported() {
					continue
				}
				if strings.HasSuffix(ts.Name.Name, "Impl") {
					pos := pkg.Fset.Position(ts.Name.Pos())
					violations = append(violations, Violation{
						File:     relativePathForPackage(pkg, pos.Filename),
						Line:     pos.Line,
						Rule:     "naming.no-impl-suffix",
						Message:  `type "` + ts.Name.Name + `" uses banned suffix "Impl"`,
						Fix:      "rename without Impl suffix",
						Severity: cfg.Sev,
					})
				}
			}
		}
	}
	return violations
}

func checkSnakeCaseFiles(pkg *packages.Package, cfg Config) []Violation {
	var violations []Violation
	seen := make(map[string]bool)
	for _, f := range pkg.GoFiles {
		if seen[f] {
			continue
		}
		seen[f] = true
		relPath := relativePathForPackage(pkg, f)
		if cfg.IsExcluded(relPath) {
			continue
		}
		base := filepath.Base(f)
		if !isSnakeCase(base) {
			violations = append(violations, Violation{
				File:     relPath,
				Rule:     "naming.snake-case-file",
				Message:  `filename "` + base + `" must be snake_case`,
				Fix:      `rename to "` + toSnakeCase(base) + `"`,
				Severity: cfg.Sev,
			})
		}
	}
	return violations
}

func isSnakeCase(filename string) bool {
	name := strings.TrimSuffix(filename, ".go")
	name = strings.TrimSuffix(name, "_test")
	for _, r := range name {
		if !unicode.IsLower(r) && !unicode.IsDigit(r) && r != '_' {
			return false
		}
	}
	return len(name) > 0
}

func toSnakeCase(filename string) string {
	ext := filepath.Ext(filename)
	name := strings.TrimSuffix(filename, ext)
	var result []rune
	for i, r := range name {
		if unicode.IsUpper(r) {
			if i > 0 {
				result = append(result, '_')
			}
			result = append(result, unicode.ToLower(r))
		} else {
			result = append(result, r)
		}
	}
	return string(result) + ext
}

func checkRepoFileInterface(pkg *packages.Package, cfg Config) []Violation {
	if !isRepoPackage(pkg.PkgPath) {
		return nil
	}

	// Iterate Syntax directly instead of cross-referencing GoFiles indices;
	// Syntax corresponds to CompiledGoFiles, which can differ from GoFiles
	// when cgo or generated files are involved.
	var violations []Violation
	for _, astFile := range pkg.Syntax {
		filename := pkg.Fset.Position(astFile.Pos()).Filename
		base := filepath.Base(filename)
		if strings.HasSuffix(base, "_test.go") {
			continue
		}
		relPath := relativePathForPackage(pkg, filename)
		if cfg.IsExcluded(relPath) {
			continue
		}
		expected := snakeToPascal(strings.TrimSuffix(base, ".go"))
		if hasInterface(astFile, expected) {
			continue
		}
		violations = append(violations, Violation{
			File:     relPath,
			Rule:     "naming.repo-file-interface",
			Message:  `file "` + base + `" in repo/ must contain interface "` + expected + `"`,
			Fix:      `add "type ` + expected + ` interface { ... }" or rename the file`,
			Severity: cfg.Sev,
		})
	}
	return violations
}

func isRepoPackage(pkgPath string) bool {
	return strings.HasSuffix(pkgPath, "/repo") || strings.Contains(pkgPath, "/repo/")
}

func snakeToPascal(s string) string {
	parts := strings.Split(s, "_")
	var b strings.Builder
	for _, p := range parts {
		if len(p) == 0 {
			continue
		}
		b.WriteString(strings.ToUpper(p[:1]))
		b.WriteString(p[1:])
	}
	return b.String()
}

func hasInterface(file *ast.File, name string) bool {
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
				if _, isIface := ts.Type.(*ast.InterfaceType); isIface {
					return true
				}
			}
		}
	}
	return false
}

var bannedLayerSuffixes = []string{
	"_svc", "_service", "_repo", "_repository",
	"_handler", "_controller", "_model", "_entity",
	"_store", "_persistence",
}

func checkNoLayerSuffix(pkg *packages.Package, cfg Config) []Violation {
	var violations []Violation
	seen := make(map[string]bool)
	for _, f := range pkg.GoFiles {
		if seen[f] {
			continue
		}
		seen[f] = true
		base := filepath.Base(f)
		if strings.HasSuffix(base, "_test.go") {
			continue
		}
		relPath := relativePathForPackage(pkg, f)
		if cfg.IsExcluded(relPath) {
			continue
		}
		dir := filepath.Base(filepath.Dir(f))
		if !layerDirNames[dir] {
			continue
		}
		name := strings.TrimSuffix(base, ".go")
		for _, suffix := range bannedLayerSuffixes {
			if trimmed, ok := strings.CutSuffix(name, suffix); ok {
				suggested := trimmed + ".go"
				violations = append(violations, Violation{
					File:     relPath,
					Rule:     "naming.no-layer-suffix",
					Message:  `filename "` + base + `" has redundant layer suffix "` + suffix + `"`,
					Fix:      `rename to "` + suggested + `"`,
					Severity: cfg.Sev,
				})
				break
			}
		}
	}
	return violations
}

func isHandlerPackage(pkgPath string) bool {
	return strings.HasSuffix(pkgPath, "/handler") ||
		strings.Contains(pkgPath, "/handler/")
}

func checkHandlerNoInterface(pkg *packages.Package, cfg Config) []Violation {
	if !isHandlerPackage(pkg.PkgPath) {
		return nil
	}
	var violations []Violation
	for _, file := range pkg.Syntax {
		filePath := relativePathForPackage(pkg, pkg.Fset.Position(file.Pos()).Filename)
		if cfg.IsExcluded(filePath) {
			continue
		}
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
				if _, isIface := ts.Type.(*ast.InterfaceType); isIface {
					pos := pkg.Fset.Position(ts.Name.Pos())
					violations = append(violations, Violation{
						File:     relativePathForPackage(pkg, pos.Filename),
						Line:     pos.Line,
						Rule:     "naming.handler-no-interface",
						Message:  `handler package defines interface "` + ts.Name.Name + `" — inject via app.Service or orchestration instead`,
						Fix:      "remove interface and use concrete type from app/ or orchestration/",
						Severity: cfg.Sev,
					})
				}
			}
		}
	}
	return violations
}

func isAppPackage(pkgPath string) bool {
	return strings.HasSuffix(pkgPath, "/app") ||
		strings.Contains(pkgPath, "/app/")
}

func checkAppNoInterface(pkg *packages.Package, cfg Config) []Violation {
	if !isAppPackage(pkg.PkgPath) {
		return nil
	}
	var violations []Violation
	for _, file := range pkg.Syntax {
		filePath := relativePathForPackage(pkg, pkg.Fset.Position(file.Pos()).Filename)
		if cfg.IsExcluded(filePath) {
			continue
		}
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
				if _, isIface := ts.Type.(*ast.InterfaceType); isIface {
					pos := pkg.Fset.Position(ts.Name.Pos())
					violations = append(violations, Violation{
						File:     relativePathForPackage(pkg, pos.Filename),
						Line:     pos.Line,
						Rule:     "naming.app-no-interface",
						Message:  `app package defines interface "` + ts.Name.Name + `" — define in core/repo or core/svc instead`,
						Fix:      "move interface to core/repo/ or core/svc/",
						Severity: cfg.Sev,
					})
				}
			}
		}
	}
	return violations
}

var handMockPrefixes = []string{"mock", "fake", "stub"}

// checkNoHandMock detects hand-rolled mock structs in _test.go files.
// Limitation: external test packages (package foo_test) are not covered
// because the loader does not set Tests: true, so those packages have
// empty GoFiles and are skipped.
func checkNoHandMock(pkg *packages.Package, cfg Config) []Violation {
	if len(pkg.GoFiles) == 0 {
		return nil
	}
	pkgDir := filepath.Dir(pkg.GoFiles[0])
	testFiles, err := filepath.Glob(filepath.Join(pkgDir, "*_test.go"))
	if err != nil || len(testFiles) == 0 {
		return nil
	}

	var violations []Violation
	fset := token.NewFileSet()
	seen := make(map[string]bool)
	for _, f := range testFiles {
		if seen[f] {
			continue
		}
		seen[f] = true
		relPath := relativePathForPackage(pkg, f)
		if cfg.IsExcluded(relPath) {
			continue
		}
		astFile, err := parser.ParseFile(fset, f, nil, 0)
		if err != nil {
			continue
		}
		structs := collectMockStructs(fset, astFile)
		if len(structs) == 0 {
			continue
		}
		baseName := filepath.Base(f)
		for _, decl := range astFile.Decls {
			fd, ok := decl.(*ast.FuncDecl)
			if !ok || fd.Recv == nil || len(fd.Recv.List) == 0 {
				continue
			}
			recvName := receiverTypeName(fd.Recv.List[0].Type)
			if line, ok := structs[recvName]; ok {
				violations = append(violations, Violation{
					File:     relPath,
					Line:     line,
					Rule:     "naming.no-handmock",
					Message:  `test file "` + baseName + `" defines hand-rolled mock "` + recvName + `" with methods — use mockery instead`,
					Fix:      "generate mock with mockery and import from mocks/ package",
					Severity: cfg.Sev,
				})
				delete(structs, recvName)
			}
		}
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
			name := ts.Name.Name
			lower := strings.ToLower(name)
			for _, prefix := range handMockPrefixes {
				if strings.HasPrefix(lower, prefix) {
					result[name] = fset.Position(ts.Name.Pos()).Line
					break
				}
			}
		}
	}
	return result
}

func receiverTypeName(expr ast.Expr) string {
	if star, ok := expr.(*ast.StarExpr); ok {
		expr = star.X
	}
	if ident, ok := expr.(*ast.Ident); ok {
		return ident.Name
	}
	return ""
}
