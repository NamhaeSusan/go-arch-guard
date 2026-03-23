package rules

import (
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
)

//nolint:unused // used by CheckDomainIsolation (next task)
func findImportFile(pkg *packages.Package, importPath, projectRoot string) string {
	absRoot, _ := filepath.Abs(projectRoot)
	fset := pkg.Fset
	for _, file := range pkg.Syntax {
		for _, imp := range file.Imports {
			path := strings.Trim(imp.Path.Value, `"`)
			if path == importPath {
				pos := fset.Position(imp.Pos())
				rel, err := filepath.Rel(absRoot, pos.Filename)
				if err != nil {
					return pos.Filename
				}
				return filepath.ToSlash(rel)
			}
		}
	}
	if len(pkg.GoFiles) > 0 {
		rel, err := filepath.Rel(absRoot, pkg.GoFiles[0])
		if err != nil {
			return pkg.GoFiles[0]
		}
		return filepath.ToSlash(rel)
	}
	return pkg.PkgPath
}

//nolint:unused // used by CheckDomainIsolation (next task)
func findImportLine(pkg *packages.Package, importPath string) int {
	fset := pkg.Fset
	for _, file := range pkg.Syntax {
		for _, imp := range file.Imports {
			path := strings.Trim(imp.Path.Value, `"`)
			if path == importPath {
				return fset.Position(imp.Pos()).Line
			}
		}
	}
	return 0
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
	prefix := projectModule + "/"
	if strings.HasPrefix(pkgPath, prefix) {
		return strings.TrimPrefix(pkgPath, prefix)
	}
	return ""
}
