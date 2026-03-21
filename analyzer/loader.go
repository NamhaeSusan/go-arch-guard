package analyzer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
)

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
			packages.NeedSyntax | packages.NeedTypes,
		Dir: absDir,
	}
	pkgs, err := packages.Load(cfg, prefixed...)
	if err != nil {
		return nil, fmt.Errorf("load packages: %w", err)
	}

	var result []*packages.Package
	var loadErrs []string
	for _, pkg := range pkgs {
		if len(pkg.Errors) > 0 {
			for _, e := range pkg.Errors {
				loadErrs = append(loadErrs, e.Error())
			}
			continue
		}
		result = append(result, pkg)
	}
	if len(loadErrs) > 0 {
		return result, fmt.Errorf("packages with errors were skipped: %s", strings.Join(loadErrs, "; "))
	}
	return result, nil
}
