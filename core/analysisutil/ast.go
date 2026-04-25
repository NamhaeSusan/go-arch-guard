package analysisutil

import (
	"go/ast"
	"go/token"
	"go/types"
	"strings"
)

// TypeSpecInfo describes a single top-level type declaration in a Go file —
// what InspectTypeSpecs extracts. Callers compute file-level metadata (e.g.
// a relative path) once outside the loop; only per-decl fields live here.
type TypeSpecInfo struct {
	Name        string
	Line        int
	IsInterface bool
	AliasFrom   string // import path of `type X = pkg.Y`; empty if not an alias
}

// InspectTypeSpecs walks the top-level type decls in file and returns one
// entry per interface declaration or per type alias whose RHS is
// `<ident>.<Name>`. Other type decls are skipped. Callers usually want to
// detect interfaces or detect re-exports of types from other packages.
func InspectTypeSpecs(file *ast.File, fset *token.FileSet) []TypeSpecInfo {
	var result []TypeSpecInfo
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
			info := TypeSpecInfo{
				Name: ts.Name.Name,
				Line: fset.Position(ts.Name.Pos()).Line,
			}
			if _, ok := ts.Type.(*ast.InterfaceType); ok {
				info.IsInterface = true
			}
			if ts.Assign != 0 {
				if sel, ok := ts.Type.(*ast.SelectorExpr); ok {
					if ident, ok := sel.X.(*ast.Ident); ok {
						info.AliasFrom = ResolveIdentImportPath(file, ident.Name)
					}
				}
			}
			if info.IsInterface || info.AliasFrom != "" {
				result = append(result, info)
			}
		}
	}
	return result
}

// ReceiverTypeName extracts the unqualified receiver type name from an
// *ast.FuncDecl receiver expression. It unwraps a leading pointer and
// handles generic receivers — *ast.IndexExpr (T[U]) and *ast.IndexListExpr
// (T[U, V]) — so callers do not silently miss methods on parameterized
// types. Returns "" if the expression does not match a recognized shape.
func ReceiverTypeName(expr ast.Expr) string {
	if star, ok := expr.(*ast.StarExpr); ok {
		expr = star.X
	}
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.IndexExpr:
		if id, ok := t.X.(*ast.Ident); ok {
			return id.Name
		}
	case *ast.IndexListExpr:
		if id, ok := t.X.(*ast.Ident); ok {
			return id.Name
		}
	}
	return ""
}

// SnakeToPascal converts a snake_case identifier (e.g. "user_repository")
// to PascalCase (e.g. "UserRepository"). Empty segments are skipped, so
// leading/trailing/double underscores do not produce empty letters. Input
// is assumed to be ASCII; non-ASCII first bytes in a segment are not
// case-converted but are preserved.
func SnakeToPascal(s string) string {
	parts := strings.Split(s, "_")
	var b strings.Builder
	for _, part := range parts {
		if part == "" {
			continue
		}
		b.WriteString(strings.ToUpper(part[:1]))
		b.WriteString(part[1:])
	}
	return b.String()
}

func ResolveIdentImportPath(file *ast.File, identName string) string {
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

func WalkFuncSignatureTypes(info *types.Info, file *ast.File, visit func(*ast.FuncDecl, *ast.Field, types.Type)) {
	if info == nil || file == nil || visit == nil {
		return
	}
	for _, decl := range file.Decls {
		fd, ok := decl.(*ast.FuncDecl)
		if !ok || fd.Type == nil {
			continue
		}
		walkFieldListTypes(info, fd, fd.Type.Params, visit)
		walkFieldListTypes(info, fd, fd.Type.Results, visit)
	}
}

func StripWrappers(t types.Type) types.Type {
	for {
		switch x := t.(type) {
		case *types.Pointer:
			t = x.Elem()
		case *types.Slice:
			t = x.Elem()
		case *types.Array:
			t = x.Elem()
		case *types.Map:
			t = x.Elem()
		case *types.Chan:
			t = x.Elem()
		default:
			return t
		}
	}
}

func NamedQualifiedName(t types.Type) string {
	named, ok := t.(*types.Named)
	if !ok {
		return ""
	}
	obj := named.Obj()
	if obj == nil || obj.Pkg() == nil {
		return ""
	}
	return obj.Pkg().Path() + "." + obj.Name()
}

func ResolveCalleeID(info *types.Info, call *ast.CallExpr) string {
	if info == nil || call == nil {
		return ""
	}
	switch fun := call.Fun.(type) {
	case *ast.SelectorExpr:
		if sel, ok := info.Selections[fun]; ok && sel != nil {
			if fn, ok := sel.Obj().(*types.Func); ok {
				return FuncQualifiedName(fn)
			}
		}
		if obj := info.Uses[fun.Sel]; obj != nil {
			if fn, ok := obj.(*types.Func); ok {
				return FuncQualifiedName(fn)
			}
		}
	case *ast.Ident:
		if obj := info.Uses[fun]; obj != nil {
			if fn, ok := obj.(*types.Func); ok {
				return FuncQualifiedName(fn)
			}
		}
	}
	return ""
}

func FuncQualifiedName(fn *types.Func) string {
	sig, ok := fn.Type().(*types.Signature)
	if !ok {
		return ""
	}
	pkg := fn.Pkg()
	if pkg == nil {
		return ""
	}
	if sig.Recv() == nil {
		return pkg.Path() + "." + fn.Name()
	}
	recv := sig.Recv().Type()
	if ptr, ok := recv.(*types.Pointer); ok {
		if named, ok := ptr.Elem().(*types.Named); ok {
			return pkg.Path() + ".(*" + named.Obj().Name() + ")." + fn.Name()
		}
	}
	if named, ok := recv.(*types.Named); ok {
		return pkg.Path() + "." + named.Obj().Name() + "." + fn.Name()
	}
	return ""
}

func walkFieldListTypes(info *types.Info, fd *ast.FuncDecl, fields *ast.FieldList, visit func(*ast.FuncDecl, *ast.Field, types.Type)) {
	if fields == nil {
		return
	}
	for _, field := range fields.List {
		t := info.TypeOf(field.Type)
		if t == nil {
			continue
		}
		visit(fd, field, t)
	}
}
