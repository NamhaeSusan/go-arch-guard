package dependency

import (
	"fmt"
	"sort"
	"strings"

	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/core/analysisutil"
)

type BlastRadius struct{ severity core.Severity }

func NewBlastRadius(opts ...Option) *BlastRadius {
	cfg := newConfig(opts, core.Warning)
	return &BlastRadius{severity: cfg.severity}
}

func (r *BlastRadius) Spec() core.RuleSpec {
	return core.RuleSpec{
		ID:              "dependency.blast-radius",
		Description:     "internal packages with unusually many transitive dependents are coupling hotspots",
		DefaultSeverity: r.severity,
		Violations:      violationSpecs(r.severity, "blast.high-coupling"),
	}
}

func (r *BlastRadius) Check(ctx *core.Context) []core.Violation {
	projectModule := analysisutil.ResolveModuleFromContext(ctx, "")
	pkgs := ctx.Pkgs()
	if warns := validateModule(pkgs, projectModule); len(warns) > 0 {
		return warns
	}
	if !hasInternalPackages(pkgs, projectModule) {
		return []core.Violation{metaLayoutNotSupported("dependency.blast-radius", projectModule)}
	}

	internalPrefix := projectModule + "/internal/"
	internalPkgs := make(map[string]bool)
	for _, pkg := range pkgs {
		if strings.HasPrefix(pkg.PkgPath, internalPrefix) &&
			!isExcludedPackage(ctx, pkg.PkgPath, projectModule) {
			internalPkgs[pkg.PkgPath] = true
		}
	}
	if len(internalPkgs) < 5 {
		return nil
	}

	forward := make(map[string]map[string]bool)
	for _, pkg := range ctx.Pkgs() {
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

	reverse := make(map[string]map[string]bool)
	for pkg := range internalPkgs {
		reverse[pkg] = make(map[string]bool)
	}
	for src, deps := range forward {
		for dep := range deps {
			reverse[dep][src] = true
		}
	}

	metrics := collectBlastMetrics(internalPkgs, forward, reverse)
	tdValues := make([]int, len(metrics))
	for i, m := range metrics {
		tdValues[i] = m.transitiveDependents
	}
	sort.Ints(tdValues)

	threshold, ok := blastThreshold(tdValues)
	if !ok {
		return nil
	}

	var violations []core.Violation
	for _, m := range metrics {
		if float64(m.transitiveDependents) <= threshold {
			continue
		}
		relPath := analysisutil.ProjectRelativePackagePath(m.path, projectModule)
		violations = append(violations, core.Violation{
			File: relPath,
			Rule: "blast.high-coupling",
			Message: fmt.Sprintf(
				"package %q has %d transitive dependents (Ca:%d Ce:%d Instability:%.2f, threshold:%.0f)",
				relPath, m.transitiveDependents, m.ca, m.ce, m.instability, threshold,
			),
			Fix:               "consider breaking this package into smaller units or introducing an interface boundary",
			DefaultSeverity:   r.severity,
			EffectiveSeverity: r.severity,
		})
	}

	sort.Slice(violations, func(i, j int) bool {
		return violations[i].File < violations[j].File
	})
	return violations
}

type blastMetrics struct {
	path                 string
	ca                   int
	ce                   int
	instability          float64
	transitiveDependents int
}

func collectBlastMetrics(internalPkgs map[string]bool, forward, reverse map[string]map[string]bool) []blastMetrics {
	metrics := make([]blastMetrics, 0, len(internalPkgs))
	for pkg := range internalPkgs {
		ca := len(reverse[pkg])
		ce := len(forward[pkg])
		var instability float64
		if ca+ce > 0 {
			instability = float64(ce) / float64(ca+ce)
		}

		visited := map[string]bool{pkg: true}
		queue := []string{pkg}
		for len(queue) > 0 {
			cur := queue[0]
			queue = queue[1:]
			for dep := range reverse[cur] {
				if visited[dep] {
					continue
				}
				visited[dep] = true
				queue = append(queue, dep)
			}
		}

		metrics = append(metrics, blastMetrics{
			path:                 pkg,
			ca:                   ca,
			ce:                   ce,
			instability:          instability,
			transitiveDependents: len(visited) - 1,
		})
	}
	return metrics
}

func blastThreshold(sorted []int) (float64, bool) {
	if len(sorted) == 0 {
		return 0, false
	}
	q1 := percentile(sorted, 25)
	q3 := percentile(sorted, 75)
	iqr := q3 - q1

	const iqrZeroFloor = 3
	if iqr == 0 {
		max := float64(sorted[len(sorted)-1])
		if max <= q3 || max < iqrZeroFloor {
			return 0, false
		}
		return q3, true
	}

	threshold := q3 + 1.5*iqr
	if threshold < 2 {
		threshold = 2
	}
	return threshold, true
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
