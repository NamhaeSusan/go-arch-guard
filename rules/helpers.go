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
				return rel
			}
		}
	}
	if len(pkg.GoFiles) > 0 {
		rel, err := filepath.Rel(absRoot, pkg.GoFiles[0])
		if err != nil {
			return pkg.GoFiles[0]
		}
		return rel
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
