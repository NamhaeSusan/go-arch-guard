package rules

import (
	"fmt"
	"slices"
	"strings"

	"golang.org/x/tools/go/packages"
)

func CheckLayerDirection(pkgs []*packages.Package, projectModule string, projectRoot string, opts ...Option) []Violation {
	cfg := NewConfig(opts...)
	m := cfg.model()
	projectModule = resolveModule(pkgs, projectModule)
	projectRoot = resolveRoot(pkgs, projectRoot)
	if warns := validateModule(pkgs, projectModule); len(warns) > 0 {
		return warns
	}
	internalPrefix := projectModule + "/internal/"

	if m.DomainDir == "" {
		return checkFlatLayerDirection(pkgs, m, cfg, projectModule, projectRoot, internalPrefix)
	}

	var violations []Violation
	for _, pkg := range pkgs {
		if isExcludedPackage(cfg, pkg.PkgPath, projectModule) {
			continue
		}
		if !strings.HasPrefix(pkg.PkgPath, internalPrefix) {
			continue
		}

		srcDomain := identifyDomainWith(m, pkg.PkgPath, internalPrefix)
		if srcDomain == "" {
			continue
		}

		srcSublayer := identifySublayerWith(m, pkg.PkgPath, internalPrefix, srcDomain)
		if srcSublayer != "" && !isKnownSublayerIn(m, srcSublayer) {
			violations = append(violations, Violation{
				File:     relativePackageFile(pkg),
				Rule:     "layer.unknown-sublayer",
				Message:  fmt.Sprintf("unknown sublayer %q in domain %q", srcSublayer, srcDomain),
				Fix:      fmt.Sprintf("use one of the supported sublayers: %v", m.Sublayers),
				Severity: cfg.Sev,
			})
			continue
		}

		for impPath := range pkg.Imports {
			if !strings.HasPrefix(impPath, internalPrefix) {
				continue
			}

			if isPkgPkgWith(m, impPath, internalPrefix) {
				if m.PkgRestricted[srcSublayer] {
					file, line := findImportPosition(pkg, impPath, projectRoot)
					violations = append(violations, Violation{
						File:     file,
						Line:     line,
						Rule:     "layer.inner-imports-pkg",
						Message:  fmt.Sprintf("inner sublayer %q must not import internal/%s in domain %q", srcSublayer, m.SharedDir, srcDomain),
						Fix:      "keep core and event layers self-contained; move shared concerns outward to app, handler, or infra",
						Severity: cfg.Sev,
					})
				}
				continue
			}

			// Skip imports to orchestration/
			if isOrchestrationPkgWith(m, impPath, internalPrefix) {
				continue
			}

			impDomain := identifyDomainWith(m, impPath, internalPrefix)
			// Only check intra-domain imports
			if impDomain != srcDomain {
				continue
			}

			impSublayer := identifySublayerWith(m, impPath, internalPrefix, impDomain)

			// Domain root (alias.go) is a facade — it may import any sublayer
			// within its own domain. Cross-domain isolation is enforced elsewhere.
			if srcSublayer == "" {
				continue
			}

			if impSublayer != "" && !isKnownSublayerIn(m, impSublayer) {
				file, line := findImportPosition(pkg, impPath, projectRoot)
				violations = append(violations, Violation{
					File:     file,
					Line:     line,
					Rule:     "layer.unknown-sublayer",
					Message:  fmt.Sprintf("unknown sublayer %q in domain %q", impSublayer, srcDomain),
					Fix:      fmt.Sprintf("use one of the supported sublayers: %v", m.Sublayers),
					Severity: cfg.Sev,
				})
				continue
			}

			// Same sublayer → always allowed
			if srcSublayer == impSublayer {
				continue
			}

			// Check allowed table
			allowed, known := m.Direction[srcSublayer]
			if !known {
				continue
			}

			isAllowed := slices.Contains(allowed, impSublayer)
			if isAllowed {
				continue
			}

			file, line := findImportPosition(pkg, impPath, projectRoot)
			violations = append(violations, Violation{
				File:     file,
				Line:     line,
				Rule:     "layer.direction",
				Message:  fmt.Sprintf("sublayer %q must not import sublayer %q in domain %q", srcSublayer, impSublayer, srcDomain),
				Fix:      fmt.Sprintf("allowed imports for %q: %v", srcSublayer, allowed),
				Severity: cfg.Sev,
			})
		}
	}
	return violations
}

