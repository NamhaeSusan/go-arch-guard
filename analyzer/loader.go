package analyzer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
)

// Load parses Go packages matching the given patterns under dir.
// When some packages contain errors (e.g. type-check failures), they are
// skipped and the successfully loaded packages are returned alongside a
// non-nil error describing what was skipped. Callers that want partial
// analysis should check len(pkgs) rather than treating err as fatal.
func Load(dir string, patterns ...string) ([]*packages.Package, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("resolve dir: %w", err)
	}
	if _, err := os.Stat(absDir); err != nil {
		return nil, fmt.Errorf("project root not found: %w", err)
	}

	prefixed := make([]string, len(patterns))
	for i, p := range patterns {
		prefixed[i] = "./" + p
	}

	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedImports | packages.NeedFiles |
			packages.NeedSyntax | packages.NeedModule |
			packages.NeedTypes | packages.NeedTypesInfo | packages.NeedDeps,
		Dir: absDir,
	}
	pkgs, err := packages.Load(cfg, prefixed...)
	if err != nil {
		return nil, fmt.Errorf("load packages: %w", err)
	}

	var result []*packages.Package
	var loadErrs []string
	for _, pkg := range pkgs {
		if depErr := firstFatalErrorInDeps(pkg); depErr != "" {
			loadErrs = append(loadErrs, depErr)
			continue
		}
		if len(pkg.Errors) == 0 {
			result = append(result, pkg)
			continue
		}
		// Tolerate pure type-check failures on the root package itself (e.g.,
		// undefined identifier) because downstream rules can still inspect the
		// AST/imports. Reject parse and list errors since those leave TypesInfo
		// unreliable.
		onlyTypeErrors := pkg.IllTyped
		for _, e := range pkg.Errors {
			if e.Kind != packages.TypeError {
				onlyTypeErrors = false
				break
			}
		}
		if onlyTypeErrors {
			result = append(result, pkg)
			continue
		}
		// Syntax or load errors on the root package are reported and it is skipped.
		for _, e := range pkg.Errors {
			loadErrs = append(loadErrs, e.Error())
		}
	}
	if len(loadErrs) > 0 {
		return result, fmt.Errorf("packages with errors were skipped: %s", strings.Join(loadErrs, "; "))
	}
	return result, nil
}

// firstFatalErrorInDeps traverses the transitive import graph of pkg
// (excluding pkg itself) and returns the first fatal error (ParseError or
// ListError) found in any dependency. TypeError and UnknownError are
// tolerated to avoid false positives from transient toolchain/export-data
// issues. An empty string means all deps are clean.
func firstFatalErrorInDeps(root *packages.Package) string {
	var found string
	packages.Visit([]*packages.Package{root}, nil, func(dep *packages.Package) {
		if found != "" || dep == root {
			return
		}
		for _, e := range dep.Errors {
			if e.Kind == packages.ParseError || e.Kind == packages.ListError {
				found = fmt.Sprintf("dependency %s: %s", dep.PkgPath, e.Error())
				return
			}
		}
	})
	return found
}
