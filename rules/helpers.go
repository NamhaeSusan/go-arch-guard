package rules

import (
	"fmt"
	"go/ast"
	"go/token"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
)

func findImportPosition(pkg *packages.Package, importPath, projectRoot string) (string, int) {
	absRoot, _ := filepath.Abs(projectRoot)
	fset := pkg.Fset
	for _, file := range pkg.Syntax {
		for _, imp := range file.Imports {
			path := strings.Trim(imp.Path.Value, `"`)
			if path == importPath {
				pos := fset.Position(imp.Pos())
				rel, err := filepath.Rel(absRoot, pos.Filename)
				if err != nil {
					return pos.Filename, pos.Line
				}
				return filepath.ToSlash(rel), pos.Line
			}
		}
	}
	if len(pkg.GoFiles) > 0 {
		rel, err := filepath.Rel(absRoot, pkg.GoFiles[0])
		if err != nil {
			return pkg.GoFiles[0], 0
		}
		return filepath.ToSlash(rel), 0
	}
	return pkg.PkgPath, 0
}

func relativePathForPackage(pkg *packages.Package, path string) string {
	if pkg != nil && pkg.Module != nil && pkg.Module.Dir != "" {
		rel, err := filepath.Rel(pkg.Module.Dir, path)
		if err == nil {
			return filepath.ToSlash(rel)
		}
	}
	return filepath.ToSlash(path)
}

func resolveModule(pkgs []*packages.Package, explicit string) string {
	if explicit != "" {
		return explicit
	}
	for _, pkg := range pkgs {
		if pkg.Module != nil && pkg.Module.Path != "" {
			return pkg.Module.Path
		}
	}
	return ""
}

func resolveRoot(pkgs []*packages.Package, explicit string) string {
	if explicit != "" {
		return explicit
	}
	for _, pkg := range pkgs {
		if pkg.Module != nil && pkg.Module.Dir != "" {
			return pkg.Module.Dir
		}
	}
	return ""
}

func validateModule(pkgs []*packages.Package, projectModule string) []Violation {
	if projectModule == "" {
		return []Violation{{
			Rule:     "meta.no-matching-packages",
			Message:  "project module could not be determined — all import checks will be skipped",
			Fix:      "pass a non-empty module path or ensure packages are loaded with NeedModule",
			Severity: Warning,
		}}
	}
	prefix := projectModule + "/"
	for _, pkg := range pkgs {
		if pkg.PkgPath == projectModule || strings.HasPrefix(pkg.PkgPath, prefix) {
			return nil
		}
	}
	return []Violation{{
		Rule:     "meta.no-matching-packages",
		Message:  fmt.Sprintf("module %q does not match any loaded package — all import checks will be skipped", projectModule),
		Fix:      "verify the module argument matches go.mod (e.g. pass the value from pkgs[0].Module.Path)",
		Severity: Warning,
	}}
}

// resolveIdentImportPath returns the import path that identName refers to
// by scanning the file's import declarations. Returns "" if not found.
func resolveIdentImportPath(file *ast.File, identName string) string {
	for _, imp := range file.Imports {
		impPath := strings.Trim(imp.Path.Value, `"`)
		alias := ""
		if imp.Name != nil {
			alias = imp.Name.Name
		} else {
			parts := strings.Split(impPath, "/")
			alias = parts[len(parts)-1]
		}
		if alias == identName {
			return impPath
		}
	}
	return ""
}

func isExcludedPackage(cfg Config, pkgPath, projectModule string) bool {
	return cfg.IsExcluded(projectRelativePackagePath(pkgPath, projectModule))
}

func projectRelativePackagePath(pkgPath, projectModule string) string {
	if pkgPath == "" || projectModule == "" {
		return ""
	}
	if pkgPath == projectModule {
		return "."
	}
	if rel, ok := strings.CutPrefix(pkgPath, projectModule+"/"); ok {
		return rel
	}
	return ""
}

// isPortSublayer reports whether the sublayer name is a port/contract layer
// (pure interface definitions like repo, gateway).
func isPortSublayer(name string) bool {
	base := name
	if i := strings.LastIndex(name, "/"); i >= 0 {
		base = name[i+1:]
	}
	return base == "repo" || base == "gateway"
}

// isContractSublayer reports whether the sublayer name is a contract layer
// (port/repo + service interfaces like svc).
func isContractSublayer(name string) bool {
	if isPortSublayer(name) {
		return true
	}
	base := name
	if i := strings.LastIndex(name, "/"); i >= 0 {
		base = name[i+1:]
	}
	return base == "svc"
}

