package rules

import (
	"fmt"
	"go/ast"
	"go/token"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"
)

// CheckContainerInterface detects interfaces that are used only as struct field
// types — never as function parameters or return types. This is a vibe-coding
// smell: the interface is being used as a value container rather than as an
// abstraction. A common cause is a wiring layer that needs to hold a value
// whose concrete type is not exposed (for example, when a domain's alias.go
// re-exports the constructor but not the type), so the developer declares a
// local interface just to give the field a type.
//
// Default severity is Warning. The rule is informational, not blocking, and is
// reported alongside other rules in RunAll without failing the test unless the
// caller upgrades severity via WithSeverity(Error).
//
// Skipped:
//   - Test files (_test.go) where mock/fake fixtures naturally use this shape.
//   - Type aliases (type Foo = pkg.Foo) — not new interface declarations.
//   - Embedded fields (anonymous embedding) in structs.
//   - Interfaces that are not used at all (different smell category).
func CheckContainerInterface(pkgs []*packages.Package, opts ...Option) []Violation {
	cfg := NewConfig(opts...)
	sev := cfg.Sev
	if !cfg.SeverityExplicit() {
		sev = Warning
	}

	var violations []Violation
	for _, pkg := range pkgs {
		violations = append(violations, checkContainerInterfacesInPkg(pkg, sev)...)
	}
	return violations
}

// checkContainerInterfacesInPkg evaluates a single package for container-only
// interface declarations.
func checkContainerInterfacesInPkg(pkg *packages.Package, sev Severity) []Violation {
	ifaces := collectNonTestInterfaces(pkg)
	if len(ifaces) == 0 {
		return nil
	}

	usage := make(map[string]*usageCounts, len(ifaces))
	for name := range ifaces {
		usage[name] = &usageCounts{}
	}

	for _, file := range pkg.Syntax {
		if isTestFile(pkg, file) {
			continue
		}
		ast.Inspect(file, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.StructType:
				if node.Fields == nil {
					return true
				}
				for _, f := range node.Fields.List {
					if len(f.Names) == 0 {
						// embedded field — composition, not container usage
						continue
					}
					countTypeRefs(f.Type, ifaces, usage, fieldUsage)
				}
			case *ast.FuncDecl:
				if node.Type == nil {
					return true
				}
				if node.Type.Params != nil {
					for _, f := range node.Type.Params.List {
						countTypeRefs(f.Type, ifaces, usage, paramUsage)
					}
				}
				if node.Type.Results != nil {
					for _, f := range node.Type.Results.List {
						countTypeRefs(f.Type, ifaces, usage, returnUsage)
					}
				}
			}
			return true
		})
	}

	var flagged []string
	for name, c := range usage {
		if c.field > 0 && c.param == 0 && c.ret == 0 {
			flagged = append(flagged, name)
		}
	}
	sort.Strings(flagged)

	var violations []Violation
	for _, name := range flagged {
		pos := ifaces[name]
		file := pos.Filename
		if rel := relativePathForPackage(pkg, file); rel != "" {
			file = rel
		}
		violations = append(violations, Violation{
			File:              file,
			Line:              pos.Line,
			Rule:              "interface.container-only",
			Message:           fmt.Sprintf("interface %q is only used as a struct field type, never as a function parameter or return type — likely a value container rather than a real abstraction", name),
			Fix:               "either use the concrete type if available (re-export it from alias.go if blocked by isolation rules), or pass this interface as a constructor parameter so it serves as a real abstraction",
			DefaultSeverity:   sev,
			EffectiveSeverity: sev,
		})
	}
	return violations
}

// usageCounts tracks how often a named interface is referenced in different
// syntactic positions within a single package.
type usageCounts struct {
	field int
	param int
	ret   int
}

type usageKind int

const (
	fieldUsage usageKind = iota
	paramUsage
	returnUsage
)

// countTypeRefs walks a type expression looking for identifiers that match
// known local interfaces and increments the appropriate usage counter.
// Recurses into composite types so `[]Foo`, `map[string]Foo`, `chan Foo`, and
// `func(Foo)` are detected, but does not cross into selector expressions
// (foreign types) or unrelated AST nodes.
func countTypeRefs(expr ast.Expr, ifaces map[string]token.Position, usage map[string]*usageCounts, kind usageKind) {
	switch e := expr.(type) {
	case *ast.Ident:
		if _, ok := ifaces[e.Name]; ok {
			c := usage[e.Name]
			switch kind {
			case fieldUsage:
				c.field++
			case paramUsage:
				c.param++
			case returnUsage:
				c.ret++
			}
		}
	case *ast.StarExpr:
		countTypeRefs(e.X, ifaces, usage, kind)
	case *ast.ArrayType:
		countTypeRefs(e.Elt, ifaces, usage, kind)
	case *ast.MapType:
		countTypeRefs(e.Key, ifaces, usage, kind)
		countTypeRefs(e.Value, ifaces, usage, kind)
	case *ast.ChanType:
		countTypeRefs(e.Value, ifaces, usage, kind)
	case *ast.FuncType:
		// nested function type used as a parameter or return: treat its inner
		// types as the same kind as the outer position
		if e.Params != nil {
			for _, f := range e.Params.List {
				countTypeRefs(f.Type, ifaces, usage, kind)
			}
		}
		if e.Results != nil {
			for _, f := range e.Results.List {
				countTypeRefs(f.Type, ifaces, usage, kind)
			}
		}
	}
}

// collectNonTestInterfaces returns named interface declarations from a package,
// excluding type aliases and excluding interfaces declared in _test.go files.
func collectNonTestInterfaces(pkg *packages.Package) map[string]token.Position {
	result := make(map[string]token.Position)
	for _, file := range pkg.Syntax {
		if isTestFile(pkg, file) {
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
				if ts.Assign != 0 {
					// type alias — not a new interface declaration
					continue
				}
				if _, isIface := ts.Type.(*ast.InterfaceType); !isIface {
					continue
				}
				result[ts.Name.Name] = pkg.Fset.Position(ts.Name.Pos())
			}
		}
	}
	return result
}

// isTestFile reports whether an AST file lives in a _test.go file.
func isTestFile(pkg *packages.Package, file *ast.File) bool {
	pos := pkg.Fset.Position(file.Pos())
	return strings.HasSuffix(pos.Filename, "_test.go")
}
