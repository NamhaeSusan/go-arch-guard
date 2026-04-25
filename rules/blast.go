package rules

import (
	"fmt"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"
)

// AnalyzeBlastRadius computes coupling metrics for internal packages and emits
// Warning violations for packages whose transitive dependents count is a
// statistical outlier (IQR method). No configuration required.
func AnalyzeBlastRadius(pkgs []*packages.Package, projectModule string, projectRoot string, opts ...Option) []Violation {
	cfg := NewConfig(append([]Option{WithSeverity(Warning)}, opts...)...)
	projectModule = resolveModule(pkgs, projectModule)
	if warns := validateModule(pkgs, projectModule); len(warns) > 0 {
		return warns
	}

	internalPrefix := projectModule + "/internal/"

	// 1. Collect internal packages
	internalPkgs := make(map[string]bool)
	for _, pkg := range pkgs {
		if strings.HasPrefix(pkg.PkgPath, internalPrefix) {
			if !isExcludedPackage(cfg, pkg.PkgPath, projectModule) {
				internalPkgs[pkg.PkgPath] = true
			}
		}
	}

	if len(internalPkgs) < 5 {
		return nil
	}

	// 2. Build forward adjacency (pkg -> internal imports)
	forward := make(map[string]map[string]bool)
	for _, pkg := range pkgs {
		if !internalPkgs[pkg.PkgPath] {
			continue
		}
		deps := make(map[string]bool)
		for impPath := range pkg.Imports {
			if internalPkgs[impPath] {
				deps[impPath] = true
			}
		}
		forward[pkg.PkgPath] = deps
	}

	// 3. Build reverse adjacency (pkg -> internal importers)
	reverse := make(map[string]map[string]bool)
	for pkg := range internalPkgs {
		reverse[pkg] = make(map[string]bool)
	}
	for src, deps := range forward {
		for dep := range deps {
			reverse[dep][src] = true
		}
	}

	// 4. Compute metrics per package
	type pkgMetrics struct {
		path                 string
		ca                   int
		ce                   int
		instability          float64
		transitiveDependents int
	}

	metrics := make([]pkgMetrics, 0, len(internalPkgs))
	for pkg := range internalPkgs {
		ca := len(reverse[pkg])
		ce := len(forward[pkg])
		var instability float64
		if ca+ce > 0 {
			instability = float64(ce) / float64(ca+ce)
		}

		// BFS for transitive dependents
		visited := make(map[string]bool)
		queue := []string{pkg}
		visited[pkg] = true
		for len(queue) > 0 {
			cur := queue[0]
			queue = queue[1:]
			for dep := range reverse[cur] {
				if !visited[dep] {
					visited[dep] = true
					queue = append(queue, dep)
				}
			}
		}
		td := len(visited) - 1 // exclude self

		metrics = append(metrics, pkgMetrics{
			path:                 pkg,
			ca:                   ca,
			ce:                   ce,
			instability:          instability,
			transitiveDependents: td,
		})
	}

	// 5. IQR outlier detection on transitive dependents
	tdValues := make([]int, len(metrics))
	for i, m := range metrics {
		tdValues[i] = m.transitiveDependents
	}
	sort.Ints(tdValues)

	q1 := percentile(tdValues, 25)
	q3 := percentile(tdValues, 75)
	iqr := q3 - q1

	const iqrZeroFloor = 3 // minimum transitive dependents to flag when IQR=0

	var threshold float64
	if iqr == 0 {
		// All packages have the same td value. The IQR method can't detect
		// outliers, but a single package with a much higher count is still a
		// hotspot. Fall back: emit for the max value only if it strictly exceeds
		// q3 and meets the floor.
		max := float64(tdValues[len(tdValues)-1])
		if max <= q3 || max < iqrZeroFloor {
			return nil
		}
		threshold = q3 // anything > q3 (i.e. == max) is the outlier
	} else {
		threshold = q3 + 1.5*iqr
		if threshold < 2 {
			threshold = 2
		}
	}

	// 6. Emit violations
	var violations []Violation
	for _, m := range metrics {
		if float64(m.transitiveDependents) > threshold {
			relPath := projectRelativePackagePath(m.path, projectModule)
			violations = append(violations, Violation{
				File: relPath,
				Rule: "blast.high-coupling",
				Message: fmt.Sprintf(
					"package %q has %d transitive dependents (Ca:%d Ce:%d Instability:%.2f, threshold:%.0f)",
					relPath, m.transitiveDependents, m.ca, m.ce, m.instability, threshold,
				),
				Fix:               "consider breaking this package into smaller units or introducing an interface boundary",
				DefaultSeverity:   cfg.Sev,
				EffectiveSeverity: cfg.Sev,
			})
		}
	}

	sort.Slice(violations, func(i, j int) bool {
		return violations[i].File < violations[j].File
	})

	return violations
}

func percentile(sorted []int, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	k := (p / 100) * float64(len(sorted)-1)
	f := int(k)
	c := f + 1
	if c >= len(sorted) {
		return float64(sorted[f])
	}
	d := k - float64(f)
	return float64(sorted[f])*(1-d) + float64(sorted[c])*d
}
