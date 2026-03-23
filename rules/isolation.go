package rules

import (
	"fmt"
	"strings"

	"golang.org/x/tools/go/packages"
)

func CheckDomainIsolation(pkgs []*packages.Package, projectModule string, projectRoot string, opts ...Option) []Violation {
	cfg := NewConfig(opts...)
	internalPrefix := projectModule + "/internal/"

	var violations []Violation
	for _, pkg := range pkgs {
		if cfg.IsExcluded(pkg.PkgPath) {
			continue
		}
		if !strings.HasPrefix(pkg.PkgPath, internalPrefix) {
			continue
		}

		srcDomain := identifyDomain(pkg.PkgPath, internalPrefix)
		srcIsSaga := isSagaPkg(pkg.PkgPath, internalPrefix)
		srcIsPkg := isPkgPkg(pkg.PkgPath, internalPrefix)

		for impPath := range pkg.Imports {
			if !strings.HasPrefix(impPath, internalPrefix) {
				continue
			}

			impDomain := identifyDomain(impPath, internalPrefix)

			// Rule 2: import pkg/ → always allowed
			if isPkgPkg(impPath, internalPrefix) {
				continue
			}

			// Rule 3: pkg/ importing a domain → violation
			if srcIsPkg && impDomain != "" {
				violations = append(violations, Violation{
					File:     findImportFile(pkg, impPath, projectRoot),
					Line:     findImportLine(pkg, impPath),
					Rule:     "isolation.pkg-imports-domain",
					Message:  fmt.Sprintf("pkg/ must not import domain %q", impDomain),
					Fix:      "pkg/ should only contain shared utilities with no domain dependencies",
					Severity: cfg.Sev,
				})
				continue
			}

			// Rule 1: same domain → always allowed
			if srcDomain != "" && impDomain == srcDomain {
				continue
			}

			// Saga rules
			if srcIsSaga {
				if impDomain == "" {
					continue
				}
				isSrcHandler := isSagaHandler(pkg.PkgPath, internalPrefix)
				// Rule 6: saga/handler/ importing saga internal → allowed (handled above by saga prefix check)
				// Rule 4: saga (non-handler) importing domain alias → allowed
				if !isSrcHandler && isDomainAlias(impPath, internalPrefix, impDomain) {
					continue
				}
				// Rule 5: saga (non-handler) importing domain sub-package → violation
				if !isSrcHandler && !isDomainAlias(impPath, internalPrefix, impDomain) {
					violations = append(violations, Violation{
						File:     findImportFile(pkg, impPath, projectRoot),
						Line:     findImportLine(pkg, impPath),
						Rule:     "isolation.saga-deep-import",
						Message:  fmt.Sprintf("saga must only import domain alias, not sub-package %q", impPath),
						Fix:      fmt.Sprintf("import the domain alias package instead: %sdomain/%s", internalPrefix, impDomain),
						Severity: cfg.Sev,
					})
					continue
				}
				// saga handler can import saga internals (rule 6) but not domain sub-packages
				if isSrcHandler && impDomain != "" && !isDomainAlias(impPath, internalPrefix, impDomain) {
					violations = append(violations, Violation{
						File:     findImportFile(pkg, impPath, projectRoot),
						Line:     findImportLine(pkg, impPath),
						Rule:     "isolation.saga-deep-import",
						Message:  fmt.Sprintf("saga handler must only import domain alias, not sub-package %q", impPath),
						Fix:      fmt.Sprintf("import the domain alias package instead: %sdomain/%s", internalPrefix, impDomain),
						Severity: cfg.Sev,
					})
					continue
				}
				continue
			}

			// Rule 7: domain A importing domain B → violation
			if srcDomain != "" && impDomain != "" && srcDomain != impDomain {
				violations = append(violations, Violation{
					File:     findImportFile(pkg, impPath, projectRoot),
					Line:     findImportLine(pkg, impPath),
					Rule:     "isolation.cross-domain",
					Message:  fmt.Sprintf("domain %q must not import domain %q", srcDomain, impDomain),
					Fix:      "use saga/ for cross-domain orchestration or move shared types to pkg/",
					Severity: cfg.Sev,
				})
				continue
			}
		}
	}
	return violations
}

func identifyDomain(pkgPath, internalPrefix string) string {
	rel := strings.TrimPrefix(pkgPath, internalPrefix)
	if !strings.HasPrefix(rel, "domain/") {
		return ""
	}
	after := strings.TrimPrefix(rel, "domain/")
	parts := strings.SplitN(after, "/", 2)
	if len(parts) == 0 || parts[0] == "" {
		return ""
	}
	return parts[0]
}

func isSagaPkg(pkgPath, internalPrefix string) bool {
	rel := strings.TrimPrefix(pkgPath, internalPrefix)
	return rel == "saga" || strings.HasPrefix(rel, "saga/")
}

func isPkgPkg(pkgPath, internalPrefix string) bool {
	rel := strings.TrimPrefix(pkgPath, internalPrefix)
	return rel == "pkg" || strings.HasPrefix(rel, "pkg/")
}

func isDomainAlias(importPath, internalPrefix, domain string) bool {
	return importPath == internalPrefix+"domain/"+domain
}

func isSagaHandler(pkgPath, internalPrefix string) bool {
	rel := strings.TrimPrefix(pkgPath, internalPrefix)
	return strings.HasPrefix(rel, "saga/handler")
}
