package analysisutil

import (
	"go/ast"
	"go/types"
	"strings"
)

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
