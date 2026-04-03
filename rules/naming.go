package rules

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"golang.org/x/tools/go/packages"
)

func CheckNaming(pkgs []*packages.Package, opts ...Option) []Violation {
	cfg := NewConfig(opts...)
	m := cfg.model()
	var violations []Violation
	for _, pkg := range pkgs {
		violations = append(violations, checkStutter(pkg, cfg)...)
		violations = append(violations, checkImplSuffix(pkg, cfg)...)
		violations = append(violations, checkSnakeCaseFiles(pkg, cfg)...)
		violations = append(violations, checkRepoFileInterfaceWith(m, pkg, cfg)...)
		violations = append(violations, checkNoLayerSuffixWith(m, pkg, cfg)...)
		violations = append(violations, checkDomainInterfaceRepoOnlyWith(m, pkg, cfg)...)
		violations = append(violations, checkNoHandMock(pkg, cfg)...)
	}
	return violations
}

func checkStutter(pkg *packages.Package, cfg Config) []Violation {
	var violations []Violation
	pkgName := pkg.Name
	pkgNameLen := len([]rune(pkgName))
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
					suggested := string([]rune(name)[pkgNameLen:])
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
	pkgRunes := []rune(pkgName)
	typeRunes := []rune(typeName)
	if len(typeRunes) <= len(pkgRunes) {
		return false
	}
	prefix := strings.ToLower(string(typeRunes[:len(pkgRunes)]))
	if prefix != strings.ToLower(pkgName) {
		return false
	}
	return unicode.IsUpper(typeRunes[len(pkgRunes)])
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
	// Strip compound extensions like .sql, .pb, .gen (e.g. queries.sql.go → queries)
	if idx := strings.IndexByte(name, '.'); idx > 0 {
		name = name[:idx]
	}
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

func checkRepoFileInterfaceWith(m Model, pkg *packages.Package, cfg Config) []Violation {
	if !hasRepoSublayer(m) {
		return nil
	}
	if !isAnyRepoPackage(pkg.PkgPath) {
		return nil
	}

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

		ifaces := collectFileInterfaces(astFile)

		// Check: expected interface must exist
		if _, ok := ifaces[expected]; !ok {
			violations = append(violations, Violation{
				File:     relPath,
				Rule:     "naming.repo-file-interface",
				Message:  `file "` + base + `" in repo/ must contain interface "` + expected + `"`,
				Fix:      `add "type ` + expected + ` interface { ... }" or rename the file`,
				Severity: cfg.Sev,
			})
		}

		// Check: only one interface per file
		if len(ifaces) > 1 {
			var extra []string
			for name := range ifaces {
				if name != expected {
					extra = append(extra, name)
				}
			}
			sort.Strings(extra)
			violations = append(violations, Violation{
				File:     relPath,
				Rule:     "naming.repo-file-extra-interface",
				Message:  `file "` + base + `" in repo/ must define only "` + expected + `", found extra: ` + strings.Join(extra, ", "),
				Fix:      "move each extra interface to its own file (e.g. " + strings.ToLower(extra[0]) + ".go)",
				Severity: cfg.Sev,
			})
		}

		// Check: interface method count
		if cfg.MaxRepoInterfaceMethods > 0 {
			for name, iface := range ifaces {
				methodCount := len(iface.Methods.List)
				if methodCount > cfg.MaxRepoInterfaceMethods {
					violations = append(violations, Violation{
						File:     relPath,
						Rule:     "naming.repo-interface-too-large",
						Message:  fmt.Sprintf(`interface "%s" has %d methods (max %d)`, name, methodCount, cfg.MaxRepoInterfaceMethods),
						Fix:      "split into smaller, focused interfaces",
						Severity: cfg.Sev,
					})
				}
			}
		}
	}
	return violations
}

func collectFileInterfaces(file *ast.File) map[string]*ast.InterfaceType {
	result := make(map[string]*ast.InterfaceType)
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
			if iface, ok := ts.Type.(*ast.InterfaceType); ok {
				result[ts.Name.Name] = iface
			}
		}
	}
	return result
}

// portSublayers are layers that serve as pure interface (port) definitions.
var portSublayers = []string{"repo", "gateway"}

func hasRepoSublayer(m Model) bool {
	for _, sl := range m.Sublayers {
		for _, ps := range portSublayers {
			if sl == ps || strings.HasSuffix(sl, "/"+ps) {
				return true
			}
		}
	}
	return false
}

