package rules

import (
	"fmt"
	"go/ast"
	"go/types"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"
)

// CheckInterfacePattern enforces interface design conventions:
//   - exported structs should not implement exported interfaces in the same package
//   - constructors must be named exactly "New"
//   - New() must return an exported interface
func CheckInterfacePattern(pkgs []*packages.Package, opts ...Option) []Violation {
	cfg := NewConfig(opts...)
	m := cfg.model()

	var violations []Violation
	for _, pkg := range pkgs {
		if isExcludedInterfacePatternPkg(m, pkg) {
			continue
		}

		ifaces := collectExportedInterfacesFromPkg(pkg)
		if len(ifaces) == 0 {
			continue
		}

		violations = append(violations, checkSingleInterfacePerPackage(pkg, ifaces, cfg)...)
		violations = append(violations, checkExportedImpl(pkg, ifaces, cfg)...)
		violations = append(violations, checkConstructorName(pkg, cfg)...)
		violations = append(violations, checkConstructorReturnsInterface(pkg, ifaces, cfg)...)
	}
	return violations
}

// isExcludedInterfacePatternPkg determines if a package should be skipped for
// interface pattern checks based on the model's InterfacePatternExclude map.
func isExcludedInterfacePatternPkg(m Model, pkg *packages.Package) bool {
	parts := strings.Split(pkg.PkgPath, "/")
	internalIdx := -1
	for i, p := range parts {
		if p == "internal" {
			internalIdx = i
			break
		}
	}
	if internalIdx < 0 || internalIdx >= len(parts)-1 {
		return true
	}

	after := parts[internalIdx+1:]

	// Always exclude SharedDir (e.g. "pkg")
	if m.SharedDir != "" && after[0] == m.SharedDir {
		return true
	}

	if m.DomainDir == "" {
		// Flat layout: check first segment after internal/
		return m.InterfacePatternExclude[after[0]]
	}

	// Domain layout: skip domain dir, then check sublayer
	if after[0] != m.DomainDir || len(after) < 3 {
		return true
	}
	// after = [domain, <domainName>, <sublayer>, ...]
	// Check both single segment (e.g. "handler") and nested (e.g. "core/repo")
	sublayer := after[2]
	if m.InterfacePatternExclude[sublayer] {
		return true
	}
	if len(after) >= 4 {
		nested := after[2] + "/" + after[3]
		if m.InterfacePatternExclude[nested] {
			return true
		}
	}
	return false
}

// checkSingleInterfacePerPackage warns when a package declares more than one
// exported interface.
func checkSingleInterfacePerPackage(pkg *packages.Package, ifaces map[string]*ast.InterfaceType, cfg Config) []Violation {
	if len(ifaces) <= 1 {
		return nil
	}
	var names []string
	for name := range ifaces {
		names = append(names, name)
	}
	sort.Strings(names)
	return []Violation{{
		File:     relativePackageFile(pkg),
		Rule:     "interface.single-per-package",
		Message:  fmt.Sprintf("package has %d exported interfaces (%s), expected at most 1", len(ifaces), strings.Join(names, ", ")),
		Fix:      "split into separate packages, one interface each",
		Severity: cfg.Sev,
	}}
}

// checkExportedImpl detects exported structs that implement an exported interface
// in the same package — the impl should be unexported.
//
// For fully-typed packages it uses go/types.Implements so signatures are
// compared properly. When the package is ill-typed (pkg.Types nil, or a
// specific interface/struct type object cannot be recovered from scope), it
// falls back to AST name-only matching. The fallback preserves the previous
// behavior's false-positive surface — a looser but strictly better outcome
// than silently dropping diagnostics during partially broken refactors.
func checkExportedImpl(pkg *packages.Package, ifaces map[string]*ast.InterfaceType, cfg Config) []Violation {
	structs := collectExportedStructs(pkg)
	if len(structs) == 0 || len(ifaces) == 0 {
		return nil
	}

	var scope *types.Scope
	if pkg.Types != nil {
		scope = pkg.Types.Scope()
	}

	// Resolve each interface to a *types.Interface when possible (treats
	// type aliases via types.Unalias). Interfaces that cannot be resolved are
	// left for the AST fallback.
	typedIfaces := make(map[string]*types.Interface, len(ifaces))
	if scope != nil {
		for name := range ifaces {
			if iface := lookupInterface(scope, name); iface != nil && iface.NumMethods() > 0 {
				typedIfaces[name] = iface
			}
		}
	}

	var violations []Violation
	for structName := range structs {
		var named *types.Named
		if scope != nil {
			if obj := scope.Lookup(structName); obj != nil {
				named, _ = types.Unalias(obj.Type()).(*types.Named)
			}
		}

		for ifaceName, astIface := range ifaces {
			matched := false
			iface, ifaceOK := typedIfaces[ifaceName]
			typedPathAvailable := named != nil && ifaceOK
			if typedPathAvailable {
				// Fully-typed path: signature-aware via types.Implements.
				ptrType := types.NewPointer(named)
				matched = types.Implements(named, iface) || types.Implements(ptrType, iface)
			}
			// Fallback to AST name-only match when:
			//   - the type objects could not be recovered, or
			//   - the package is ill-typed and the typed check said "no match"
			//     (signatures may be unreliable when types are incomplete).
			if !matched && (!typedPathAvailable || pkg.IllTyped) {
				matched = structMatchesInterfaceByName(pkg, structName, astIface)
			}
			if matched {
				violations = append(violations, Violation{
					File:     relativePackageFile(pkg),
					Rule:     "interface.exported-impl",
					Message:  fmt.Sprintf("type %q is exported but implements interface %q; make it unexported", structName, ifaceName),
					Fix:      fmt.Sprintf("rename to %q", strings.ToLower(structName[:1])+structName[1:]),
					Severity: cfg.Sev,
				})
			}
		}
	}
	return violations
}

