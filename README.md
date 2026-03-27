# go-arch-guard

[한국어](README.ko.md)

Architecture guardrails for Go projects via `go test`.

Define isolation, layer-direction, structure, naming, and blast-radius rules, then fail regular tests when the project shape drifts. No CLI to learn. No separate config format. Just Go tests.

## Opinionated Defaults

The rules, paths (`internal/domain/`, `internal/orchestration/`, `internal/pkg/`), sublayer names, and layer-direction matrix shipped with this library reflect **NamhaeSusan's conventions**. They are not universal Go best practices.

If you want to adopt `go-arch-guard` in your own project, treat the current ruleset as a **reference implementation** and adjust (or rewrite) rules to match your team's architecture.

## Why

Architecture usually decays through a few broad mistakes, not through deep theoretical violations:

- cross-domain imports
- hidden composition roots
- package placement drift
- naming that breaks the intended project shape

`go-arch-guard` catches those coarse mistakes early via static analysis. It does not try to model every semantic nuance inside Go packages, and if Go already rejects something by itself (such as import cycles), that is not a primary target here.

## Install

```bash
go get github.com/NamhaeSusan/go-arch-guard
```

## Claude Code Plugin

This repository can also be used as a Claude Code plugin marketplace.

Install it in Claude Code:

1. Add this repository as a marketplace.
2. Install the bundled `go-arch-guard` plugin.

```text
/plugin marketplace add NamhaeSusan/go-arch-guard
/plugin install go-arch-guard@go-arch-guard-marketplace
```

Validate the marketplace/plugin layout locally with:

```bash
claude plugin validate .
```

Plugin versioning is automated for skill updates on `main`. When `SKILL.md` or `plugins/go-arch-guard/skills/go-arch-guard/SKILL.md` changes, GitHub Actions bumps `plugins/go-arch-guard/.claude-plugin/plugin.json` using this repository rule:

- start at `0.0.1`
- patch bump by default
- `0.0.99 -> 0.1.0`
- `0.99.99` is rejected

## Quick Start

Create `architecture_test.go` in your project root:

```go
package myproject_test

import (
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/analyzer"
	"github.com/NamhaeSusan/go-arch-guard/report"
	"github.com/NamhaeSusan/go-arch-guard/rules"
)

func TestArchitecture(t *testing.T) {
	root := "."
	module := "github.com/yourmodule"

	pkgs, err := analyzer.Load(root, "internal/...", "cmd/...")
	if err != nil {
		// Load returns valid packages alongside the error when only some
		// packages fail (e.g. a single type error). Use t.Log so analysis
		// continues on the packages that did load successfully.
		t.Log(err)
	}
	if len(pkgs) == 0 {
		t.Fatalf("no packages loaded: %v", err)
	}

	t.Run("domain isolation", func(t *testing.T) {
		report.AssertNoViolations(t, rules.CheckDomainIsolation(pkgs, module, root))
	})
	t.Run("layer direction", func(t *testing.T) {
		report.AssertNoViolations(t, rules.CheckLayerDirection(pkgs, module, root))
	})
	t.Run("naming", func(t *testing.T) {
		report.AssertNoViolations(t, rules.CheckNaming(pkgs))
	})
	t.Run("structure", func(t *testing.T) {
		report.AssertNoViolations(t, rules.CheckStructure(root))
	})
	t.Run("blast radius", func(t *testing.T) {
		report.AssertNoViolations(t, rules.AnalyzeBlastRadius(pkgs, module, root))
	})
}
```

Run:

```bash
go test -run TestArchitecture -v
```

Sample output when violations exist:

```text
=== RUN   TestArchitecture/domain_isolation
    [ERROR] violation: domain "order" must not import domain "user" (file: internal/domain/order/app/service.go:5, rule: isolation.cross-domain, fix: use orchestration/ for cross-domain orchestration or move shared types to pkg/)
    found 1 architecture violation(s)
--- FAIL: TestArchitecture/domain_isolation
```

### Simplified Usage

Pass empty strings for `module` and `root` to auto-extract them from the loaded packages:

```go
t.Run("domain isolation", func(t *testing.T) {
	report.AssertNoViolations(t, rules.CheckDomainIsolation(pkgs, "", ""))
})
t.Run("layer direction", func(t *testing.T) {
	report.AssertNoViolations(t, rules.CheckLayerDirection(pkgs, "", ""))
})
```

