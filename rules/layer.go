package rules

import (
	"fmt"
	"slices"
	"strings"

	"golang.org/x/tools/go/packages"
)

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
		if srcSublayer != "" && !isKnownSublayer(srcSublayer) {
			violations = append(violations, Violation{
				File:     relativePackageFile(pkg),
				Rule:     "layer.unknown-sublayer",
				Message:  fmt.Sprintf("unknown sublayer %q in domain %q", srcSublayer, srcDomain),
				Fix:      fmt.Sprintf("use one of the supported sublayers: %v", knownDomainSublayers),
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

			if impSublayer != "" && !isKnownSublayer(impSublayer) {
				violations = append(violations, Violation{
					File:     findImportFile(pkg, impPath, projectRoot),
					Line:     findImportLine(pkg, impPath),
					Rule:     "layer.unknown-sublayer",
					Message:  fmt.Sprintf("unknown sublayer %q in domain %q", impSublayer, srcDomain),
					Fix:      fmt.Sprintf("use one of the supported sublayers: %v", knownDomainSublayers),
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

func relativePackageFile(pkg *packages.Package) string {
	if len(pkg.GoFiles) == 0 {
		return pkg.PkgPath
	}
	return relativePathForPackage(pkg, pkg.GoFiles[0])
}
