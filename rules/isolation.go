package rules

import (
	"fmt"
	"strings"

	"golang.org/x/tools/go/packages"
)

func CheckDomainIsolation(pkgs []*packages.Package, projectModule string, projectRoot string, opts ...Option) []Violation {
	cfg := NewConfig(opts...)
	internalPrefix := projectModule + "/internal/"
	cmdPrefix := projectModule + "/cmd"

	var violations []Violation
	for _, pkg := range pkgs {
		if cfg.IsExcluded(pkg.PkgPath) {
			continue
		}

		// Check cmd/ packages: they must only import domain aliases, not sub-packages
		if pkg.PkgPath == cmdPrefix || strings.HasPrefix(pkg.PkgPath, cmdPrefix+"/") {
			for impPath := range pkg.Imports {
				if !strings.HasPrefix(impPath, internalPrefix) {
					continue
				}
				impDomain := identifyDomain(impPath, internalPrefix)
				if impDomain == "" {
					continue
				}
				if isDomainAlias(impPath, internalPrefix, impDomain) {
					continue // alias import is allowed
				}
				violations = append(violations, Violation{
					File:     findImportFile(pkg, impPath, projectRoot),
					Line:     findImportLine(pkg, impPath),
					Rule:     "isolation.cmd-deep-import",
					Message:  fmt.Sprintf("cmd/ must only import domain alias, not sub-package %q", impPath),
					Fix:      fmt.Sprintf("import the domain alias package instead: %sdomain/%s", internalPrefix, impDomain),
					Severity: cfg.Sev,
				})
			}
			continue
		}

		if !strings.HasPrefix(pkg.PkgPath, internalPrefix) {
			continue
		}

		srcDomain := identifyDomain(pkg.PkgPath, internalPrefix)
		srcIsOrchestration := isOrchestrationPkg(pkg.PkgPath, internalPrefix)
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

			// Rule 1: same domain → always allowed
			if srcDomain != "" && impDomain == srcDomain {
				continue
			}

			// Orchestration rules
			if srcIsOrchestration {
				if impDomain == "" {
					continue
				}
				isSrcHandler := isOrchestrationHandler(pkg.PkgPath, internalPrefix)
				// Rule 6: orchestration/handler/ importing orchestration internal → allowed (handled above by orchestration prefix check)
				// Rule 4: orchestration (non-handler) importing domain alias → allowed
				if !isSrcHandler && isDomainAlias(impPath, internalPrefix, impDomain) {
					continue
				}
				// Rule 5: orchestration (non-handler) importing domain sub-package → violation
				if !isSrcHandler && !isDomainAlias(impPath, internalPrefix, impDomain) {
					violations = append(violations, Violation{
						File:     findImportFile(pkg, impPath, projectRoot),
						Line:     findImportLine(pkg, impPath),
						Rule:     "isolation.orchestration-deep-import",
						Message:  fmt.Sprintf("orchestration must only import domain alias, not sub-package %q", impPath),
						Fix:      fmt.Sprintf("import the domain alias package instead: %sdomain/%s", internalPrefix, impDomain),
						Severity: cfg.Sev,
					})
					continue
				}
				// orchestration handler can import orchestration internals (rule 6) but not domain sub-packages
				if isSrcHandler && impDomain != "" && !isDomainAlias(impPath, internalPrefix, impDomain) {
					violations = append(violations, Violation{
						File:     findImportFile(pkg, impPath, projectRoot),
						Line:     findImportLine(pkg, impPath),
						Rule:     "isolation.orchestration-deep-import",
						Message:  fmt.Sprintf("orchestration handler must only import domain alias, not sub-package %q", impPath),
						Fix:      fmt.Sprintf("import the domain alias package instead: %sdomain/%s", internalPrefix, impDomain),
						Severity: cfg.Sev,
					})
					continue
				}
				continue
			}

			// pkg/ must stay domain-unaware and orchestration-unaware
			if srcIsPkg {
				if impDomain != "" {
					violations = append(violations, Violation{
						File:     findImportFile(pkg, impPath, projectRoot),
						Line:     findImportLine(pkg, impPath),
						Rule:     "isolation.pkg-imports-domain",
						Message:  fmt.Sprintf("pkg/ must not import domain %q", impDomain),
						Fix:      "pkg/ should only contain shared utilities with no domain or orchestration dependencies",
						Severity: cfg.Sev,
					})
					continue
				}
				if isOrchestrationPkg(impPath, internalPrefix) {
					violations = append(violations, Violation{
						File:     findImportFile(pkg, impPath, projectRoot),
						Line:     findImportLine(pkg, impPath),
						Rule:     "isolation.pkg-imports-orchestration",
						Message:  "pkg/ must not import orchestration",
						Fix:      "move orchestration-aware code to internal/orchestration or cmd/",
						Severity: cfg.Sev,
					})
					continue
				}
			}

			if srcDomain != "" && isOrchestrationPkg(impPath, internalPrefix) {
				violations = append(violations, Violation{
					File:     findImportFile(pkg, impPath, projectRoot),
					Line:     findImportLine(pkg, impPath),
					Rule:     "isolation.domain-imports-orchestration",
					Message:  fmt.Sprintf("domain %q must not import orchestration", srcDomain),
					Fix:      "move cross-domain coordination to internal/orchestration callers instead of domain internals",
					Severity: cfg.Sev,
				})
				continue
			}

			// Rule 7: domain A importing domain B → violation
			if srcDomain != "" && impDomain != "" && srcDomain != impDomain {
				violations = append(violations, Violation{
					File:     findImportFile(pkg, impPath, projectRoot),
					Line:     findImportLine(pkg, impPath),
					Rule:     "isolation.cross-domain",
					Message:  fmt.Sprintf("domain %q must not import domain %q", srcDomain, impDomain),
					Fix:      "use orchestration/ for cross-domain orchestration or move shared types to pkg/",
					Severity: cfg.Sev,
				})
				continue
			}

			// Rule 8: non-domain internal packages other than orchestration/cmd/pkg must not import domains
			if srcDomain == "" && impDomain != "" {
				violations = append(violations, Violation{
					File:     findImportFile(pkg, impPath, projectRoot),
					Line:     findImportLine(pkg, impPath),
					Rule:     "isolation.internal-imports-domain",
					Message:  fmt.Sprintf("package %q must not import domain %q", pkg.PkgPath, impDomain),
					Fix:      "move domain orchestration to internal/orchestration or app wiring to cmd/",
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

func isOrchestrationPkg(pkgPath, internalPrefix string) bool {
	rel := strings.TrimPrefix(pkgPath, internalPrefix)
	return rel == "orchestration" || strings.HasPrefix(rel, "orchestration/")
}

func isPkgPkg(pkgPath, internalPrefix string) bool {
	rel := strings.TrimPrefix(pkgPath, internalPrefix)
	return rel == "pkg" || strings.HasPrefix(rel, "pkg/")
}

func isDomainAlias(importPath, internalPrefix, domain string) bool {
	return importPath == internalPrefix+"domain/"+domain
}

func isOrchestrationHandler(pkgPath, internalPrefix string) bool {
	rel := strings.TrimPrefix(pkgPath, internalPrefix)
	return strings.HasPrefix(rel, "orchestration/handler")
}
