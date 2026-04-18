package rules

import (
	"fmt"
	"go/ast"
	"strings"

	"golang.org/x/tools/go/packages"
)

// CheckCrossDomainAnonymous detects anonymous interfaces declared outside of
// their referenced domain — and outside the designated orchestration layer —
// whose method signatures touch types from another domain. This catches
// cmd-style consumers that declare inline ad-hoc abstractions over domain
// types when those abstractions belong in the orchestration layer.
//
// Default severity is Error. The project convention this rule enforces:
// cross-domain abstractions are owned by the orchestration package, not by
// arbitrary wiring code (cmd/, internal/pkg/, etc.). Any anonymous interface
// that abstracts a domain type from outside both the source domain and the
// orchestration layer creates a parallel uncontrolled cross-domain surface.
//
// Skipped:
//   - Test files (_test.go) where mock/fake fixtures naturally use this shape.
//   - Empty interfaces (interface{}) and interfaces without methods.
//   - Embedded interface types (e.g. interface{ io.Reader }).
//   - Same-domain references (an anonymous interface inside internal/domain/X
//     that references internal/domain/X types).
//   - Packages inside internal/<OrchestrationDir>/ — orchestration is the
//     designated cross-domain coordination layer and is exempt by design.
//   - Models with no DomainDir (flat layouts like ConsumerWorker, Batch).
//
// The rule uses the Model's DomainDir and OrchestrationDir settings to
// identify domain and orchestration packages.
func CheckCrossDomainAnonymous(pkgs []*packages.Package, opts ...Option) []Violation {
	cfg := NewConfig(opts...)
	m := cfg.model()
	if m.DomainDir == "" {
		// flat layout — no domain concept, rule does not apply
		return nil
	}

	var violations []Violation
	for _, pkg := range pkgs {
		violations = append(violations, checkCrossDomainAnonymousInPkg(pkg, m, cfg.Sev)...)
	}
	return violations
}

func checkCrossDomainAnonymousInPkg(pkg *packages.Package, m Model, sev Severity) []Violation {
	if isOrchestrationPath(pkg.PkgPath, m) {
		// Orchestration is the designated cross-domain coordination layer.
		// Anonymous interfaces here are by-design.
		return nil
	}
	currentDomain := owningDomainForPath(pkg.PkgPath, m)

	var violations []Violation
	for _, file := range pkg.Syntax {
		if isTestFile(pkg, file) {
			continue
		}
		// Walk top-level declarations explicitly so we never visit a named
		// interface declaration's body as if it were anonymous.
		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.GenDecl:
				// type declarations: skip the declared type itself, but inspect
				// any composite types it contains (struct fields, etc.)
				for _, spec := range d.Specs {
					ts, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}
					// If the spec itself is an interface declaration, the
					// interface is *named*, not anonymous. Skip it entirely.
					if _, isIface := ts.Type.(*ast.InterfaceType); isIface {
						continue
					}
					// Otherwise (e.g. struct type), inspect for anonymous
					// interfaces in field types.
					violations = append(violations, inspectExprForAnonIface(ts.Type, file, pkg, currentDomain, m, sev)...)
				}
			case *ast.FuncDecl:
				if d.Type == nil {
					continue
				}
				if d.Type.Params != nil {
					for _, f := range d.Type.Params.List {
						violations = append(violations, inspectExprForAnonIface(f.Type, file, pkg, currentDomain, m, sev)...)
					}
				}
				if d.Type.Results != nil {
					for _, f := range d.Type.Results.List {
						violations = append(violations, inspectExprForAnonIface(f.Type, file, pkg, currentDomain, m, sev)...)
					}
				}
			}
		}
	}
	return violations
}

// inspectExprForAnonIface walks a type expression and reports any anonymous
// interface that contains methods referencing cross-domain types.
func inspectExprForAnonIface(expr ast.Expr, file *ast.File, pkg *packages.Package, currentDomain string, m Model, sev Severity) []Violation {
	var violations []Violation
	ast.Inspect(expr, func(n ast.Node) bool {
		iface, ok := n.(*ast.InterfaceType)
		if !ok {
			return true
		}
		violations = append(violations, checkAnonymousInterface(iface, file, pkg, currentDomain, m, sev)...)
		// Don't recurse into the interface body — its methods are checked above.
		return false
	})
	return violations
}

