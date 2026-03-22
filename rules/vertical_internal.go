package rules

import (
	"strings"

	"golang.org/x/tools/go/packages"
)

var allowedInternalImports = map[string]map[string]bool{
	"handler": {"app": true, "port": true},
	"app":     {"domain": true, "policy": true, "model": true, "repo": true, "event": true, "port": true},
	"domain":  {"model": true, "event": true},
	"policy":  {"model": true, "event": true},
	"infra":   {"model": true, "repo": true, "event": true, "port": true},
	"model":   {"event": true},
	"repo":    {"model": true},
	"event":   {},
	"port":    {"model": true},
}

func CheckVerticalSliceInternal(pkgs []*packages.Package, projectModule string, projectRoot string, opts ...Option) []Violation {
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
		if srcDomain == "" || srcDomain == "shared" {
			continue
		}

		srcSublayer := identifySublayer(pkgPath, internalPrefix, srcDomain)
		if srcSublayer == "" {
			continue
		}

		for importPath := range pkg.Imports {
			if !strings.HasPrefix(importPath, internalPrefix) {
				continue
			}

			dstDomain := identifyVerticalDomain(importPath, internalPrefix)
			if dstDomain == "" || dstDomain == "shared" {
				continue
			}

			// Skip cross-domain imports (CheckVerticalSlice handles those)
			if srcDomain != dstDomain {
				continue
			}

			dstSublayer := identifySublayer(importPath, internalPrefix, dstDomain)
			if dstSublayer == "" {
				continue
			}

			// Same sublayer is always allowed
			if srcSublayer == dstSublayer {
				continue
			}

			allowed, exists := allowedInternalImports[srcSublayer]
			if !exists {
				continue
			}

			if !allowed[dstSublayer] {
				violations = append(violations, Violation{
					File:     findImportFile(pkg, importPath, projectRoot),
					Line:     findImportLine(pkg, importPath),
					Rule:     "vertical.internal-layer-direction",
					Message:  `sublayer "` + srcSublayer + `" imports "` + dstSublayer + `" in domain "` + srcDomain + `"`,
					Fix:      "check allowed layer direction: " + srcSublayer + " cannot import " + dstSublayer,
					Severity: cfg.Sev,
				})
			}
		}
	}
	return violations
}

func identifySublayer(pkgPath, internalPrefix, domain string) string {
	domainPrefix := internalPrefix + domain + "/"
	if !strings.HasPrefix(pkgPath, domainPrefix) {
		// pkgPath is exactly the domain root (alias package)
		return ""
	}
	rel := strings.TrimPrefix(pkgPath, domainPrefix)
	parts := strings.SplitN(rel, "/", 2)
	if len(parts) == 0 || parts[0] == "" {
		return ""
	}
	return parts[0]
}