func isAnyRepoPackage(pkgPath string) bool {
	idx := strings.Index(pkgPath, "/internal/")
	if idx < 0 {
		return false
	}
	rel := pkgPath[idx+len("/internal/"):]
	for _, ps := range portSublayers {
		if strings.HasSuffix(rel, "/"+ps) || strings.Contains(rel, "/"+ps+"/") {
			return true
		}
	}
	return false
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

var bannedLayerSuffixes = []string{
	"_svc", "_service", "_repo", "_repository",
	"_handler", "_controller", "_model", "_entity",
	"_store", "_persistence",
}

func checkNoLayerSuffixWith(m Model, pkg *packages.Package, cfg Config) []Violation {
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
		if !m.LayerDirNames[dir] {
			continue
		}
		// Skip files matching a TypePattern prefix (e.g. worker_xxx.go in worker/)
		if isTypePatternFile(m, dir, base) {
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

func isDomainPackageWith(m Model, pkgPath string) bool {
	return strings.Contains(pkgPath, "/internal/"+m.DomainDir+"/")
}

func isRepoPackageWith(m Model, pkgPath string) bool {
	for _, sl := range m.Sublayers {
		if sl == "repo" || strings.HasSuffix(sl, "/repo") {
			// Check if pkgPath ends with this sublayer or is inside it
			suffix := "/" + sl
			if strings.HasSuffix(pkgPath, suffix) || strings.Contains(pkgPath, suffix+"/") {
				return true
			}
		}
	}
	return false
}

func repoSublayerName(m Model) string {
	for _, sl := range m.Sublayers {
		if sl == "repo" || strings.HasSuffix(sl, "/repo") {
			return sl
		}
	}
	return "core/repo"
}

// checkDomainInterfaceRepoOnlyWith flags interface declarations in any domain
// sublayer other than repo/. In DDD models, interfaces (ports) belong in the
// repo sublayer; handler/, app/, core/svc/ etc. should not define interfaces.
func checkDomainInterfaceRepoOnlyWith(m Model, pkg *packages.Package, cfg Config) []Violation {
	if !m.RequireAlias {
		return nil
	}
	if !isDomainPackageWith(m, pkg.PkgPath) || isRepoPackageWith(m, pkg.PkgPath) {
		return nil
	}
	repoName := repoSublayerName(m)
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
				// Direct interface definition
				if _, isIface := ts.Type.(*ast.InterfaceType); isIface {
					pos := pkg.Fset.Position(ts.Name.Pos())
					violations = append(violations, Violation{
						File:     relativePathForPackage(pkg, pos.Filename),
						Line:     pos.Line,
						Rule:     "naming.domain-interface-repo-only",
						Message:  `interface "` + ts.Name.Name + `" must be defined in ` + repoName + `/, not in ` + path.Base(path.Dir(pkg.PkgPath)) + `/`,
						Fix:      "move interface to " + repoName + "/",
						Severity: cfg.Sev,
					})
				}
				// Type alias from repo sublayer (re-exporting interface)
				if ts.Assign != 0 {
					if sel, ok := ts.Type.(*ast.SelectorExpr); ok {
						if ident, ok := sel.X.(*ast.Ident); ok {
							impPath := resolveIdentImportPath(file, ident.Name)
							if isRepoImportPath(m, impPath) {
								pos := pkg.Fset.Position(ts.Name.Pos())
								violations = append(violations, Violation{
									File:     relativePathForPackage(pkg, pos.Filename),
									Line:     pos.Line,
									Rule:     "naming.domain-interface-repo-only",
									Message:  `type alias "` + ts.Name.Name + `" re-exports interface from ` + repoName + ` — suspected cross-domain dependency; use ` + m.OrchestrationDir + `/ instead`,
									Fix:      "remove alias and move cross-domain coordination to " + m.OrchestrationDir + "/",
									Severity: cfg.Sev,
								})
							}
						}
					}
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

func isRepoImportPath(m Model, impPath string) bool {
	for _, sl := range m.Sublayers {
		if sl == "repo" || strings.HasSuffix(sl, "/repo") {
			if strings.Contains(impPath, "/"+sl) {
				return true
			}
		}
	}
	return false
}

func isTypePatternFile(m Model, dir, filename string) bool {
	name := strings.TrimSuffix(filename, ".go")
	for _, tp := range m.TypePatterns {
		if tp.Dir == dir && strings.HasPrefix(name, tp.FilePrefix+"_") {
			return true
		}
	}
	return false
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