// hasPortSublayer reports whether the model has any port sublayer.
func hasPortSublayer(m Model) bool {
	for _, sl := range m.Sublayers {
		if isPortSublayer(sl) {
			return true
		}
	}
	return false
}

// matchPortSublayer returns the port sublayer name if pkgPath references one, "" otherwise.
func matchPortSublayer(m Model, pkgPath string) string {
	for _, sl := range m.Sublayers {
		if !isPortSublayer(sl) {
			continue
		}
		if strings.HasSuffix(pkgPath, "/"+sl) || strings.Contains(pkgPath, "/"+sl+"/") {
			return sl
		}
	}
	return ""
}

// matchContractSublayer returns the contract sublayer name if pkgPath references one, "" otherwise.
func matchContractSublayer(m Model, pkgPath string) string {
	for _, sl := range m.Sublayers {
		if !isContractSublayer(sl) {
			continue
		}
		if strings.HasSuffix(pkgPath, "/"+sl) || strings.Contains(pkgPath, "/"+sl+"/") {
			return sl
		}
	}
	return ""
}

// portSublayerName returns the first port sublayer name from the model, or "core/repo" as fallback.
func portSublayerName(m Model) string {
	for _, sl := range m.Sublayers {
		if isPortSublayer(sl) {
			return sl
		}
	}
	return "core/repo"
}

// typeSpecInfo holds information about a type spec found during AST inspection.
type typeSpecInfo struct {
	Name      string
	Pos       token.Position
	IsIface   bool   // direct interface definition
	AliasFrom string // non-empty if type alias re-exports from this import path
}

// inspectTypeSpecs walks a file's type specs and returns info about interfaces
// and type aliases that re-export from other packages.
func inspectTypeSpecs(file *ast.File, fset *token.FileSet) []typeSpecInfo {
	var result []typeSpecInfo
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
			info := typeSpecInfo{
				Name: ts.Name.Name,
				Pos:  fset.Position(ts.Name.Pos()),
			}
			if _, isIface := ts.Type.(*ast.InterfaceType); isIface {
				info.IsIface = true
			}
			if ts.Assign != 0 {
				if sel, ok := ts.Type.(*ast.SelectorExpr); ok {
					if ident, ok := sel.X.(*ast.Ident); ok {
						info.AliasFrom = resolveIdentImportPath(file, ident.Name)
					}
				}
			}
			if info.IsIface || info.AliasFrom != "" {
				result = append(result, info)
			}
		}
	}
	return result
}

// collectInterfacesFromFile returns interface types from a single AST file.
// If exportedOnly is true, only exported interfaces are returned.
func collectInterfacesFromFile(file *ast.File, exportedOnly bool) map[string]*ast.InterfaceType {
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
			if exportedOnly && !ts.Name.IsExported() {
				continue
			}
			if iface, ok := ts.Type.(*ast.InterfaceType); ok {
				result[ts.Name.Name] = iface
			}
		}
	}
	return result
}

// collectExportedInterfacesFromPkg returns all exported interfaces across all files in a package.
func collectExportedInterfacesFromPkg(pkg *packages.Package) map[string]*ast.InterfaceType {
	result := make(map[string]*ast.InterfaceType)
	for _, file := range pkg.Syntax {
		for name, iface := range collectInterfacesFromFile(file, true) {
			result[name] = iface
		}
	}
	return result
}

// receiverTypeName extracts the type name from a method receiver expression.
func receiverTypeName(expr ast.Expr) string {
	if star, ok := expr.(*ast.StarExpr); ok {
		expr = star.X
	}
	if ident, ok := expr.(*ast.Ident); ok {
		return ident.Name
	}
	return ""
}

// collectMethods returns a set of "TypeName.MethodName" entries for all methods in the package.
func collectMethods(pkg *packages.Package) map[string]bool {
	result := make(map[string]bool)
	for _, file := range pkg.Syntax {
		for _, decl := range file.Decls {
			fd, ok := decl.(*ast.FuncDecl)
			if !ok || fd.Recv == nil || len(fd.Recv.List) == 0 {
				continue
			}
			typeName := receiverTypeName(fd.Recv.List[0].Type)
			if typeName != "" {
				result[typeName+"."+fd.Name.Name] = true
			}
		}
	}
	return result
}

func deduplicateMetaViolations(violations []Violation) []Violation {
	seen := make(map[string]bool)
	result := violations[:0]
	for _, v := range violations {
		if strings.HasPrefix(v.Rule, "meta.") {
			if seen[v.Rule] {
				continue
			}
			seen[v.Rule] = true
		}
		result = append(result, v)
	}
	return result
}