If the module cannot be determined (e.g. packages loaded without module metadata), a `meta.no-matching-packages` warning is emitted.

## Target Architecture

`go-arch-guard` assumes a domain-centric vertical-slice layout.

At the `internal/` top level, only `domain/`, `orchestration/`, and `pkg/` are allowed. Additional top-level support packages are rejected.

```text
cmd/
`-- api/
    |-- main.go
    |-- wire.go
    `-- routes.go

internal/
|-- domain/
|   |-- order/
|   |   |-- alias.go
|   |   |-- app/
|   |   |   `-- service.go
|   |   |-- core/
|   |   |   |-- model/
|   |   |   |   `-- order.go
|   |   |   |-- repo/
|   |   |   |   `-- repository.go
|   |   |   `-- svc/
|   |   |       `-- order.go
|   |   |-- event/
|   |   |   `-- events.go
|   |   |-- handler/
|   |   |   `-- http/
|   |   |       `-- handler.go
|   |   `-- infra/
|   |       `-- persistence/
|   |           `-- store.go
|   `-- user/
|       `-- ...
|-- orchestration/
|   |-- handler/
|   |   `-- http/
|   |       `-- handler.go
|   `-- create_order.go
`-- pkg/
    |-- middleware/
    `-- transport/http/
```

### Domain Root

Each domain root package is the public import surface for that domain. The root must define `alias.go`, and the root may not contain additional non-test Go files.

Example:

```go
// internal/domain/order/alias.go
package order

import (
	"mymodule/internal/domain/order/app"
	orderhttp "mymodule/internal/domain/order/handler/http"
)

type Service = app.Service
type Handler = orderhttp.Handler
```

Outside code should import `internal/domain/order`, not `internal/domain/order/app` or `internal/domain/order/handler/http`.

### Domain Layers

Within a domain, the modeled sublayers are:

- `handler`
- `app`
- `core`
- `core/model`
- `core/repo`
- `core/svc`
- `event`
- `infra`

Unknown domain sublayers are rejected.

### Orchestration

`internal/orchestration` is the cross-domain coordination layer.

- It may import domain roots only, not domain sub-packages.
- For domain imports, orchestration must go through the domain root package (`alias.go`), not `app/`, `handler/`, or other domain sub-packages.
- It may import shared helpers in `internal/pkg/...` when needed.
- It may also import other non-domain internal packages when those packages exist.
- It is still a protected layer from the outside: `cmd/...` and `internal/orchestration/...` may depend on orchestration, but domains, `pkg`, and other internal packages may not.

In other words, `CheckDomainIsolation` restricts how orchestration reaches domains, not every non-domain internal dependency orchestration may use. Whether extra internal packages are allowed in the tree is checked separately by `CheckStructure`.

### Shared Packages

`internal/pkg` is for shared utilities.

- `pkg` may not import domains.
- `pkg` may not import orchestration.
- Inner domain layers (`core`, `core/model`, `core/repo`, `core/svc`, `event`) may not import `internal/pkg/...`.

## Rules

### Domain Isolation

`rules.CheckDomainIsolation(pkgs, module, root, opts...)`

Purpose:

- block cross-domain imports
- force external access through the domain root package
- keep orchestration and `pkg` from becoming hidden dependency shortcuts

| Rule | Meaning |
|------|---------|
| `isolation.cross-domain` | domain A must not import domain B |
| `isolation.cmd-deep-import` | `cmd/` must only import domain roots, not sub-packages |
| `isolation.orchestration-deep-import` | `orchestration/` must only import domain roots, not sub-packages |
| `isolation.pkg-imports-domain` | `pkg/` must not import any domain |
| `isolation.pkg-imports-orchestration` | `pkg/` must not import orchestration |
| `isolation.domain-imports-orchestration` | domains must not import orchestration |
| `isolation.internal-imports-orchestration` | non-cmd/orchestration packages must not import orchestration |
| `isolation.internal-imports-domain` | unregistered internal packages must not import domains |

Import matrix:

