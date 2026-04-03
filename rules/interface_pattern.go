package rules

import (
	"fmt"
	"go/ast"
	"strings"

	"golang.org/x/tools/go/packages"
)

// CheckInterfacePattern enforces interface design conventions:
//   - exported structs should not implement exported interfaces in the same package
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

		violations = append(violations, checkExportedImpl(pkg, ifaces, cfg)...)
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