// checkAnonymousInterface inspects an anonymous interface's method signatures
// for cross-domain type references and emits a violation per offending domain.
func checkAnonymousInterface(iface *ast.InterfaceType, file *ast.File, pkg *packages.Package, currentDomain string, m Model, sev Severity) []Violation {
	if iface.Methods == nil || len(iface.Methods.List) == 0 {
		return nil
	}

	hasMethodDecl := false
	crossDomains := make(map[string]bool)

	for _, method := range iface.Methods.List {
		// Embedded types have no Names. Skip — we only check declared methods.
		if len(method.Names) == 0 {
			continue
		}
		funcType, ok := method.Type.(*ast.FuncType)
		if !ok {
			continue
		}
		hasMethodDecl = true
		collectCrossDomainRefs(funcType, file, currentDomain, m, crossDomains)
	}

	if !hasMethodDecl || len(crossDomains) == 0 {
		return nil
	}

	pos := pkg.Fset.Position(iface.Pos())
	file_ := pos.Filename
	if rel := relativePathForPackage(pkg, file_); rel != "" {
		file_ = rel
	}

	var domains []string
	for d := range crossDomains {
		domains = append(domains, d)
	}
	sortStrings(domains)

	return []Violation{{
		File:     file_,
		Line:     pos.Line,
		Rule:     "interface.cross-domain-anonymous",
		Message:  fmt.Sprintf("anonymous interface declared in package %q references types from domain(s) %v — cross-domain abstractions must be owned by the orchestration layer, not declared inline outside it", pkg.PkgPath, domains),
		Fix:      "move this adapter/abstraction into internal/" + m.OrchestrationDir + "/ — that's the designated place for cross-domain coordination. Wiring code (cmd/, etc.) should depend on orchestration constructors, not declare its own cross-domain interfaces",
		Severity: sev,
	}}
}

// collectCrossDomainRefs walks a function type's parameters and results and
// records every domain (other than currentDomain) referenced by the
// signatures.
func collectCrossDomainRefs(funcType *ast.FuncType, file *ast.File, currentDomain string, m Model, out map[string]bool) {
	visit := func(expr ast.Expr) {
		walkTypeExprForDomainRefs(expr, file, currentDomain, m, out)
	}
	if funcType.Params != nil {
		for _, f := range funcType.Params.List {
			visit(f.Type)
		}
	}
	if funcType.Results != nil {
		for _, f := range funcType.Results.List {
			visit(f.Type)
		}
	}
}

// walkTypeExprForDomainRefs traverses a type expression looking for selector
// expressions of the form "pkg.Type" where "pkg" resolves (via the file's
// imports) to a domain package different from currentDomain.
func walkTypeExprForDomainRefs(expr ast.Expr, file *ast.File, currentDomain string, m Model, out map[string]bool) {
	ast.Inspect(expr, func(n ast.Node) bool {
		sel, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		ident, ok := sel.X.(*ast.Ident)
		if !ok {
			return true
		}
		importPath := resolveIdentImportPath(file, ident.Name)
		if importPath == "" {
			return true
		}
		refDomain := owningDomainForPath(importPath, m)
		if refDomain == "" {
			return true
		}
		if refDomain != currentDomain {
			out[refDomain] = true
		}
		return true
	})
}

// owningDomainForPath returns the domain segment of an import path, e.g.
// "trelab-server/internal/domain/user/handler/http" → "user". Returns "" if
// the path does not live under internal/<DomainDir>/.
func owningDomainForPath(pkgPath string, m Model) string {
	if m.DomainDir == "" {
		return ""
	}
	parts := strings.Split(pkgPath, "/")
	for i := 0; i+2 < len(parts); i++ {
		if parts[i] == "internal" && parts[i+1] == m.DomainDir {
			return parts[i+2]
		}
	}
	return ""
}

// isOrchestrationPath reports whether a package path lives under
// internal/<OrchestrationDir>/. Used to exempt the orchestration layer from
// cross-domain anonymous-interface checks since orchestration is the
// designated place for cross-domain coordination.
func isOrchestrationPath(pkgPath string, m Model) bool {
	if m.OrchestrationDir == "" {
		return false
	}
	parts := strings.Split(pkgPath, "/")
	for i := 0; i+1 < len(parts); i++ {
		if parts[i] == "internal" && parts[i+1] == m.OrchestrationDir {
			return true
		}
	}
	return false
}

// sortStrings is a tiny helper to keep this file dependency-free of "sort".
func sortStrings(xs []string) {
	for i := 1; i < len(xs); i++ {
		for j := i; j > 0 && xs[j] < xs[j-1]; j-- {
			xs[j], xs[j-1] = xs[j-1], xs[j]
		}
	}
}
