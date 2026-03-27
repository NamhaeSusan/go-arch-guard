# Blast Radius Guard Design

## Overview

Analyze the internal package dependency graph and report packages with abnormally high coupling as Warnings. Zero configuration required — uses statistical outlier detection that scales automatically with project size.

## Motivation

During vibe coding, it's easy to create hidden coupling hotspots — packages that half the project depends on. When these packages change, the blast radius is large and breakage is unpredictable. This guard surfaces those hotspots as Warnings so AI coding agents notice them.

## API

```go
func AnalyzeBlastRadius(
    pkgs []*packages.Package,
    projectModule string,    // "" for auto-detect
    projectRoot string,      // "" for auto-detect
    opts ...Option,          // WithExclude, WithSeverity
) []Violation
```

Same signature as all other Check functions. Returns `[]Violation` with Warning severity by default.

## Scope

- **Internal packages only** — `internal/domain/`, `internal/pkg/`, `internal/orchestration/`, `cmd/`
- External dependencies (stdlib, third-party) are excluded from the graph
- Consistent with all other go-arch-guard rules

## Metrics (Internal)

For each internal package, compute:

| Metric | Definition |
|--------|-----------|
| **Ca (Afferent Coupling)** | Count of internal packages that import this package |
| **Ce (Efferent Coupling)** | Count of internal packages this package imports |
| **Instability** | `Ce / (Ca + Ce)` — 0 = stable, 1 = unstable |
| **Direct Dependents** | Same as Ca |
| **Transitive Dependents** | Full reverse-reachable set via BFS on the inverted dependency graph |

## Outlier Detection

Uses **IQR (Interquartile Range)** method on Transitive Dependents distribution:

1. Collect Transitive Dependents count for all internal packages
2. Compute Q1 (25th percentile) and Q3 (75th percentile)
3. IQR = Q3 - Q1
4. Threshold = Q3 + 1.5 × IQR
5. Packages exceeding threshold → Warning violation

**Edge cases:**
- Fewer than 5 internal packages → skip analysis (not meaningful)
- IQR = 0 (all packages have same dependents) → no outliers possible, skip
- Threshold < 2 → floor at 2 (a package with 1 dependent is never a problem)

## Violation Output

```
Rule:     blast-radius.high-coupling
File:     internal/domain/order/core/model
Message:  package "core/model" has 20 transitive dependents (Ca:15 Ce:0 Instability:0.00, threshold:8)
Fix:      consider breaking this package into smaller units or introducing an interface boundary
Severity: Warning
```

Rule ID: `blast-radius.high-coupling`

## Implementation

### Files

| File | Purpose |
|------|---------|
| `rules/blast.go` | Graph construction, metrics, outlier detection, Violation generation |
| `rules/blast_test.go` | Unit tests with synthetic package graphs |

### Algorithm

```
1. Filter pkgs to internal packages only (using projectModule prefix)
2. Build adjacency list: pkg → set of internal imports
3. Build reverse adjacency list: pkg → set of internal importers
4. For each package:
   a. Ca = len(reverse[pkg])
   b. Ce = len(forward[pkg])
   c. Instability = Ce / (Ca + Ce), or 0.0 if Ca+Ce == 0
   d. Transitive Dependents = BFS on reverse graph from pkg
5. Compute IQR threshold on Transitive Dependents distribution
6. Emit Warning Violation for each package exceeding threshold
```

### Test Strategy

- **Dedicated testdata project** (`testdata/blast/`) — a real Go project with controlled topology:
  - `testdata/blast/go.mod`
  - Star topology: a shared `core/model` package imported by 10+ packages → outlier detected
  - Leaf packages with low coupling → no warning
  - Enough packages (>5) to make statistical analysis meaningful
- **Existing testdata** — run against `testdata/valid` and `testdata/invalid` to verify no panics and reasonable output
- **Edge cases** — fewer than 5 packages, zero imports

### Documentation Updates

- `README.md` — add `AnalyzeBlastRadius` to API reference
- `SKILL.md` — add blast radius rule description
