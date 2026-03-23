package rules

import (
	"fmt"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
)

func findImportPosition(pkg *packages.Package, importPath, projectRoot string) (string, int) {
	absRoot, _ := filepath.Abs(projectRoot)
	fset := pkg.Fset
	for _, file := range pkg.Syntax {
		for _, imp := range file.Imports {
			path := strings.Trim(imp.Path.Value, `"`)
			if path == importPath {
				pos := fset.Position(imp.Pos())
				rel, err := filepath.Rel(absRoot, pos.Filename)
				if err != nil {
					return pos.Filename, pos.Line
				}
				return filepath.ToSlash(rel), pos.Line
			}
		}
	}
	if len(pkg.GoFiles) > 0 {
		rel, err := filepath.Rel(absRoot, pkg.GoFiles[0])
		if err != nil {
			return pkg.GoFiles[0], 0
		}
		return filepath.ToSlash(rel), 0
	}
	return pkg.PkgPath, 0
}

func relativePathForPackage(pkg *packages.Package, path string) string {
	if pkg != nil && pkg.Module != nil && pkg.Module.Dir != "" {
		rel, err := filepath.Rel(pkg.Module.Dir, path)
		if err == nil {
			return filepath.ToSlash(rel)
		}
	}
	return filepath.ToSlash(path)
}

func resolveModule(pkgs []*packages.Package, explicit string) string {
	if explicit != "" {
		return explicit
	}
	for _, pkg := range pkgs {
		if pkg.Module != nil && pkg.Module.Path != "" {
			return pkg.Module.Path
		}
	}
	return ""
}

func resolveRoot(pkgs []*packages.Package, explicit string) string {
	if explicit != "" {
		return explicit
	}
	for _, pkg := range pkgs {
		if pkg.Module != nil && pkg.Module.Dir != "" {
			return pkg.Module.Dir
		}
	}
	return ""
}

func validateModule(pkgs []*packages.Package, projectModule string) []Violation {
	if projectModule == "" {
		return []Violation{{
			Rule:     "meta.no-matching-packages",
			Message:  "project module could not be determined — all import checks will be skipped",
			Fix:      "pass a non-empty module path or ensure packages are loaded with NeedModule",
			Severity: Warning,
		}}
	}
	prefix := projectModule + "/"
	for _, pkg := range pkgs {
		if pkg.PkgPath == projectModule || strings.HasPrefix(pkg.PkgPath, prefix) {
			return nil
		}
	}
	return []Violation{{
		Rule:     "meta.no-matching-packages",
		Message:  fmt.Sprintf("module %q does not match any loaded package — all import checks will be skipped", projectModule),
		Fix:      "verify the module argument matches go.mod (e.g. pass the value from pkgs[0].Module.Path)",
		Severity: Warning,
	}}
}

func isExcludedPackage(cfg Config, pkgPath, projectModule string) bool {
	return cfg.IsExcluded(projectRelativePackagePath(pkgPath, projectModule))
}

func projectRelativePackagePath(pkgPath, projectModule string) string {
	if pkgPath == "" || projectModule == "" {
		return ""
	}
	if pkgPath == projectModule {
		return "."
	}
	if rel, ok := strings.CutPrefix(pkgPath, projectModule+"/"); ok {
		return rel
	}
	return ""
}