| from | to | allowed? |
|------|----|----------|
| same domain | same domain | Yes |
| anyone | `internal/pkg/...` | Yes |
| `orchestration/...` | domain root | Yes |
| `orchestration/...` | domain sub-package | No |
| `orchestration/...` | `internal/pkg/...` | Yes |
| `orchestration/...` | other non-domain internal package | Yes |
| `cmd/...` | `internal/orchestration/...` | Yes |
| `cmd/...` | domain root | Yes |
| `cmd/...` | domain sub-package | No |
| domain | `internal/orchestration/...` | No |
| `internal/pkg/...` | any domain | No |
| `internal/pkg/...` | `internal/orchestration/...` | No |
| other internal package | any domain | No |
| other internal package | `internal/orchestration/...` | No |
| domain A | domain B | No |

### Layer Direction

`rules.CheckLayerDirection(pkgs, module, root, opts...)`

Purpose:

- enforce allowed intra-domain dependency direction
- reject unknown domain sublayers
- keep inner layers free of `internal/pkg/...`

| Rule | Meaning |
|------|---------|
| `layer.direction` | import violates the allowed layer direction |
| `layer.inner-imports-pkg` | inner layer (`core/*`, `event`) imports `internal/pkg/` |
| `layer.unknown-sublayer` | unknown sublayer found in domain |

Allowed imports inside one domain:

| from | allowed to import |
|------|-------------------|
| `handler` | `app` |
| `app` | `core/model`, `core/repo`, `core/svc`, `event` |
| `core` | `core/model` |
| `core/model` | nothing |
| `core/repo` | `core/model` |
| `core/svc` | `core/model` |
| `event` | `core/model` |
| `infra` | `core/repo`, `core/model`, `event` |

Notes:

- same-sublayer imports are allowed
- the domain root package is not checked by `CheckLayerDirection`
- `core`, `core/model`, `core/repo`, `core/svc`, and `event` may not import `internal/pkg/...`

Examples:

```go
// OK: infra imports core/repo (allowed)
// in internal/domain/order/infra/persistence/store.go
import "mymodule/internal/domain/order/core/repo"

// BAD: core/svc imports core/repo (not in allowed list)
// in internal/domain/order/core/svc/order.go
import "mymodule/internal/domain/order/core/repo" // layer.direction

// BAD: handler imports infra directly (not in allowed list)
// in internal/domain/order/handler/http/handler.go
import "mymodule/internal/domain/order/infra/persistence" // layer.direction

// BAD: core/model imports internal/pkg (inner layer restriction)
// in internal/domain/order/core/model/order.go
import "mymodule/internal/pkg/clock" // layer.inner-imports-pkg
```

### Structure

`rules.CheckStructure(root, opts...)`

| Rule | Meaning |
|------|---------|
| `structure.internal-top-level` | only `domain`, `orchestration`, and `pkg` are allowed directly under `internal/` |
| `structure.banned-package` | `util`, `common`, `misc`, `helper`, `shared`, `services` are banned anywhere under `internal/` |
| `structure.legacy-package` | `router`, `bootstrap`, or misplaced `app`/`handler`/`infra` directories under `internal/` |
| `structure.middleware-placement` | `middleware/` must live at `internal/pkg/middleware/` |
| `structure.domain-root-alias-required` | each domain root must define `alias.go` |
| `structure.domain-root-alias-package` | `alias.go` package name must match the domain directory |
| `structure.domain-root-alias-only` | the domain root may contain only `alias.go` as a non-test Go file |
| `structure.domain-alias-no-interface` | `alias.go` must not re-export interfaces — suspected cross-domain dependency |
| `structure.domain-model-required` | each domain must have at least one direct non-test Go file in `core/model/` |
| `structure.dto-placement` | `dto.go` or `*_dto.go` must not live in inner domain layers (`core/`, `event/`) or `infra/`; allowed in `handler/` and `app/` |

### Naming

`rules.CheckNaming(pkgs, opts...)`

This rule set is intentionally more opinionated than the boundary rules.

| Rule | Meaning |
|------|---------|
| `naming.no-stutter` | exported type repeats the package name |
| `naming.no-impl-suffix` | exported type ends with `Impl` |
| `naming.snake-case-file` | file name is not snake_case |
| `naming.repo-file-interface` | a file under `repo/` does not define the matching interface |
| `naming.no-layer-suffix` | file name redundantly repeats the layer name |
| `naming.domain-interface-repo-only` | domain interface defined outside `core/repo/`, or type alias re-exporting `core/repo` interface |
| `naming.no-handmock` | test file defines a hand-rolled mock/fake/stub struct with methods |

### Blast Radius

`rules.AnalyzeBlastRadius(pkgs, module, root, opts...)`

