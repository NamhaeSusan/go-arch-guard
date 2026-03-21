package rules

import (
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
)

type Layer int

const (
	LayerUnknown Layer = iota
	LayerDomain
	LayerApp
	LayerHandler
	LayerInfra
)

var allowedImports = map[Layer][]Layer{
	LayerHandler: {LayerApp, LayerDomain},
	LayerApp:     {LayerDomain},
	LayerInfra:   {LayerDomain},
	LayerDomain:  {},
}

func CheckDependency(pkgs []*packages.Package, projectModule string, projectRoot string, opts ...Option) []Violation {
	cfg := NewConfig(opts...)
	var violations []Violation
	internalPrefix := projectModule + "/internal/"

	for _, pkg := range pkgs {
		pkgPath := pkg.PkgPath
		if !strings.HasPrefix(pkgPath, internalPrefix) {
			continue
		}

		relPath := strings.TrimPrefix(pkgPath, projectModule+"/")
		if cfg.IsExcluded(relPath + "/") {
			continue
		}

		srcLayer := classifyLayer(pkgPath, internalPrefix)
		if srcLayer == LayerUnknown {
			continue
		}
		srcDomain := extractDomain(pkgPath, internalPrefix)

		for importPath := range pkg.Imports {
			if !strings.HasPrefix(importPath, internalPrefix) {
				continue
			}
			dstLayer := classifyLayer(importPath, internalPrefix)
			if dstLayer == LayerUnknown {
				continue
			}

			if srcLayer == LayerDomain && dstLayer == LayerDomain {
				dstDomain := extractDomain(importPath, internalPrefix)
				if srcDomain != dstDomain {
					violations = append(violations, Violation{
						File:     findImportFile(pkg, importPath, projectRoot),
						Line:     findImportLine(pkg, importPath),
						Rule:     "dependency.domain-isolation",
						Message:  `domain "` + srcDomain + `" imports domain "` + dstDomain + `" directly`,
						Fix:      "use app layer to coordinate between domains",
						Severity: cfg.Sev,
					})
				}
				continue
			}

			if srcLayer == LayerDomain && dstLayer != LayerDomain {
				violations = append(violations, Violation{
					File:     findImportFile(pkg, importPath, projectRoot),
					Line:     findImportLine(pkg, importPath),
					Rule:     "dependency.domain-purity",
					Message:  `"` + relPath + `" imports "` + strings.TrimPrefix(importPath, projectModule+"/") + `"`,
					Fix:      "domain must not depend on any other layer",
					Severity: cfg.Sev,
				})
				continue
			}

			// Same-layer sub-package imports are allowed (except domain,
			// which is handled above with cross-domain isolation).
			if srcLayer == dstLayer {
				continue
			}

			if !isAllowed(srcLayer, dstLayer) {
				violations = append(violations, Violation{
					File:     findImportFile(pkg, importPath, projectRoot),
					Line:     findImportLine(pkg, importPath),
					Rule:     "dependency.layer-direction",
					Message:  `"` + layerName(srcLayer) + `" imports "` + layerName(dstLayer) + `" directly`,
					Fix:      layerFix(srcLayer),
					Severity: cfg.Sev,
				})
			}
		}
	}
	return violations
}

func classifyLayer(pkgPath, internalPrefix string) Layer {
	rel := strings.TrimPrefix(pkgPath, internalPrefix)
	parts := strings.SplitN(rel, "/", 2)
	switch parts[0] {
	case "domain":
		return LayerDomain
	case "app":
		return LayerApp
	case "handler":
		return LayerHandler
	case "infra":
		return LayerInfra
	default:
		return LayerUnknown
	}
}

func extractDomain(pkgPath, internalPrefix string) string {
	rel := strings.TrimPrefix(pkgPath, internalPrefix)
	parts := strings.SplitN(rel, "/", 3)
	if len(parts) >= 2 {
		return parts[1]
	}
	return ""
}

func isAllowed(src, dst Layer) bool {
	for _, a := range allowedImports[src] {
		if dst == a {
			return true
		}
	}
	return false
}

func layerName(l Layer) string {
	switch l {
	case LayerDomain:
		return "domain"
	case LayerApp:
		return "app"
	case LayerHandler:
		return "handler"
	case LayerInfra:
		return "infra"
	default:
		return "unknown"
	}
}

func layerFix(src Layer) string {
	switch src {
	case LayerHandler:
		return "handler must only depend on app or domain"
	case LayerApp:
		return "app must only depend on domain"
	case LayerInfra:
		return "infra must only depend on domain"
	default:
		return "check layer dependency direction"
	}
}

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
