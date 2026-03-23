package rules

import (
	"fmt"
	"slices"
	"strings"

	"golang.org/x/tools/go/packages"
)

var allowedLayerImports = map[string][]string{
	// Domain root ("") is a facade (alias.go) that re-exports from all sublayers.
	// It is exempt from layer direction checks; cross-domain isolation is enforced
	// separately by CheckDomainIsolation.
	"handler":    {"app"},
	"app":        {"core/model", "core/repo", "core/svc", "event"},
	"core/svc":   {"core/model"},
	"core/repo":  {"core/model"},
	"infra":      {"core/repo", "core/model", "event"},
	"event":      {"core/model"},
	"core/model": {},
	"core":       {"core/model"},
}

var knownSublayers = map[string]bool{
	"handler":    true,
	"app":        true,
	"core/svc":   true,
	"core/repo":  true,
	"infra":      true,
	"event":      true,
	"core/model": true,
	"core":       true,
}

var pkgRestrictedSublayers = map[string]bool{
	"core":       true,
	"core/model": true,
	"core/repo":  true,
	"core/svc":   true,
	"event":      true,
}

func CheckLayerDirection(pkgs []*packages.Package, projectModule string, projectRoot string, opts ...Option) []Violation {
	cfg := NewConfig(opts...)
	internalPrefix := projectModule + "/internal/"

	var violations []Violation
	for _, pkg := range pkgs {
		if isExcludedPackage(cfg, pkg.PkgPath, projectModule) {
			continue
		}
		if !strings.HasPrefix(pkg.PkgPath, internalPrefix) {
			continue
		}

		srcDomain := identifyDomain(pkg.PkgPath, internalPrefix)
		if srcDomain == "" {
			continue
		}

		srcSublayer := identifySublayer(pkg.PkgPath, internalPrefix, srcDomain)
		if srcSublayer != "" && !knownSublayers[srcSublayer] {
			violations = append(violations, Violation{
				File:     relativePackageFile(pkg),
				Rule:     "layer.unknown-sublayer",
				Message:  fmt.Sprintf("unknown sublayer %q in domain %q", srcSublayer, srcDomain),
				Fix:      fmt.Sprintf("use one of the supported sublayers: %v", knownSublayerList()),
				Severity: cfg.Sev,
			})
			continue
		}

		for impPath := range pkg.Imports {
			if !strings.HasPrefix(impPath, internalPrefix) {
				continue
			}

			if isPkgPkg(impPath, internalPrefix) {
				if pkgRestrictedSublayers[srcSublayer] {
					violations = append(violations, Violation{
						File:     findImportFile(pkg, impPath, projectRoot),
						Line:     findImportLine(pkg, impPath),
						Rule:     "layer.inner-imports-pkg",
						Message:  fmt.Sprintf("inner sublayer %q must not import internal/pkg in domain %q", srcSublayer, srcDomain),
						Fix:      "keep core and event layers self-contained; move shared concerns outward to app, handler, or infra",
						Severity: cfg.Sev,
					})
				}
				continue
			}

			// Skip imports to orchestration/
			if isOrchestrationPkg(impPath, internalPrefix) {
				continue
			}

			impDomain := identifyDomain(impPath, internalPrefix)
			// Only check intra-domain imports
			if impDomain != srcDomain {
				continue
			}

			impSublayer := identifySublayer(impPath, internalPrefix, impDomain)

			// Domain root (alias.go) is a facade — it may import any sublayer
			// within its own domain. Cross-domain isolation is enforced elsewhere.
			if srcSublayer == "" {
				continue
			}

			if impSublayer != "" && !knownSublayers[impSublayer] {
				violations = append(violations, Violation{
					File:     findImportFile(pkg, impPath, projectRoot),
					Line:     findImportLine(pkg, impPath),
					Rule:     "layer.unknown-sublayer",
					Message:  fmt.Sprintf("unknown sublayer %q in domain %q", impSublayer, srcDomain),
					Fix:      fmt.Sprintf("use one of the supported sublayers: %v", knownSublayerList()),
					Severity: cfg.Sev,
				})
				continue
			}

			// Same sublayer → always allowed
			if srcSublayer == impSublayer {
				continue
			}

			// Check allowed table
			allowed, known := allowedLayerImports[srcSublayer]
			if !known {
				continue
			}

			isAllowed := slices.Contains(allowed, impSublayer)
			if isAllowed {
				continue
			}

			violations = append(violations, Violation{
				File:     findImportFile(pkg, impPath, projectRoot),
				Line:     findImportLine(pkg, impPath),
				Rule:     "layer.direction",
				Message:  fmt.Sprintf("sublayer %q must not import sublayer %q in domain %q", srcSublayer, impSublayer, srcDomain),
				Fix:      fmt.Sprintf("allowed imports for %q: %v", srcSublayer, allowed),
				Severity: cfg.Sev,
			})
		}
	}
	return violations
}

func identifySublayer(pkgPath, internalPrefix, domain string) string {
	domainPrefix := internalPrefix + "domain/" + domain + "/"
	if !strings.HasPrefix(pkgPath, domainPrefix) {
		return "" // domain root package (alias.go)
	}
	rel := strings.TrimPrefix(pkgPath, domainPrefix)
	parts := strings.SplitN(rel, "/", 3)

	if parts[0] == "core" && len(parts) >= 2 {
		return "core/" + parts[1]
	}
	return parts[0]
}

func knownSublayerList() []string {
	return []string{"handler", "app", "core", "core/model", "core/repo", "core/svc", "event", "infra"}
}

func relativePackageFile(pkg *packages.Package) string {
	if len(pkg.GoFiles) == 0 {
		return pkg.PkgPath
	}
	return relativePathForPackage(pkg, pkg.GoFiles[0])
}
