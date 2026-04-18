package rules

import (
	"go/ast"
	"go/token"
	"maps"

	"golang.org/x/tools/go/packages"
)

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
		maps.Copy(result, collectInterfacesFromFile(file, true))
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
