package rules

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/NamhaeSusan/go-arch-guard/core/analysisutil"
	"golang.org/x/tools/go/packages"
)

// relPathFromRoot turns an absolute file path into a forward-slash path
// relative to projectRoot. Falls back to the original path on error.
func relPathFromRoot(projectRoot, filename string) string {
	return analysisutil.RelPathFromRoot(projectRoot, filename)
}

func findImportPosition(pkg *packages.Package, importPath, projectRoot string) (string, int) {
	return analysisutil.FindImportPosition(pkg, importPath, projectRoot)
}

func relativePathForPackage(pkg *packages.Package, path string) string {
	return analysisutil.RelativePathForPackage(pkg, path)
}

func resolveModule(pkgs []*packages.Package, explicit string) string {
	return analysisutil.ResolveModule(pkgs, explicit)
}

func resolveRoot(pkgs []*packages.Package, explicit string) string {
	return analysisutil.ResolveRoot(pkgs, explicit)
}

func validateModule(pkgs []*packages.Package, projectModule string) []Violation {
	if projectModule == "" {
		return []Violation{{
			Rule:              "meta.no-matching-packages",
			Message:           "project module could not be determined — all import checks will be skipped",
			Fix:               "pass a non-empty module path or ensure packages are loaded with NeedModule",
			DefaultSeverity:   Warning,
			EffectiveSeverity: Warning,
		}}
	}
	prefix := projectModule + "/"
	for _, pkg := range pkgs {
		if pkg.PkgPath == projectModule || strings.HasPrefix(pkg.PkgPath, prefix) {
			return nil
		}
	}
	return []Violation{{
		Rule:              "meta.no-matching-packages",
		Message:           fmt.Sprintf("module %q does not match any loaded package — all import checks will be skipped", projectModule),
		Fix:               "verify the module argument matches go.mod (e.g. pass the value from pkgs[0].Module.Path)",
		DefaultSeverity:   Warning,
		EffectiveSeverity: Warning,
	}}
}

// resolveIdentImportPath returns the import path that identName refers to
// by scanning the file's import declarations. Returns "" if not found.
func resolveIdentImportPath(file *ast.File, identName string) string {
	return analysisutil.ResolveIdentImportPath(file, identName)
}

func isExcludedPackage(cfg Config, pkgPath, projectModule string) bool {
	return cfg.IsExcluded(projectRelativePackagePath(pkgPath, projectModule))
}

func projectRelativePackagePath(pkgPath, projectModule string) string {
	return analysisutil.ProjectRelativePackagePath(pkgPath, projectModule)
}
