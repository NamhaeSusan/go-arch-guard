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
	if !hasPortSublayer(m) {
		return nil
	}
	if matchPortSublayer(m, pkg.PkgPath) == "" {
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

		ifaces := collectInterfacesFromFile(astFile, false)

		// Check: expected interface must exist
		if _, ok := ifaces[expected]; !ok {
			violations = append(violations, Violation{
				File:     relPath,
				Rule:     "structure.repo-file-interface",
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
				Rule:     "structure.repo-file-extra-interface",
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
						Rule:     "interface.too-many-methods",
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

func layerSuffixes(m Model) []string {
	seen := make(map[string]bool)
	var out []string
	for dir := range m.LayerDirNames {
		suffix := "_" + dir
		if !seen[suffix] {
			seen[suffix] = true
			out = append(out, suffix)
		}
	}
	return out
}

func checkNoLayerSuffixWith(m Model, pkg *packages.Package, cfg Config) []Violation {
	banned := layerSuffixes(m)
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
		for _, suffix := range banned {
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
	return matchPortSublayer(m, pkgPath) != ""
}

// isRepoPortName reports whether the given name follows the repository-port
// naming convention (ends in "Repository" or "Repo"). Only repo-port interfaces
// must live in the repo sublayer; consumer-defined interfaces (the Go idiom
// where a package declares the small interface it consumes) are allowed in
// app/, handler/, svc/ etc. at the site where they are used.
func isRepoPortName(name string) bool {
	return strings.HasSuffix(name, "Repository") || strings.HasSuffix(name, "Repo")
}

// checkDomainInterfaceRepoOnlyWith flags repository-port interface declarations
// (names ending in "Repository"/"Repo") declared outside the repo sublayer, and
// type aliases that re-export an interface from another domain's repo package.
// Consumer-defined interfaces are allowed wherever they are used.
func checkDomainInterfaceRepoOnlyWith(m Model, pkg *packages.Package, cfg Config) []Violation {
	if !m.RequireAlias {
		return nil
	}
	if !isDomainPackageWith(m, pkg.PkgPath) || isRepoPackageWith(m, pkg.PkgPath) {
		return nil
	}
	repoName := portSublayerName(m)
	var violations []Violation
	for _, file := range pkg.Syntax {
		filePath := relativePathForPackage(pkg, pkg.Fset.Position(file.Pos()).Filename)
		if cfg.IsExcluded(filePath) {
			continue
		}
		for _, info := range inspectTypeSpecs(file, pkg.Fset) {
			relFile := relativePathForPackage(pkg, info.Pos.Filename)
			if info.IsIface && isRepoPortName(info.Name) {
				violations = append(violations, Violation{
					File:     relFile,
					Line:     info.Pos.Line,
					Rule:     "structure.interface-placement",
					Message:  `interface "` + info.Name + `" matches repository-port naming and must be defined in ` + repoName + `/, not in ` + path.Base(path.Dir(pkg.PkgPath)) + `/`,
					Fix:      "move to " + repoName + "/, or rename if it's a consumer-defined interface",
					Severity: cfg.Sev,
				})
			}
			if info.AliasFrom != "" && isRepoPackageWith(m, info.AliasFrom) {
				violations = append(violations, Violation{
					File:     relFile,
					Line:     info.Pos.Line,
					Rule:     "structure.interface-placement",
					Message:  `type alias "` + info.Name + `" re-exports interface from ` + repoName + ` — suspected cross-domain dependency; use ` + m.OrchestrationDir + `/ instead`,
					Fix:      "remove alias and move cross-domain coordination to " + m.OrchestrationDir + "/",
					Severity: cfg.Sev,
				})
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
					Rule:     "testing.no-handmock",
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

func isTypePatternFile(m Model, dir, filename string) bool {
	name := strings.TrimSuffix(filename, ".go")
	for _, tp := range m.TypePatterns {
		if tp.Dir == dir && strings.HasPrefix(name, tp.FilePrefix+"_") {
			return true
		}
	}
	return false
}
