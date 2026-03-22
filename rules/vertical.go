package rules

import (
	"strings"

	"golang.org/x/tools/go/packages"
)

func CheckVerticalSlice(pkgs []*packages.Package, projectModule string, projectRoot string, opts ...Option) []Violation {
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

		srcDomain := identifyVerticalDomain(pkgPath, internalPrefix)
		if srcDomain == "" {
			continue
		}

		for importPath := range pkg.Imports {
			if !strings.HasPrefix(importPath, internalPrefix) {
				continue
			}

			dstDomain := identifyVerticalDomain(importPath, internalPrefix)
			if dstDomain == "" {
				continue
			}

			// Same domain → always allowed
			if srcDomain == dstDomain {
				continue
			}

			// Import shared/ → always allowed
			if dstDomain == "shared" {
				continue
			}

			// shared/ importing a domain → violation
			if srcDomain == "shared" {
				violations = append(violations, Violation{
					File:     findImportFile(pkg, importPath, projectRoot),
					Line:     findImportLine(pkg, importPath),
					Rule:     "vertical.shared-imports-domain",
					Message:  `shared package imports domain "` + dstDomain + `"`,
					Fix:      "shared must not import any domain package",
					Severity: cfg.Sev,
				})
				continue
			}

			// Cross-domain from usecase to alias/port → allowed
			if isUsecasePkg(pkgPath, internalPrefix, srcDomain) &&
				isAliasOrPort(importPath, internalPrefix, dstDomain) {
				continue
			}

			// Any other cross-domain import → violation
			violations = append(violations, Violation{
				File:     findImportFile(pkg, importPath, projectRoot),
				Line:     findImportLine(pkg, importPath),
				Rule:     "vertical.cross-domain-isolation",
				Message:  `domain "` + srcDomain + `" imports domain "` + dstDomain + `" from non-usecase package`,
				Fix:      "only app/usecase may import other domains (root alias or port only)",
				Severity: cfg.Sev,
			})
		}
	}
	return violations
}

func identifyVerticalDomain(pkgPath, internalPrefix string) string {
	rel := strings.TrimPrefix(pkgPath, internalPrefix)
	parts := strings.SplitN(rel, "/", 2)
	if len(parts) == 0 || parts[0] == "" {
		return ""
	}
	return parts[0]
}

func isUsecasePkg(pkgPath, internalPrefix, domain string) bool {
	usecasePrefix := internalPrefix + domain + "/app/usecase"
	return pkgPath == usecasePrefix || strings.HasPrefix(pkgPath, usecasePrefix+"/")
}

func isAliasOrPort(importPath, internalPrefix, targetDomain string) bool {
	domainRoot := internalPrefix + targetDomain
	if importPath == domainRoot {
		return true
	}
	portPrefix := domainRoot + "/port"
	return importPath == portPrefix || strings.HasPrefix(importPath, portPrefix+"/")
}
