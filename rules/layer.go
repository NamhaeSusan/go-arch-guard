package rules

import (
	"fmt"
	"strings"

	"golang.org/x/tools/go/packages"
)

var allowedLayerImports = map[string][]string{
	"":           {"app"},
	"handler":    {"app"},
	"app":        {"core/model", "core/repo", "core/svc"},
	"core/svc":   {"core/model"},
	"core/repo":  {"core/model"},
	"infra":      {"core/repo", "core/model"},
	"event":      {"core/model"},
	"core/model": {},
}

func CheckLayerDirection(pkgs []*packages.Package, projectModule string, projectRoot string, opts ...Option) []Violation {
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
		if srcDomain == "" {
			continue
		}

		srcSublayer := identifySublayer(pkg.PkgPath, internalPrefix, srcDomain)

		for impPath := range pkg.Imports {
			if !strings.HasPrefix(impPath, internalPrefix) {
				continue
			}

			// Skip imports to pkg/ or orchestration/
			if isPkgPkg(impPath, internalPrefix) || isOrchestrationPkg(impPath, internalPrefix) {
				continue
			}

			impDomain := identifyDomain(impPath, internalPrefix)
			// Only check intra-domain imports
			if impDomain != srcDomain {
				continue
			}

			impSublayer := identifySublayer(impPath, internalPrefix, impDomain)

			// Same sublayer → always allowed
			if srcSublayer == impSublayer {
				continue
			}

			// Check allowed table
			allowed, known := allowedLayerImports[srcSublayer]
			if !known {
				continue
			}

			isAllowed := false
			for _, a := range allowed {
				if impSublayer == a {
					isAllowed = true
					break
				}
			}
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
