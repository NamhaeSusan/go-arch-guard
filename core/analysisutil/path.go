package analysisutil

import (
	"path/filepath"
	"strings"

	"github.com/NamhaeSusan/go-arch-guard/core"
	"golang.org/x/tools/go/packages"
)

func NormalizeMatchPath(path string) string {
	path = strings.ReplaceAll(path, `\`, `/`)
	path = filepath.ToSlash(path)
	for strings.HasPrefix(path, "./") {
		path = strings.TrimPrefix(path, "./")
	}
	return path
}

func RelPathFromRoot(projectRoot, filename string) string {
	absRoot, err := filepath.Abs(projectRoot)
	if err != nil {
		return filename
	}
	rel, err := filepath.Rel(absRoot, filename)
	if err != nil {
		return filename
	}
	return filepath.ToSlash(rel)
}

func FindImportPosition(pkg *packages.Package, importPath, projectRoot string) (string, int) {
	fset := pkg.Fset
	for _, file := range pkg.Syntax {
		for _, imp := range file.Imports {
			path := strings.Trim(imp.Path.Value, `"`)
			if path == importPath {
				pos := fset.Position(imp.Pos())
				return RelPathFromRoot(projectRoot, pos.Filename), pos.Line
			}
		}
	}
	if len(pkg.GoFiles) > 0 {
		return RelPathFromRoot(projectRoot, pkg.GoFiles[0]), 0
	}
	return pkg.PkgPath, 0
}

func RelativePathForPackage(pkg *packages.Package, path string) string {
	if pkg != nil && pkg.Module != nil && pkg.Module.Dir != "" {
		rel, err := filepath.Rel(pkg.Module.Dir, path)
		if err == nil {
			return filepath.ToSlash(rel)
		}
	}
	return filepath.ToSlash(path)
}

func ResolveModule(pkgs []*packages.Package, explicit string) string {
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

func ResolveRoot(pkgs []*packages.Package, explicit string) string {
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

func ResolveModuleFromContext(ctx *core.Context, explicit string) string {
	if ctx == nil {
		return explicit
	}
	return ResolveModule(ctx.Pkgs(), firstNonEmpty(explicit, ctx.Module()))
}

func ResolveRootFromContext(ctx *core.Context, explicit string) string {
	if ctx == nil {
		return explicit
	}
	return ResolveRoot(ctx.Pkgs(), firstNonEmpty(explicit, ctx.Root()))
}

func ProjectRelativePackagePath(pkgPath, projectModule string) string {
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

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
