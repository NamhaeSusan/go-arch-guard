package tui

import (
	"strings"

	"golang.org/x/tools/go/packages"
)

// PkgMetrics holds coupling metrics for a single package.
type PkgMetrics struct {
	Ca                   int
	Ce                   int
	Instability          float64
	TransitiveDependents int
}

// MetricsIndex maps full package paths to their metrics.
type MetricsIndex map[string]*PkgMetrics

// BuildMetricsIndex computes coupling metrics for internal packages.
func BuildMetricsIndex(pkgs []*packages.Package, module string) MetricsIndex {
	internalPrefix := module + "/internal/"

	internalPkgs := make(map[string]bool)
	for _, pkg := range pkgs {
		if strings.HasPrefix(pkg.PkgPath, internalPrefix) {
			internalPkgs[pkg.PkgPath] = true
		}
	}

	// Forward adjacency.
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

	// Reverse adjacency.
	reverse := make(map[string]map[string]bool)
	for pkg := range internalPkgs {
		reverse[pkg] = make(map[string]bool)
	}
	for src, deps := range forward {
		for dep := range deps {
			reverse[dep][src] = true
		}
	}

	idx := make(MetricsIndex)
	for pkg := range internalPkgs {
		ca := len(reverse[pkg])
		ce := len(forward[pkg])
		var instability float64
		if ca+ce > 0 {
			instability = float64(ce) / float64(ca+ce)
		}

		// BFS for transitive dependents.
		visited := map[string]bool{pkg: true}
		queue := []string{pkg}
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

		idx[pkg] = &PkgMetrics{
			Ca:                   ca,
			Ce:                   ce,
			Instability:          instability,
			TransitiveDependents: len(visited) - 1,
		}
	}
	return idx
}
