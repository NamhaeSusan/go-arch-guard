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
		if len(pkg.Errors) == 0 {
			result = append(result, pkg)
			continue
		}
		// IllTyped indicates type-checking failed but AST/syntax is available.
		// Type-checking errors (e.g., undefined) are tolerated; parsing errors are not.
		// Parser errors contain phrases like "expected ')", "missing ','", etc.
		hasSyntaxError := false
		for _, e := range pkg.Errors {
			errMsg := e.Error()
			// Look for parser-specific patterns: expected/unexpected followed by quotes,
			// or "missing" followed by a quote or comma phrase.
			if strings.Contains(errMsg, "expected '") || strings.Contains(errMsg, "unexpected") ||
				strings.Contains(errMsg, "missing '") || strings.Contains(errMsg, "syntax error") {
				hasSyntaxError = true
				break
			}
		}
		if pkg.IllTyped && !hasSyntaxError {
			// Type-checking only; package is usable for analysis.
			result = append(result, pkg)
			continue
		}
		// Syntax or load errors are reported and packages skipped.
		for _, e := range pkg.Errors {
			loadErrs = append(loadErrs, e.Error())
		}
	}
	if len(loadErrs) > 0 {
		return result, fmt.Errorf("packages with errors were skipped: %s", strings.Join(loadErrs, "; "))
	}
	return result, nil
}
