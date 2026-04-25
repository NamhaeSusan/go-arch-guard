package analysisutil

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"testing"
)

func TestReceiverTypeName(t *testing.T) {
	src := `package sample
type Foo struct{}
func (f Foo) A() {}
func (f *Foo) B() {}
type Bar[T any] struct{}
func (b Bar[T]) C() {}
func (b *Bar[T]) D() {}
type Baz[K comparable, V any] struct{}
func (z *Baz[K, V]) E() {}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "sample.go", src, parser.AllErrors)
	if err != nil {
		t.Fatal(err)
	}

	want := []string{"Foo", "Foo", "Bar", "Bar", "Baz"}
	var got []string
	for _, decl := range file.Decls {
		fd, ok := decl.(*ast.FuncDecl)
		if !ok || fd.Recv == nil {
			continue
		}
		got = append(got, ReceiverTypeName(fd.Recv.List[0].Type))
	}
	if len(got) != len(want) {
		t.Fatalf("got %d names, want %d: %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestSnakeToPascal(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"user_repository", "UserRepository"},
		{"order", "Order"},
		{"a_b_c", "ABC"},
		{"", ""},
		{"_leading", "Leading"},
		{"trailing_", "Trailing"},
		{"double__underscore", "DoubleUnderscore"},
	}
	for _, tc := range cases {
		if got := SnakeToPascal(tc.in); got != tc.want {
			t.Fatalf("SnakeToPascal(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestResolveIdentImportPath(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "sample.go", `package sample
import repo "example.com/app/internal/repo"
type Alias = repo.Store
`, parser.AllErrors)
	if err != nil {
		t.Fatal(err)
	}

	if got := ResolveIdentImportPath(file, "repo"); got != "example.com/app/internal/repo" {
		t.Fatalf("ResolveIdentImportPath() = %q", got)
	}
	if got := ResolveIdentImportPath(file, "missing"); got != "" {
		t.Fatalf("ResolveIdentImportPath(missing) = %q, want empty", got)
	}
}

func TestStripWrappersAndNamedQualifiedName(t *testing.T) {
	pkg := types.NewPackage("example.com/app/internal/repo", "repo")
	obj := types.NewTypeName(token.NoPos, pkg, "Store", nil)
	named := types.NewNamed(obj, types.NewStruct(nil, nil), nil)
	wrapped := types.NewSlice(types.NewPointer(named))

	got := NamedQualifiedName(StripWrappers(wrapped))
	want := "example.com/app/internal/repo.Store"
	if got != want {
		t.Fatalf("NamedQualifiedName(StripWrappers()) = %q, want %q", got, want)
	}
}

func TestWalkFuncSignatureTypes(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "sample.go", `package sample
type Item struct{}
func Process(items []Item, err error) (*Item, error) { return nil, nil }
`, parser.AllErrors)
	if err != nil {
		t.Fatal(err)
	}

	info := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Defs:  make(map[*ast.Ident]types.Object),
		Uses:  make(map[*ast.Ident]types.Object),
	}
	conf := types.Config{Importer: importer.Default()}
	pkg, err := conf.Check("example.com/app/internal/sample", fset, []*ast.File{file}, info)
	if err != nil {
		t.Fatal(err)
	}
	_ = pkg

	var got []string
	WalkFuncSignatureTypes(info, file, func(_ *ast.FuncDecl, _ *ast.Field, t types.Type) {
		got = append(got, types.TypeString(t, func(p *types.Package) string { return p.Path() }))
	})

	want := []string{
		"[]example.com/app/internal/sample.Item",
		"error",
		"*example.com/app/internal/sample.Item",
		"error",
	}
	if len(got) != len(want) {
		t.Fatalf("WalkFuncSignatureTypes() len = %d, want %d (%v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("WalkFuncSignatureTypes()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestResolveCalleeID(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "sample.go", `package sample
type Service struct{}
func Top() {}
func (Service) Value() {}
func (*Service) Pointer() {}
func Use(s Service) {
	Top()
	s.Value()
	s.Pointer()
}
`, parser.AllErrors)
	if err != nil {
		t.Fatal(err)
	}

	info := &types.Info{
		Types:      make(map[ast.Expr]types.TypeAndValue),
		Defs:       make(map[*ast.Ident]types.Object),
		Uses:       make(map[*ast.Ident]types.Object),
		Selections: make(map[*ast.SelectorExpr]*types.Selection),
	}
	conf := types.Config{Importer: importer.Default()}
	if _, err := conf.Check("example.com/app/internal/sample", fset, []*ast.File{file}, info); err != nil {
		t.Fatal(err)
	}

	var got []string
	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		got = append(got, ResolveCalleeID(info, call))
		return true
	})

	want := []string{
		"example.com/app/internal/sample.Top",
		"example.com/app/internal/sample.Service.Value",
		"example.com/app/internal/sample.(*Service).Pointer",
	}
	if len(got) != len(want) {
		t.Fatalf("ResolveCalleeID() len = %d, want %d (%v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("ResolveCalleeID()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