Purpose:

- surface internal packages with abnormally high coupling
- zero configuration — uses IQR-based statistical outlier detection
- default severity is Warning (does not fail tests unless overridden)

| Rule | Meaning |
|------|---------|
| `blast-radius.high-coupling` | package has statistically outlying transitive dependents |

Metrics computed per internal package:

| Metric | Definition |
|--------|-----------|
| Ca (Afferent Coupling) | count of internal packages that import this package |
| Ce (Efferent Coupling) | count of internal packages this package imports |
| Instability | Ce / (Ca + Ce) — 0 = stable, 1 = unstable |
| Transitive Dependents | full reverse-reachable set via BFS |

Outlier detection uses Q3 + 1.5 × IQR on the Transitive Dependents distribution. Projects with fewer than 5 internal packages skip analysis.

## Options

### Severity

Default severity is `Error`. To log violations without failing the test, use `Warning`:

```go
violations := rules.CheckDomainIsolation(
	pkgs, module, root,
	rules.WithSeverity(rules.Warning),
)
report.AssertNoViolations(t, violations) // passes — only Error fails
```

### Exclude Paths

Skip subtrees during migration:

```go
rules.CheckDomainIsolation(
	pkgs, module, root,
	rules.WithExclude("internal/legacy/..."),
	rules.WithExclude("internal/domain/payment/core/model/..."),
)
```

- Patterns are project-relative paths with forward slashes
- `...` matches the exact root and all descendants
- Do not use module-qualified paths (`github.com/yourmodule/...`)
- The same format applies across all check functions

## Diagnostics

| Rule | Meaning |
|------|---------|
| `meta.no-matching-packages` | the `module` argument does not match any loaded package — usually a misconfiguration |

## TUI Viewer

Visualize your project's package structure and dependencies in an interactive terminal UI.

```bash
go run ./cmd/tui .
# or specify a project directory
go run ./cmd/tui /path/to/your/project
```

Features:
- Package tree with architectural layer color-coding (blue: `cmd/`, green: `domain/`, yellow: `orchestration/`, gray: `pkg/`)
- Select a package to see its imports and reverse dependencies
- Keyboard: `↑↓` navigate, `Enter` expand/collapse, `q` quit

## API Reference

| Function | Description |
|----------|-------------|
| `analyzer.Load(dir, patterns...)` | load Go packages for analysis |
| `rules.CheckDomainIsolation(pkgs, module, root, opts...)` | cross-domain and orchestration boundary checks (`""` auto-extracts from packages) |
| `rules.CheckLayerDirection(pkgs, module, root, opts...)` | intra-domain dependency direction checks (`""` auto-extracts from packages) |
| `rules.CheckNaming(pkgs, opts...)` | naming convention checks |
| `rules.CheckStructure(root, opts...)` | filesystem structure checks |
| `rules.AnalyzeBlastRadius(pkgs, module, root, opts...)` | coupling outlier detection (default Warning) (`""` auto-extracts from packages) |
| `report.AssertNoViolations(t, violations)` | fail test on `Error` violations |
| `rules.WithSeverity(rules.Warning)` | downgrade violations to warnings |
| `rules.WithExclude("internal/path/...")` | skip a project-relative subtree or file |

## External Import Hygiene — Enforce via AI Tool Instructions, Not This Library

`go-arch-guard` checks **project-internal** imports only. It does not and **will not** restrict which external packages a layer may use (e.g., `core/model` importing `gorm.io/gorm`).

This is intentional. A rule like `WithBannedImport("core/...", "gorm.io/...")` sounds simple but quickly becomes an allowlist maintenance burden that outweighs its value. External dependency hygiene is better enforced via AI tool instructions and code review.

**Copy the constraints below into your AI tool's system prompt or project rules file (e.g., `CLAUDE.md`, `AGENTS.md`, `.cursorrules`):**

```text
# External Import Constraints (go-arch-guard does NOT enforce these — you must)

- core/model, core/repo, core/svc, event — stdlib only, no third-party imports
- handler — HTTP/gRPC framework allowed, no persistence libraries
- infra — persistence/messaging libraries allowed, no HTTP framework imports
- app — generally free, but should not import infrastructure libraries directly
```

This is the intended enforcement mechanism. Do not open issues or PRs requesting `go-arch-guard` to add external import rules.

## License

MIT
