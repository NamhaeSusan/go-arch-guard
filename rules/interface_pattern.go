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

		// checkExportedImpl is scope-driven (covers alias chains) and is
		// independent of AST-collected interfaces.
		violations = append(violations, checkExportedImpl(pkg, cfg)...)

		ifaces := collectExportedInterfacesFromPkg(pkg)
		if len(ifaces) == 0 {
			continue
		}

		violations = append(violations, checkSingleInterfacePerPackage(pkg, ifaces, cfg)...)
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
		File:              relativePackageFile(pkg),
		Rule:              "interface.single-per-package",
		Message:           fmt.Sprintf("package has %d exported interfaces (%s), expected at most 1", len(ifaces), strings.Join(names, ", ")),
		Fix:               "split into separate packages, one interface each",
		DefaultSeverity:   cfg.Sev,
		EffectiveSeverity: cfg.Sev,
	}}
}

// checkExportedImpl detects exported struct types that implement exported
// interfaces. Uses go/types.Implements; silently skips pairs where either
// type object cannot be recovered (e.g. IllTyped packages with missing
// type info). This is intentional — name-only comparison was the bug we
// were fixing in #17, and running an architecture rule on broken code
// produces low-signal noise regardless.
//
// The set of interfaces considered is scope-driven: any exported name in
// pkg.Types.Scope() whose unaliased underlying type is *types.Interface.
// This covers direct definitions, direct aliases, and alias chains.
func checkExportedImpl(pkg *packages.Package, cfg Config) []Violation {
	if pkg.Types == nil {
		return nil
	}
	structs := collectExportedStructs(pkg)
	if len(structs) == 0 {
		return nil
	}
	scope := pkg.Types.Scope()

	// Collect exported interfaces from scope (covers alias chains too).
	typedIfaces := make(map[string]*types.Interface)
	for _, name := range scope.Names() {
		if !ast.IsExported(name) {
			continue
		}
		if iface := lookupInterface(scope, name); iface != nil && iface.NumMethods() > 0 {
			typedIfaces[name] = iface
		}
	}
	if len(typedIfaces) == 0 {
		return nil
	}

	var violations []Violation
	for structName := range structs {
		obj := scope.Lookup(structName)
		if obj == nil {
			continue
		}
		named, ok := types.Unalias(obj.Type()).(*types.Named)
		if !ok {
			continue
		}
		ptrType := types.NewPointer(named)

		for ifaceName, iface := range typedIfaces {
			if types.Implements(named, iface) || types.Implements(ptrType, iface) {
				violations = append(violations, Violation{
					File:              relativePackageFile(pkg),
					Rule:              "interface.exported-impl",
					Message:           fmt.Sprintf("type %q is exported but implements interface %q; make it unexported", structName, ifaceName),
					Fix:               fmt.Sprintf("rename to %q", strings.ToLower(structName[:1])+structName[1:]),
					DefaultSeverity:   cfg.Sev,
					EffectiveSeverity: cfg.Sev,
				})
			}
		}
	}
	return violations
}

// lookupInterface resolves name in scope to a *types.Interface, unwrapping
// type aliases (including alias chains). Returns nil when name is not an
// interface type.
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
					File:              relativePackageFile(pkg),
					Rule:              "interface.constructor-name",
					Message:           fmt.Sprintf("constructor %q must be named \"New\"; NewXxx variants are not allowed", name),
					Fix:               "rename to \"New\"",
					DefaultSeverity:   cfg.Sev,
					EffectiveSeverity: cfg.Sev,
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
				File:              relativePackageFile(pkg),
				Rule:              "interface.constructor-returns-interface",
				Message:           fmt.Sprintf("New() returns %s, should return an interface", returnStr),
				Fix:               fix,
				DefaultSeverity:   cfg.Sev,
				EffectiveSeverity: cfg.Sev,
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
