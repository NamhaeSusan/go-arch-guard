package rules

import (
	"fmt"
	"strings"

	"golang.org/x/tools/go/packages"
)

func CheckDomainIsolation(pkgs []*packages.Package, projectModule string, projectRoot string, opts ...Option) []Violation {
	cfg := NewConfig(opts...)
	m := cfg.model()
	projectModule = resolveModule(pkgs, projectModule)
	projectRoot = resolveRoot(pkgs, projectRoot)
	if warns := validateModule(pkgs, projectModule); len(warns) > 0 {
		return warns
	}
	internalPrefix := projectModule + "/internal/"
	cmdPrefix := projectModule + "/cmd"

	var violations []Violation
	for _, pkg := range pkgs {
		if isExcludedPackage(cfg, pkg.PkgPath, projectModule) {
			continue
		}

		// Check cmd/ packages: they must only import domain aliases, not sub-packages
		if pkg.PkgPath == cmdPrefix || strings.HasPrefix(pkg.PkgPath, cmdPrefix+"/") {
			for impPath := range pkg.Imports {
				if !strings.HasPrefix(impPath, internalPrefix) {
					continue
				}
				impDomain := identifyDomainWith(m, impPath, internalPrefix)
				if impDomain == "" {
					continue
				}
				if isDomainAliasWith(m, impPath, internalPrefix, impDomain) {
					continue
				}
				file, line := findImportPosition(pkg, impPath, projectRoot)
				violations = append(violations, Violation{
					File:     file,
					Line:     line,
					Rule:     "isolation.cmd-deep-import",
					Message:  fmt.Sprintf("cmd/ must only import domain alias, not sub-package %q", impPath),
					Fix:      fmt.Sprintf("import the domain alias package instead: %s%s/%s", internalPrefix, m.DomainDir, impDomain),
					Severity: cfg.Sev,
				})
			}
			continue
		}

		if !strings.HasPrefix(pkg.PkgPath, internalPrefix) {
			continue
		}

		srcDomain := identifyDomainWith(m, pkg.PkgPath, internalPrefix)
		srcIsOrchestration := isOrchestrationPkgWith(m, pkg.PkgPath, internalPrefix)
		srcIsPkg := isPkgPkgWith(m, pkg.PkgPath, internalPrefix)

		for impPath := range pkg.Imports {
			if !strings.HasPrefix(impPath, internalPrefix) {
				continue
			}

			impDomain := identifyDomainWith(m, impPath, internalPrefix)

			// Rule 2: import pkg/ → always allowed
			if isPkgPkgWith(m, impPath, internalPrefix) {
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
				if isDomainAliasWith(m, impPath, internalPrefix, impDomain) {
					continue
				}
				label := m.OrchestrationDir
				if isOrchestrationHandlerWith(m, pkg.PkgPath, internalPrefix) {
					label = m.OrchestrationDir + " handler"
				}
				file, line := findImportPosition(pkg, impPath, projectRoot)
				violations = append(violations, Violation{
					File:     file,
					Line:     line,
					Rule:     "isolation.orchestration-deep-import",
					Message:  fmt.Sprintf("%s must only import domain alias, not sub-package %q", label, impPath),
					Fix:      fmt.Sprintf("import the domain alias package instead: %s%s/%s", internalPrefix, m.DomainDir, impDomain),
					Severity: cfg.Sev,
				})
				continue
			}

			// pkg/ must stay domain-unaware and orchestration-unaware
			if srcIsPkg {
				if impDomain != "" {
					file, line := findImportPosition(pkg, impPath, projectRoot)
					violations = append(violations, Violation{
						File:     file,
						Line:     line,
						Rule:     "isolation.pkg-imports-domain",
						Message:  fmt.Sprintf("%s/ must not import domain %q", m.SharedDir, impDomain),
						Fix:      fmt.Sprintf("%s/ should only contain shared utilities with no domain or orchestration dependencies", m.SharedDir),
						Severity: cfg.Sev,
					})
					continue
				}
				if isOrchestrationPkgWith(m, impPath, internalPrefix) {
					file, line := findImportPosition(pkg, impPath, projectRoot)
					violations = append(violations, Violation{
						File:     file,
						Line:     line,
						Rule:     "isolation.pkg-imports-orchestration",
						Message:  fmt.Sprintf("%s/ must not import %s", m.SharedDir, m.OrchestrationDir),
						Fix:      fmt.Sprintf("move %s-aware code to internal/%s or cmd/", m.OrchestrationDir, m.OrchestrationDir),
						Severity: cfg.Sev,
					})
					continue
				}
			}

			if isOrchestrationPkgWith(m, impPath, internalPrefix) {
				if srcDomain != "" {
					file, line := findImportPosition(pkg, impPath, projectRoot)
					violations = append(violations, Violation{
						File:     file,
						Line:     line,
						Rule:     "isolation.domain-imports-orchestration",
						Message:  fmt.Sprintf("domain %q must not import %s", srcDomain, m.OrchestrationDir),
						Fix:      fmt.Sprintf("move cross-domain coordination to internal/%s callers instead of domain internals", m.OrchestrationDir),
						Severity: cfg.Sev,
					})
					continue
				}
				file, line := findImportPosition(pkg, impPath, projectRoot)
				violations = append(violations, Violation{
					File:     file,
					Line:     line,
					Rule:     "isolation.internal-imports-orchestration",
					Message:  fmt.Sprintf("package %q must not import %s", pkg.PkgPath, m.OrchestrationDir),
					Fix:      fmt.Sprintf("only cmd/ and internal/%s may depend on %s", m.OrchestrationDir, m.OrchestrationDir),
					Severity: cfg.Sev,
				})
				continue
			}

			// Rule 7: domain A importing domain B → violation
			if srcDomain != "" && impDomain != "" && srcDomain != impDomain {
				file, line := findImportPosition(pkg, impPath, projectRoot)
				violations = append(violations, Violation{
					File:     file,
					Line:     line,
					Rule:     "isolation.cross-domain",
					Message:  fmt.Sprintf("domain %q must not import domain %q", srcDomain, impDomain),
					Fix:      fmt.Sprintf("use %s/ for cross-domain orchestration or move shared types to %s/", m.OrchestrationDir, m.SharedDir),
					Severity: cfg.Sev,
				})
				continue
			}

			// Rule 8: non-domain internal packages other than orchestration/cmd/pkg must not import domains
			if srcDomain == "" && impDomain != "" {
				file, line := findImportPosition(pkg, impPath, projectRoot)
				violations = append(violations, Violation{
					File:     file,
					Line:     line,
					Rule:     "isolation.internal-imports-domain",
					Message:  fmt.Sprintf("package %q must not import domain %q", pkg.PkgPath, impDomain),
					Fix:      fmt.Sprintf("move domain orchestration to internal/%s or app wiring to cmd/", m.OrchestrationDir),
					Severity: cfg.Sev,
				})
				continue
			}
		}
	}
	return violations
}

func isOrchestrationPkgWith(m Model, pkgPath, internalPrefix string) bool {
	rel := strings.TrimPrefix(pkgPath, internalPrefix)
	return rel == m.OrchestrationDir || strings.HasPrefix(rel, m.OrchestrationDir+"/")
}

func isPkgPkgWith(m Model, pkgPath, internalPrefix string) bool {
	rel := strings.TrimPrefix(pkgPath, internalPrefix)
	return rel == m.SharedDir || strings.HasPrefix(rel, m.SharedDir+"/")
}

func isDomainAliasWith(m Model, importPath, internalPrefix, domain string) bool {
	return importPath == internalPrefix+m.DomainDir+"/"+domain
}

func isOrchestrationHandlerWith(m Model, pkgPath, internalPrefix string) bool {
	rel := strings.TrimPrefix(pkgPath, internalPrefix)
	return strings.HasPrefix(rel, m.OrchestrationDir+"/handler")
}