func identifySublayerWith(m Model, pkgPath, internalPrefix, domain string) string {
	domainPrefix := internalPrefix + m.DomainDir + "/" + domain + "/"
	if !strings.HasPrefix(pkgPath, domainPrefix) {
		return "" // domain root package
	}
	rel := strings.TrimPrefix(pkgPath, domainPrefix)
	parts := strings.SplitN(rel, "/", 3)

	// Check if this is a nested sublayer (e.g. "core/model")
	if len(parts) >= 2 {
		nested := parts[0] + "/" + parts[1]
		if slices.Contains(m.Sublayers, nested) {
			return nested
		}
		// If parts[0] is a root sublayer that has known nested sublayers defined
		// (e.g. "core" has "core/model", "core/repo"), then an unrecognised nested
		// path like "core/extra" must surface as unknown rather than collapsing to
		// the root "core". Sublayers without nested siblings (e.g. "handler",
		// "infra") keep their old behaviour of ignoring subdirectories.
		if hasNestedSublayers(m, parts[0]) {
			return nested
		}
	}
	return parts[0]
}

// hasNestedSublayers reports whether root has at least one known nested sublayer
// defined (e.g. "core" has "core/model", "core/repo"). Sublayers that only
// appear as a root entry (e.g. "handler", "infra") return false.
func hasNestedSublayers(m Model, root string) bool {
	prefix := root + "/"
	for _, s := range m.Sublayers {
		if strings.HasPrefix(s, prefix) {
			return true
		}
	}
	return false
}

func identifyDomainWith(m Model, pkgPath, internalPrefix string) string {
	rel := strings.TrimPrefix(pkgPath, internalPrefix)
	if !strings.HasPrefix(rel, m.DomainDir+"/") {
		return ""
	}
	after := strings.TrimPrefix(rel, m.DomainDir+"/")
	parts := strings.SplitN(after, "/", 2)
	if parts[0] == "" {
		return ""
	}
	return parts[0]
}

func checkFlatLayerDirection(pkgs []*packages.Package, m Model, cfg Config, projectModule, projectRoot, internalPrefix string) []Violation {
	var violations []Violation
	for _, pkg := range pkgs {
		if isExcludedPackage(cfg, pkg.PkgPath, projectModule) {
			continue
		}
		if !strings.HasPrefix(pkg.PkgPath, internalPrefix) {
			continue
		}

		srcSublayer := identifyFlatSublayer(m, pkg.PkgPath, internalPrefix)
		if srcSublayer == "" {
			continue
		}

		for impPath := range pkg.Imports {
			if !strings.HasPrefix(impPath, internalPrefix) {
				continue
			}

			if isPkgPkgWith(m, impPath, internalPrefix) {
				if m.PkgRestricted[srcSublayer] {
					file, line := findImportPosition(pkg, impPath, projectRoot)
					violations = append(violations, Violation{
						File:     file,
						Line:     line,
						Rule:     "layer.inner-imports-pkg",
						Message:  fmt.Sprintf("inner sublayer %q must not import internal/%s", srcSublayer, m.SharedDir),
						Fix:      "keep inner layers self-contained; move shared concerns to an outer layer",
						Severity: cfg.Sev,
					})
				}
				continue
			}

			impSublayer := identifyFlatSublayer(m, impPath, internalPrefix)
			if impSublayer == "" || srcSublayer == impSublayer {
				continue
			}

			allowed, known := m.Direction[srcSublayer]
			if !known {
				continue
			}
			if slices.Contains(allowed, impSublayer) {
				continue
			}

			file, line := findImportPosition(pkg, impPath, projectRoot)
			violations = append(violations, Violation{
				File:     file,
				Line:     line,
				Rule:     "layer.direction",
				Message:  fmt.Sprintf("sublayer %q must not import sublayer %q", srcSublayer, impSublayer),
				Fix:      fmt.Sprintf("allowed imports for %q: %v", srcSublayer, allowed),
				Severity: cfg.Sev,
			})
		}
	}
	return violations
}

func identifyFlatSublayer(m Model, pkgPath, internalPrefix string) string {
	rel := strings.TrimPrefix(pkgPath, internalPrefix)
	parts := strings.SplitN(rel, "/", 2)
	if parts[0] == "" {
		return ""
	}
	if slices.Contains(m.Sublayers, parts[0]) {
		return parts[0]
	}
	return ""
}

func relativePackageFile(pkg *packages.Package) string {
	if len(pkg.GoFiles) == 0 {
		return pkg.PkgPath
	}
	return relativePathForPackage(pkg, pkg.GoFiles[0])
}