// lookupInterface resolves name in scope to a *types.Interface, unwrapping
// type aliases. Returns nil when name is not an interface type.
func lookupInterface(scope *types.Scope, name string) *types.Interface {
	obj := scope.Lookup(name)
	if obj == nil {
		return nil
	}
	t := types.Unalias(obj.Type())
	if iface, ok := t.Underlying().(*types.Interface); ok {
		return iface
	}
	return nil
}

// structMatchesInterfaceByName returns true when every method name declared on
// the interface has a corresponding method declared on structName in pkg's AST.
// Signatures are not compared — used only as an ill-typed fallback.
func structMatchesInterfaceByName(pkg *packages.Package, structName string, iface *ast.InterfaceType) bool {
	if iface == nil || iface.Methods == nil || len(iface.Methods.List) == 0 {
		return false
	}
	methods := collectMethods(pkg)
	seen := false
	for _, m := range iface.Methods.List {
		for _, name := range m.Names {
			seen = true
			if !methods[structName+"."+name.Name] {
				return false
			}
		}
	}
	return seen
}

// collectExportedStructs returns names of all exported struct types in a package.
func collectExportedStructs(pkg *packages.Package) map[string]bool {
	result := make(map[string]bool)
	for _, file := range pkg.Syntax {
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
				if _, ok := ts.Type.(*ast.StructType); ok {
					result[ts.Name.Name] = true
				}
			}
		}
	}
	return result
}

// checkConstructorName flags any exported function starting with "New" that is
// not exactly "New". Methods (with receiver) are skipped.
func checkConstructorName(pkg *packages.Package, cfg Config) []Violation {
	var violations []Violation
	for _, file := range pkg.Syntax {
		for _, decl := range file.Decls {
			fd, ok := decl.(*ast.FuncDecl)
			if !ok || fd.Recv != nil {
				continue
			}
			name := fd.Name.Name
			if !fd.Name.IsExported() {
				continue
			}
			if strings.HasPrefix(name, "New") && name != "New" {
				violations = append(violations, Violation{
					File:     relativePackageFile(pkg),
					Rule:     "interface.constructor-name",
					Message:  fmt.Sprintf("constructor %q must be named \"New\"; NewXxx variants are not allowed", name),
					Fix:      "rename to \"New\"",
					Severity: cfg.Sev,
				})
			}
		}
	}
	return violations
}

// checkConstructorReturnsInterface ensures a function named exactly "New" returns
// an exported interface from the same package as its first return type.
func checkConstructorReturnsInterface(pkg *packages.Package, ifaces map[string]*ast.InterfaceType, cfg Config) []Violation {
	var violations []Violation
	for _, file := range pkg.Syntax {
		for _, decl := range file.Decls {
			fd, ok := decl.(*ast.FuncDecl)
			if !ok || fd.Recv != nil || fd.Name.Name != "New" {
				continue
			}
			if fd.Type.Results == nil || len(fd.Type.Results.List) == 0 {
				continue
			}

			firstRet := fd.Type.Results.List[0].Type
			returnStr := formatTypeExpr(firstRet)

			// Check if the first return type is an exported interface in this package.
			if ident, ok := firstRet.(*ast.Ident); ok && ifaces[ident.Name] != nil {
				continue // valid — returns interface
			}

			fix := "return an interface type"
			if len(ifaces) == 1 {
				for ifaceName := range ifaces {
					fix = fmt.Sprintf("return %s instead", ifaceName)
				}
			}

			violations = append(violations, Violation{
				File:     relativePackageFile(pkg),
				Rule:     "interface.constructor-returns-interface",
				Message:  fmt.Sprintf("New() returns %s, should return an interface", returnStr),
				Fix:      fix,
				Severity: cfg.Sev,
			})
		}
	}
	return violations
}

// formatTypeExpr returns a string representation of a type expression.
func formatTypeExpr(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.StarExpr:
		return "*" + formatTypeExpr(e.X)
	case *ast.SelectorExpr:
		return formatTypeExpr(e.X) + "." + e.Sel.Name
	case *ast.ArrayType:
		return "[]" + formatTypeExpr(e.Elt)
	case *ast.MapType:
		return "map[" + formatTypeExpr(e.Key) + "]" + formatTypeExpr(e.Value)
	default:
		return "unknown"
	}
}
