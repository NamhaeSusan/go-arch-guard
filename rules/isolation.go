package rules

import (
	"fmt"
	"strings"

	"golang.org/x/tools/go/packages"
)

func CheckDomainIsolation(pkgs []*packages.Package, projectModule string, projectRoot string, opts ...Option) []Violation {
	cfg := NewConfig(opts...)
	m := cfg.model()

	// Flat layout has no domains — skip isolation checks entirely.
	if m.DomainDir == "" {
		return nil
	}

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
			// When alias is not required, cmd/ may import any domain sub-package.
			if !m.RequireAlias {
				continue
			}
			for impPath := range pkg.Imports {
				if !strings.HasPrefix(impPath, internalPrefix) {
					continue
				}
				imp := classifyInternalPackage(m, impPath, internalPrefix)
				// Only domain sub-packages (kindDomain) trigger cmd-deep-import.
				// Domain root (alias) is allowed; shared/orchestration/unclassified
				// are not domain imports and fall outside this rule.
				if imp.Kind != kindDomain {
					continue
				}
				file, line := findImportPosition(pkg, impPath, projectRoot)
				violations = append(violations, Violation{
					File:     file,
					Line:     line,
					Rule:     "isolation.cmd-deep-import",
					Message:  fmt.Sprintf("cmd/ must only import domain alias, not sub-package %q", impPath),
					Fix:      fmt.Sprintf("import the domain alias package instead: %s%s/%s", internalPrefix, m.DomainDir, imp.Domain),
					Severity: cfg.Sev,
				})
			}
			continue
		}

		if !strings.HasPrefix(pkg.PkgPath, internalPrefix) {
			continue
		}

		src := classifyInternalPackage(m, pkg.PkgPath, internalPrefix)
		// kindUnclassified sources are packages under internal/ that don't fit
		// any known bucket (domain/, orchestration/, shared/). They may still
		// import domains/orchestration and must be checked as "stray" sources.

		for impPath := range pkg.Imports {
			if !strings.HasPrefix(impPath, internalPrefix) {
				continue
			}

			imp := classifyInternalPackage(m, impPath, internalPrefix)

			// Rule 2: import pkg/ → always allowed for anyone except pkg/ itself
			// (pkg/ → pkg/ is just intra-shared and still allowed).
			if imp.Kind == kindShared && src.Kind != kindShared {
				continue
			}

			// Rule 1: same domain → always allowed (kindDomain or kindDomainRoot
			// in the same domain).
			if (src.Kind == kindDomain || src.Kind == kindDomainRoot) &&
				(imp.Kind == kindDomain || imp.Kind == kindDomainRoot) &&
				src.Domain != "" && src.Domain == imp.Domain {
				continue
			}

			// Orchestration sources
			if src.Kind == kindOrchestration {
				if imp.Kind != kindDomain && imp.Kind != kindDomainRoot {
					continue
				}
				// When alias is not required, orchestration may import
				// any sub-package within domains (composition root pattern).
				if !m.RequireAlias {
					continue
				}
				if imp.Kind == kindDomainRoot {
					continue
				}
				// imp is a domain sub-package (kindDomain) and alias is required.
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
					Fix:      fmt.Sprintf("import the domain alias package instead: %s%s/%s", internalPrefix, m.DomainDir, imp.Domain),
					Severity: cfg.Sev,
				})
				continue
			}

			// pkg/ must stay domain-unaware and orchestration-unaware
			if src.Kind == kindShared {
				if imp.Kind == kindDomain || imp.Kind == kindDomainRoot {
					file, line := findImportPosition(pkg, impPath, projectRoot)
					violations = append(violations, Violation{
						File:     file,
						Line:     line,
						Rule:     "isolation.pkg-imports-domain",
						Message:  fmt.Sprintf("%s/ must not import domain %q", m.SharedDir, imp.Domain),
						Fix:      fmt.Sprintf("%s/ should only contain shared utilities with no domain or orchestration dependencies", m.SharedDir),
						Severity: cfg.Sev,
					})
					continue
				}
				if imp.Kind == kindOrchestration {
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

			if imp.Kind == kindOrchestration {
				if src.Kind == kindDomain || src.Kind == kindDomainRoot {
					file, line := findImportPosition(pkg, impPath, projectRoot)
					violations = append(violations, Violation{
						File:     file,
						Line:     line,
						Rule:     "isolation.domain-imports-orchestration",
						Message:  fmt.Sprintf("domain %q must not import %s", src.Domain, m.OrchestrationDir),
						Fix:      fmt.Sprintf("move cross-domain coordination to internal/%s callers instead of domain internals", m.OrchestrationDir),
						Severity: cfg.Sev,
					})
					continue
				}
				file, line := findImportPosition(pkg, impPath, projectRoot)
				violations = append(violations, Violation{
					File:     file,
					Line:     line,
					Rule:     "isolation.stray-imports-orchestration",
					Message:  fmt.Sprintf("package %q must not import %s", pkg.PkgPath, m.OrchestrationDir),
					Fix:      fmt.Sprintf("only cmd/ and internal/%s may depend on %s", m.OrchestrationDir, m.OrchestrationDir),
					Severity: cfg.Sev,
				})
				continue
			}

			// Rule 7: domain A importing domain B → violation
			if (src.Kind == kindDomain || src.Kind == kindDomainRoot) &&
				(imp.Kind == kindDomain || imp.Kind == kindDomainRoot) &&
				src.Domain != "" && imp.Domain != "" && src.Domain != imp.Domain {
				file, line := findImportPosition(pkg, impPath, projectRoot)
				violations = append(violations, Violation{
					File:     file,
					Line:     line,
					Rule:     "isolation.cross-domain",
					Message:  fmt.Sprintf("domain %q must not import domain %q", src.Domain, imp.Domain),
					Fix:      fmt.Sprintf("use %s/ for cross-domain orchestration or move shared types to %s/", m.OrchestrationDir, m.SharedDir),
					Severity: cfg.Sev,
				})
				continue
			}

			// Rule 8: unclassified internal packages must not import domains
			// (kindUnclassified is handled explicitly here, not silently skipped).
			if src.Kind == kindUnclassified &&
				(imp.Kind == kindDomain || imp.Kind == kindDomainRoot) {
				file, line := findImportPosition(pkg, impPath, projectRoot)
				violations = append(violations, Violation{
					File:     file,
					Line:     line,
					Rule:     "isolation.stray-imports-domain",
					Message:  fmt.Sprintf("package %q must not import domain %q", pkg.PkgPath, imp.Domain),
					Fix:      fmt.Sprintf("move domain orchestration to internal/%s or app wiring to cmd/", m.OrchestrationDir),
					Severity: cfg.Sev,
				})
				continue
			}
		}
	}
	return violations
}

func isUnderInternalDir(pkgPath, internalPrefix, dir string) bool {
	if dir == "" {
		return false
	}
	rel := strings.TrimPrefix(pkgPath, internalPrefix)
	return rel == dir || strings.HasPrefix(rel, dir+"/")
}

func isOrchestrationHandlerWith(m Model, pkgPath, internalPrefix string) bool {
	rel := strings.TrimPrefix(pkgPath, internalPrefix)
	return strings.HasPrefix(rel, m.OrchestrationDir+"/handler")
}
