package rules

import (
	"fmt"
	"go/ast"
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

		ifaces := collectExportedInterfaces(pkg)
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

	// Always exclude pkg/ (SharedDir)
	if after[0] == "pkg" {
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
	sublayer := after[2]
	return m.InterfacePatternExclude[sublayer]
}

// collectExportedInterfaces returns all exported interface types in a package.
func collectExportedInterfaces(pkg *packages.Package) map[string]*ast.InterfaceType {
	result := make(map[string]*ast.InterfaceType)
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
				if iface, ok := ts.Type.(*ast.InterfaceType); ok {
					result[ts.Name.Name] = iface
				}
			}
		}
	}
	return result
}

// interfaceMethodNames extracts method names from an interface type.
func interfaceMethodNames(iface *ast.InterfaceType) map[string]bool {
	methods := make(map[string]bool)
	if iface.Methods == nil {
		return methods
	}
	for _, m := range iface.Methods.List {
		for _, name := range m.Names {
			methods[name.Name] = true
		}
	}
	return methods
}

// checkSingleInterfacePerPackage warns when a package declares more than one
// exported interface. Always uses Warning severity.
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
		Severity: Warning,
	}}
}

// checkExportedImpl detects exported structs that implement an exported interface
// in the same package — the impl should be unexported.
func checkExportedImpl(pkg *packages.Package, ifaces map[string]*ast.InterfaceType, cfg Config) []Violation {
	methods := collectMethods(pkg)
	structs := collectExportedStructs(pkg)

	var violations []Violation
	for structName := range structs {
		for ifaceName, iface := range ifaces {
			ifaceMethods := interfaceMethodNames(iface)
			if len(ifaceMethods) == 0 {
				continue
			}
			allMatch := true
			for mName := range ifaceMethods {
				if !methods[structName+"."+mName] {
					allMatch = false
					break
				}
			}
			if allMatch {
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
