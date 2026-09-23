package analysisutil

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"testing"
)

func TestExplicitCallsAndNestedSignatureTypes(t *testing.T) {
	fs := token.NewFileSet()
	f, err := parser.ParseFile(fs, "audit.go", `package p
 type Tx struct{}
 type Alias = Tx
 func Start[T any](v T) {}
 func Plain() {}
 func Use(a Alias, b map[*Tx]string, c func() *Tx) { Start(1); Start[int](1); (Plain)(); Plain() }
 `, 0)
	if err != nil {
		t.Fatal(err)
	}
	info := &types.Info{Types: map[ast.Expr]types.TypeAndValue{}, Uses: map[*ast.Ident]types.Object{}, Defs: map[*ast.Ident]types.Object{}}
	if _, err := (&types.Config{}).Check("example.com/p", fs, []*ast.File{f}, info); err != nil {
		t.Fatal(err)
	}
	calls := 0
	ast.Inspect(f, func(n ast.Node) bool {
		if c, ok := n.(*ast.CallExpr); ok {
			calls++
			if ResolveCalleeID(info, c) == "" {
				t.Errorf("unresolved call at %s", fs.Position(c.Pos()))
			}
		}
		return true
	})
	if calls != 4 {
		t.Fatalf("calls=%d", calls)
	}
	fields := 0
	WalkFuncSignatureTypes(info, f, func(fd *ast.FuncDecl, _ *ast.Field, typ types.Type) {
		if fd.Name.Name != "Use" {
			return
		}
		fields++
		found := false
		WalkSignatureNamedTypes(typ, func(id string) {
			if id == "example.com/p.Tx" {
				found = true
			}
		})
		if !found {
			t.Errorf("missed Tx in %s", typ)
		}
	})
	if fields != 3 {
		t.Fatalf("fields=%d", fields)
	}
}

func TestInspectGenericRawImportAlias(t *testing.T) {
	fs := token.NewFileSet()
	f, err := parser.ParseFile(fs, "audit.go", "package p; import `example.com/repo`; type X = repo.Store[int]", 0)
	if err != nil {
		t.Fatal(err)
	}
	got := InspectTypeSpecs(f, fs)
	if len(got) != 1 || got[0].AliasFrom != "example.com/repo" {
		t.Fatalf("got %+v", got)
	}
}

func TestInspectTypedImportPackageName(t *testing.T) {
	fs := token.NewFileSet()
	f, err := parser.ParseFile(fs, "audit.go", `package p; import "example.com/repo/v2"; type X = repo.Store`, 0)
	if err != nil {
		t.Fatal(err)
	}
	info := &types.Info{Uses: map[*ast.Ident]types.Object{}}
	ast.Inspect(f, func(n ast.Node) bool {
		if id, ok := n.(*ast.Ident); ok && id.Name == "repo" {
			info.Uses[id] = types.NewPkgName(id.Pos(), nil, "repo", types.NewPackage("example.com/repo/v2", "repo"))
		}
		return true
	})
	got := InspectTypeSpecs(f, fs, info)
	if len(got) != 1 || got[0].AliasFrom != "example.com/repo/v2" {
		t.Fatalf("got %+v", got)
	}
}
