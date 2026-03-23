# go-arch-guard

Architecture guardrails for Go projects via `go test`.

Define isolation, layer-direction, structure, and naming rules, then fail regular tests when the project shape drifts. No CLI to learn. No separate config format. Just Go tests.

## Why

Architecture usually decays through a few broad mistakes, not through deep theoretical violations:

- cross-domain imports
- hidden composition roots
- package placement drift
- naming that breaks the intended project shape

`go-arch-guard` is meant to catch those coarse mistakes early.

## Scope

This library is intentionally about broad project-shape guardrails, not purity proofs.

- It focuses on static analysis only.
- It does not try to model every semantic nuance inside Go packages.
- It prefers low-surprise checks over doctrinal rules.

If Go already rejects something by itself, such as import cycles, that is not a primary target here.

## Target Architecture

`go-arch-guard` assumes a domain-centric vertical-slice layout.

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
- It may import other non-domain internal support packages when needed.
- It is still a protected layer from the outside: `cmd/...` and `internal/orchestration/...` may depend on orchestration, but domains, `pkg`, and other internal packages may not.

In other words, orchestration is restricted on the domain boundary, not forced into total isolation from every internal helper package.

### Shared Packages

`internal/pkg` is for shared utilities.

- `pkg` may not import domains.
- `pkg` may not import orchestration.
- Inner domain layers (`core`, `core/model`, `core/repo`, `core/svc`, `event`) may not import `internal/pkg/...`.

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
		t.Fatal(err)
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
}
```

Run:

```bash
go test -run TestArchitecture -v
```

## Rules

### Domain Isolation

`rules.CheckDomainIsolation(pkgs, module, root, opts...)`

Purpose:

- block cross-domain imports
- force external access through the domain root package
- keep orchestration and `pkg` from becoming hidden dependency shortcuts

Import matrix:

| from | to | allowed? |
|------|----|----------|
| same domain | same domain | Yes |
| anyone | `internal/pkg/...` | Yes |
| `orchestration/...` | domain root | Yes |
| `orchestration/...` | domain sub-package | No |
| `orchestration/...` | other non-domain internal packages | Yes |
| `cmd/...` | `internal/orchestration/...` | Yes |
| `cmd/...` | domain root | Yes |
| `cmd/...` | domain sub-package | No |
| domain | `internal/orchestration/...` | No |
| `internal/pkg/...` | any domain | No |
| `internal/pkg/...` | `internal/orchestration/...` | No |
| other internal package | any domain | No |
| other internal package | `internal/orchestration/...` | No |
| domain A | domain B | No |

Examples:

```go
// OK: orchestration imports domain roots
import "mymodule/internal/domain/order"
import "mymodule/internal/domain/user"

// OK: orchestration imports an internal helper package
import "mymodule/internal/pkg/clock"

// BAD: orchestration deep-imports a domain
import "mymodule/internal/domain/user/core/model" // isolation.orchestration-deep-import

// BAD: shared package imports orchestration
import "mymodule/internal/orchestration" // isolation.pkg-imports-orchestration

// BAD: domain imports orchestration
import "mymodule/internal/orchestration" // isolation.domain-imports-orchestration

// BAD: config imports a domain directly
import "mymodule/internal/domain/user" // isolation.internal-imports-domain
```

### Layer Direction

`rules.CheckLayerDirection(pkgs, module, root, opts...)`

Purpose:

- enforce allowed intra-domain dependency direction
- reject unknown domain sublayers
- keep inner layers free of `internal/pkg/...`

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
// OK
import "mymodule/internal/domain/order/core/model"
import "mymodule/internal/domain/order/core/repo"

// BAD: svc should not know repo
import "mymodule/internal/domain/order/core/repo" // layer.direction

// BAD: handler should not talk to infra directly
import "mymodule/internal/domain/order/infra/persistence" // layer.direction

// BAD: inner layer imports pkg
import "mymodule/internal/pkg/clock" // layer.inner-imports-pkg
```

### Structure

`rules.CheckStructure(root, opts...)`

Filesystem checks:

| Rule | Meaning |
|------|---------|
| `structure.banned-package` | `util`, `common`, `misc`, `helper`, `shared` are banned anywhere under `internal/` |
| `structure.legacy-package` | `router`, `bootstrap`, or misplaced `app`/`handler`/`infra` directories under `internal/` |
| `structure.middleware-placement` | `middleware/` must live under `internal/pkg/` |
| `structure.domain-root-alias-required` | each domain root must define `alias.go` |
| `structure.domain-root-alias-package` | `alias.go` package name must match the domain directory |
| `structure.domain-root-alias-only` | the domain root may contain only `alias.go` as a non-test Go file |
| `structure.domain-model-required` | each domain must have a non-empty `core/model/` |
| `structure.dto-placement` | `dto.go` or `*_dto.go` must not live under `domain/` or `infra/` |

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
| `naming.handler-no-exported-interface` | handler package defines an exported interface |

## Paths, Excludes, and Output

### Root

- `root` is a filesystem path passed to `analyzer.Load`, `CheckDomainIsolation`, `CheckLayerDirection`, and `CheckStructure`.
- It is usually `"."` in normal project usage.

### Exclude Paths

Exclude patterns are matched against project-relative paths only.

Use:

```go
rules.WithExclude("internal/legacy/...")
rules.WithExclude("internal/domain/payment/core/model/...")
```

Do not use module-qualified paths such as:

```go
rules.WithExclude("github.com/yourmodule/internal/legacy/...")
```

Rules:

- `...` matches the exact root and all descendants
- use forward-slash project-relative paths
- the same exclude format applies across isolation, layer, naming, and structure checks

### Violation File Paths

Violation `File` values are reported as project-relative paths.

## Severity

Default severity is `Error`.

```go
violations := rules.CheckDomainIsolation(pkgs, module, root)
```

To log violations without failing the test:

```go
violations := rules.CheckDomainIsolation(
	pkgs,
	module,
	root,
	rules.WithSeverity(rules.Warning),
)

report.AssertNoViolations(t, violations)
```

`report.AssertNoViolations` fails only when at least one violation has `Error` severity.

## Gradual Adoption

Start with warnings:

```go
violations := rules.CheckDomainIsolation(
	pkgs,
	module,
	root,
	rules.WithSeverity(rules.Warning),
)
report.AssertNoViolations(t, violations)
```

Exclude legacy subtrees while migrating:

```go
violations := rules.CheckDomainIsolation(
	pkgs,
	module,
	root,
	rules.WithExclude("internal/legacy/..."),
)
```

Then remove excludes as the project converges on the target shape.

## API Reference

| Function | Description |
|----------|-------------|
| `analyzer.Load(dir, patterns...)` | load Go packages for analysis |
| `rules.CheckDomainIsolation(pkgs, module, root, opts...)` | cross-domain and orchestration boundary checks |
| `rules.CheckLayerDirection(pkgs, module, root, opts...)` | intra-domain dependency direction checks |
| `rules.CheckNaming(pkgs, opts...)` | naming convention checks |
| `rules.CheckStructure(root, opts...)` | filesystem structure checks |
| `report.AssertNoViolations(t, violations)` | fail test on `Error` violations |
| `rules.WithSeverity(rules.Warning)` | downgrade violations to warnings |
| `rules.WithExclude("internal/path/...")` | skip a project-relative subtree or file |

## Violation Reference

| Rule | Function | Description |
|------|----------|-------------|
| `isolation.cross-domain` | `CheckDomainIsolation` | one domain imports another domain |
| `isolation.internal-imports-domain` | `CheckDomainIsolation` | unsupported internal package imports a domain |
| `isolation.orchestration-deep-import` | `CheckDomainIsolation` | orchestration imports a domain sub-package |
| `isolation.cmd-deep-import` | `CheckDomainIsolation` | `cmd/...` imports a domain sub-package |
| `isolation.pkg-imports-domain` | `CheckDomainIsolation` | `internal/pkg/...` imports a domain |
| `isolation.pkg-imports-orchestration` | `CheckDomainIsolation` | `internal/pkg/...` imports orchestration |
| `isolation.domain-imports-orchestration` | `CheckDomainIsolation` | a domain imports orchestration |
| `isolation.internal-imports-orchestration` | `CheckDomainIsolation` | unsupported internal package imports orchestration |
| `layer.direction` | `CheckLayerDirection` | disallowed layer dependency inside one domain |
| `layer.inner-imports-pkg` | `CheckLayerDirection` | inner domain layer imports `internal/pkg/...` |
| `layer.unknown-sublayer` | `CheckLayerDirection` | unsupported domain sublayer |
| `naming.no-stutter` | `CheckNaming` | exported type repeats package name |
| `naming.no-impl-suffix` | `CheckNaming` | exported type ends with `Impl` |
| `naming.snake-case-file` | `CheckNaming` | file name is not snake_case |
| `naming.repo-file-interface` | `CheckNaming` | repo file is missing its matching interface |
| `naming.no-layer-suffix` | `CheckNaming` | file name redundantly repeats the layer |
| `naming.handler-no-exported-interface` | `CheckNaming` | handler package defines an exported interface |
| `structure.banned-package` | `CheckStructure` | banned generic package name under `internal/` |
| `structure.legacy-package` | `CheckStructure` | legacy or misplaced package directory |
| `structure.middleware-placement` | `CheckStructure` | middleware directory is outside `internal/pkg/` |
| `structure.domain-root-alias-required` | `CheckStructure` | domain root is missing `alias.go` |
| `structure.domain-root-alias-package` | `CheckStructure` | `alias.go` package name mismatches the domain name |
| `structure.domain-root-alias-only` | `CheckStructure` | domain root contains non-test Go files other than `alias.go` |
| `structure.domain-model-required` | `CheckStructure` | domain is missing a non-empty `core/model/` |
| `structure.dto-placement` | `CheckStructure` | DTO file is placed under `domain/` or `infra/` |

## License

MIT
