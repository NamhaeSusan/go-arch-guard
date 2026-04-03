# go-arch-guard

[![CI](https://github.com/NamhaeSusan/go-arch-guard/actions/workflows/ci.yml/badge.svg)](https://github.com/NamhaeSusan/go-arch-guard/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/NamhaeSusan/go-arch-guard/branch/main/graph/badge.svg)](https://codecov.io/gh/NamhaeSusan/go-arch-guard)
[![Go Report Card](https://goreportcard.com/badge/github.com/NamhaeSusan/go-arch-guard)](https://goreportcard.com/report/github.com/NamhaeSusan/go-arch-guard)

[한국어](README.ko.md)

Architecture guardrails for Go projects via `go test`, built for AI coding agents and fast-moving teams.

Define isolation, layer-direction, structure, naming, and blast-radius rules, then fail regular tests when the project shape drifts. Ships with **DDD**, **Clean Architecture**, **Layered**, **Hexagonal**, **Modular Monolith**, **Consumer/Worker**, **Batch**, and **Event-Driven Pipeline** presets, and supports fully custom architecture models. No CLI to learn. No separate config format. Just Go tests.

AI-agent-friendly by default:

- `scaffold.ArchitectureTest(...)` generates a ready-to-copy `architecture_test.go`
- `rules.RunAll(...)` runs the recommended rule bundle in one call
- `report.MarshalJSONReport(...)` emits machine-readable violations for bots and remediation loops

## Why

Architecture usually decays through a few broad mistakes, not through deep theoretical violations:

- cross-domain imports
- hidden composition roots
- package placement drift
- naming that breaks the intended project shape

`go-arch-guard` catches those coarse mistakes early via static analysis. It is designed to be simple enough for AI agents to scaffold and maintain, while still being useful for humans reviewing the resulting boundaries. It does not try to model every semantic nuance inside Go packages, and if Go already rejects something by itself (such as import cycles), that is not a primary target here.

## Install

```bash
go get github.com/NamhaeSusan/go-arch-guard
```

## Quick Start

### Generate a preset template

For AI agents or scaffolding tools, generate a ready-to-copy `architecture_test.go`
instead of hand-copying the snippets below:

```go
import "github.com/NamhaeSusan/go-arch-guard/scaffold"

src, err := scaffold.ArchitectureTest(
    scaffold.PresetHexagonal,
    scaffold.ArchitectureTestOptions{PackageName: "myapp_test"},
)
```

`PackageName` must be a valid Go package identifier. Do not derive it blindly
from a hyphenated module basename.

Available presets: `PresetDDD`, `PresetCleanArch`, `PresetLayered`,
`PresetHexagonal`, `PresetModularMonolith`, `PresetConsumerWorker`, `PresetBatch`,
`PresetEventPipeline`.

### Recommended shortcut

If you want the recommended rule bundle without manually appending each check:

```go
violations := rules.RunAll(pkgs, "", "")
report.AssertNoViolations(t, violations)
```

Pass `opts...` only when you need a non-default model or severity/exclude options.

### Per-rule control (DDD example)

For finer control over individual checks, compose them manually:

```go
func TestArchitecture(t *testing.T) {
    pkgs, err := analyzer.Load(".", "internal/...", "cmd/...")
    if err != nil {
        t.Log(err)
    }
    if len(pkgs) == 0 {
        t.Fatalf("no packages loaded: %v", err)
    }

    t.Run("domain isolation", func(t *testing.T) {
        report.AssertNoViolations(t, rules.CheckDomainIsolation(pkgs, "", ""))
    })
    t.Run("layer direction", func(t *testing.T) {
        report.AssertNoViolations(t, rules.CheckLayerDirection(pkgs, "", ""))
    })
    t.Run("naming", func(t *testing.T) {
        report.AssertNoViolations(t, rules.CheckNaming(pkgs))
    })
    t.Run("structure", func(t *testing.T) {
        report.AssertNoViolations(t, rules.CheckStructure("."))
    })
    t.Run("blast radius", func(t *testing.T) {
        report.AssertNoViolations(t, rules.AnalyzeBlastRadius(pkgs, "", ""))
    })
}
```

For other presets, add `opts` with the model function:

```go
m := rules.CleanArch() // or Layered(), Hexagonal(), ModularMonolith(), ConsumerWorker(), Batch(), EventPipeline()
opts := []rules.Option{rules.WithModel(m)}

rules.CheckDomainIsolation(pkgs, "", "", opts...)
rules.CheckLayerDirection(pkgs, "", "", opts...)
// ... same pattern for all Check* functions
```

### Custom Model

```go
m := rules.NewModel(
    rules.WithDomainDir("module"),
    rules.WithSharedDir("lib"),
    rules.WithSublayers([]string{"api", "logic", "data"}),
    rules.WithDirection(map[string][]string{
        "api":   {"logic"},
        "logic": {"data"},
        "data":  {},
    }),
)
opts := []rules.Option{rules.WithModel(m)}
```

Run:

```bash
go test -run TestArchitecture -v
```

Sample output when violations exist:

```text
=== RUN   TestArchitecture/domain_isolation
    [ERROR] violation: domain "order" must not import domain "user"
    (file: internal/domain/order/app/service.go:5,
     rule: isolation.cross-domain,
     fix: use orchestration/ for cross-domain orchestration or move shared types to pkg/)
--- FAIL: TestArchitecture/domain_isolation
```

Pass empty strings for `module` and `root` to auto-extract from loaded packages. If the module cannot be determined, a `meta.no-matching-packages` warning is emitted.

## Presets

| Preset | Type | Sublayers | Direction |
|--------|------|-----------|-----------|
| `DDD()` | Domain | handler, app, core/model, core/repo, core/svc, event, infra | handler->app->core/\*, infra->core/repo |
| `CleanArch()` | Domain | handler, usecase, entity, gateway, infra | handler->usecase->entity+gateway, infra->gateway |
| `Layered()` | Domain | handler, service, repository, model | handler->service->repository+model |
| `Hexagonal()` | Domain | handler, usecase, port, domain, adapter | handler->usecase->port+domain, adapter->port+domain |
| `ModularMonolith()` | Domain | api, application, core, infrastructure | api->application->core, infrastructure->core |
| `ConsumerWorker()` | Flat | worker, service, store, model | worker→service+model, service→store+model, store→model |
| `Batch()` | Flat | job, service, store, model | job→service+model, service→store+model, store→model |
| `EventPipeline()` | Flat | command, aggregate, event, projection, eventstore, readstore, model | command→aggregate+eventstore+model, aggregate→event+model, projection→event+readstore+model |

Domain presets use `internal/domain/{name}/{layer}/` layout.
Flat presets use `internal/{layer}/` layout (no domain directory).

See [preset details](docs/presets.md) for full layout diagrams and direction tables.

### Custom Model Options

Start from DDD defaults and override what you need:

```go
m := rules.NewModel(
    rules.WithDomainDir("module"),          // internal/module/ instead of internal/domain/
    rules.WithOrchestrationDir("workflow"), // internal/workflow/
    rules.WithSharedDir("lib"),             // internal/lib/
    rules.WithSublayers([]string{"api", "logic", "data"}),
    rules.WithDirection(map[string][]string{
        "api":   {"logic"},
        "logic": {"data"},
        "data":  {},
    }),
    rules.WithRequireAlias(false),
    rules.WithRequireModel(false),
)
```

All model options:

| Option | Description |
|--------|-------------|
| `WithSublayers([]string{...})` | recognized sublayer names |
| `WithDirection(map[string][]string{...})` | allowed import direction matrix |
| `WithPkgRestricted(map[string]bool{...})` | sublayers that must not import shared pkg |
| `WithDomainDir("domain")` | top-level directory name for domains |
| `WithOrchestrationDir("orchestration")` | top-level directory name for orchestration |
| `WithSharedDir("pkg")` | top-level directory name for shared packages |
| `WithRequireAlias(bool)` | whether domain roots must define alias.go |
| `WithAliasFileName("alias.go")` | name of the alias file |
| `WithRequireModel(bool)` | whether domains must have a model directory |
| `WithModelPath("core/model")` | path to domain model directory |
| `WithDTOAllowedLayers([]string{...})` | sublayers where DTOs are allowed |
| `WithBannedPkgNames([]string{...})` | package names banned under internal/ |
| `WithLegacyPkgNames([]string{...})` | package names that trigger migration warnings |
| `WithLayerDirNames(map[string]bool{...})` | directory names considered "layer-like" for naming checks |
| `WithInterfacePatternExclude(map[string]bool{...})` | layers to skip for interface pattern checks |

## Isolation Rules

`rules.CheckDomainIsolation(pkgs, module, root, opts...)`

Prevents domains from leaking into each other. Without isolation, a change in domain A
can silently break domain B --- the most common source of unintended coupling in DDD projects.

### `isolation.cross-domain`

Domains must not import other domains directly.

```go
// internal/domain/order/app/service.go
package app

import _ "myapp/internal/domain/user/app"  // violation
```

```go
// use orchestration for cross-domain coordination
package orchestration

import (
    "myapp/internal/domain/order"
    "myapp/internal/domain/user"
)
```

### `isolation.cmd-deep-import`

`cmd/` must only import domain root packages (alias), not sub-packages.

```go
// cmd/server/main.go
import _ "myapp/internal/domain/order/app"  // too deep

import _ "myapp/internal/domain/order"  // domain root only
```

### `isolation.orchestration-deep-import`

Orchestration must only import domain roots, keeping the coupling surface minimal.

```go
// internal/orchestration/checkout.go
import _ "myapp/internal/domain/order/app"  // too deep

import _ "myapp/internal/domain/order"  // domain root only
```

### `isolation.pkg-imports-domain`

Shared `pkg/` must not import any domain --- it should be domain-agnostic.

```go
// internal/pkg/logger/logger.go
import _ "myapp/internal/domain/order"  // violation: pkg depends on domain
```

### `isolation.pkg-imports-orchestration`

Shared `pkg/` must not import orchestration.

### `isolation.domain-imports-orchestration`

Domains must not import orchestration --- orchestration coordinates domains, not the reverse.

### `isolation.internal-imports-orchestration`

Only `cmd/` and orchestration itself may depend on orchestration.

### `isolation.internal-imports-domain`

Non-domain internal packages (other than orchestration/cmd/pkg) must not import domains.

**Import matrix:**

| from | domain root | domain sub-pkg | orchestration | shared pkg |
|------|:-:|:-:|:-:|:-:|
| **same domain** | Yes | Yes | No | Yes |
| **other domain** | No | No | No | Yes |
| **orchestration** | Yes | No | Yes | Yes |
| **cmd** | Yes | No | Yes | Yes |
| **shared pkg** | No | No | No | Yes |

> **Flat-layout presets** (ConsumerWorker, Batch, EventPipeline): isolation rules are
> skipped entirely --- there are no domains to isolate.

## Layer Direction Rules

`rules.CheckLayerDirection(pkgs, module, root, opts...)`

Prevents reverse dependencies between layers. Without direction enforcement,
inner layers (model, entity) gradually accumulate imports from outer layers,
making them impossible to extract or test independently.

### `layer.direction`

Imports must follow the allowed direction defined by the preset's direction matrix.

```go
// DDD preset: core/svc may only import core/model
package svc // internal/domain/order/core/svc/

import _ "myapp/internal/domain/order/app"  // reverse direction

import _ "myapp/internal/domain/order/core/model"  // allowed
```

### `layer.inner-imports-pkg`

Inner layers marked as `PkgRestricted` must not import shared `pkg/`.
This keeps core domain logic free of infrastructure concerns.

```go
// DDD: core/model is PkgRestricted
package model // internal/domain/order/core/model/

import _ "myapp/internal/pkg/logger"  // model must be self-contained
```

### `layer.unknown-sublayer`

Detects directories under a domain that don't match any recognized sublayer name.

```
internal/domain/order/utils/   "utils" is not a recognized sublayer
```

> **Flat-layout presets**: layers are checked at `internal/` top level instead of within domains.

## Structure Rules

`rules.CheckStructure(root, opts...)`

Enforces filesystem layout conventions that prevent structural drift during vibe coding.

### `structure.internal-top-level`

Only allowed directories may exist at the `internal/` top level.

```
// DDD: only domain/, orchestration/, pkg/ allowed
internal/
  domain/          allowed
  orchestration/   allowed
  pkg/             allowed
  config/          not in allowed list
```

### `structure.banned-package`

Blocks vague package names that become dumping grounds.

Banned by default: `util`, `common`, `misc`, `helper`, `shared`, `services`

```
internal/domain/order/app/util/   "util" is banned
```

### `structure.legacy-package`

Warns about package names that should be migrated: `router`, `bootstrap`

### `structure.misplaced-layer`

Layer directories (`app`, `handler`, `infra`) must only exist inside domain slices,
not floating at the internal/ top level.

### `structure.middleware-placement`

`middleware/` must live in `internal/pkg/middleware/`, not scattered across domains.

### `structure.domain-root-alias-required` (DDD only)

Each domain root must define an `alias.go` file as its public API surface.

### `structure.domain-root-alias-package`

The alias file's package name must match the directory name.

### `structure.domain-root-alias-only`

Domain root directories may only contain `alias.go` --- all other code goes in sublayers.

### `structure.domain-alias-no-interface`

Alias files must not re-export interfaces --- this leaks cross-domain contracts.

### `structure.domain-model-required` (DDD only)

Each domain must have a `core/model/` directory with at least one Go file.

### `structure.dto-placement`

DTO files (`dto.go`, `*_dto.go`) may only exist in allowed layers (handler, app).

## Naming Rules

`rules.CheckNaming(pkgs, opts...)`

Enforces Go naming conventions that keep the codebase consistent and grep-friendly.

### `naming.no-stutter`

Exported types must not repeat the package name.

```go
package repo

type RepoOrder struct{}  // stutters: repo.RepoOrder
type Order struct{}      // clean: repo.Order
```

### `naming.no-impl-suffix`

Exported types must not end with `Impl`. Use unexported types instead.

```go
type OrderServiceImpl struct{}  // Impl suffix
type orderService struct{}      // unexported
```

### `naming.snake-case-file`

All Go filenames must be snake_case.

```
OrderService.go   violation
order_service.go  correct
```

### `naming.repo-file-interface`

Files in `repo/` (or `core/repo/`) must contain an interface matching the filename.

```go
// order.go in repo/ must define:
type Order interface { ... }  // matches filename
```

### `naming.no-layer-suffix`

Filenames must not redundantly repeat the layer name.

```
// inside service/ directory:
order_service.go  "_service" suffix is redundant
order.go          correct
```

### `naming.domain-interface-repo-only` (DDD only)

Interfaces within a domain must be defined in `core/repo/`, not scattered across layers.

### `naming.no-handmock`

Test files must not define hand-rolled mock/fake/stub structs with methods.
Use mockery or other generation tools instead.

### `naming.type-pattern-mismatch` (flat presets)

Files matching a TypePattern prefix must define the corresponding type.

```go
// worker/worker_order.go must define:
type OrderWorker struct{}  // expected

type SomethingElse struct{}  // expected OrderWorker
```

### `naming.type-pattern-missing-method` (flat presets)

Types matched by TypePattern must have the required method.

```go
type OrderWorker struct{}
// missing Process method  --- violation

func (w *OrderWorker) Process(ctx context.Context) error { ... }  // correct
```

## Interface Pattern Rules

`rules.CheckInterfacePattern(pkgs, opts...)`

Enforces Go interface best practices: private implementation, `New()`-only constructor,
interface return type, and single interface per package.

### `interface.exported-impl`

Exported structs must not implement interfaces --- make implementation types unexported
to prevent consumers from depending on the concrete type.

```go
type RepositoryImpl struct{ db *sql.DB }  // exported struct implements interface
type repository struct{ db *sql.DB }      // unexported --- correct
```

### `interface.constructor-name`

Constructors must be named `New`, not `NewXxx` variants. This enforces a consistent
factory pattern across all packages.

```go
func NewRepository(db *sql.DB) Repository  // NewXxx not allowed
func New(db *sql.DB) Repository            // correct
```

### `interface.constructor-returns-interface`

`New()` must return an interface, not a concrete type. This ensures callers depend
on the contract, not the implementation.

```go
func New(db *sql.DB) *repository  // returns concrete type
func New(db *sql.DB) Repository   // returns interface --- correct
```

### `interface.single-per-package`

At most one exported interface per package (Warning). Multiple interfaces in one package
typically signal that the package has too many responsibilities.

Excluded layers per preset (entry points, model, event, pkg) are controlled by `InterfacePatternExclude`.

## Blast Radius

`rules.AnalyzeBlastRadius(pkgs, module, root, opts...)`

Surfaces internal packages with abnormally high coupling via IQR-based statistical outlier detection. Default severity is Warning. Skips projects with fewer than 5 internal packages.

| Rule | Meaning |
|------|---------|
| `blast-radius.high-coupling` | package has statistically outlying transitive dependents |

| Metric | Definition |
|--------|-----------|
| Ca (Afferent Coupling) | packages that import this package |
| Ce (Efferent Coupling) | packages this package imports |
| Instability | Ce / (Ca + Ce) |
| Transitive Dependents | full reverse-reachable set via BFS |

## Options

### Severity

```go
// Log violations without failing the test
rules.CheckDomainIsolation(pkgs, "", "", rules.WithSeverity(rules.Warning))
```

### Exclude Paths

```go
// Skip subtrees during migration
rules.CheckDomainIsolation(pkgs, "", "",
    rules.WithExclude("internal/legacy/..."),
)
```

Patterns are project-relative paths with forward slashes. `...` matches the root and all descendants.

## TUI Viewer

Visualize your project's package structure and dependencies in an interactive terminal UI.

```bash
go run github.com/NamhaeSusan/go-arch-guard/cmd/tui .
```

Features: health-status tree coloring, imports/reverse dependencies/coupling metrics, violation details, search/filter (`/`), keyboard navigation.

## API Reference

| Function | Description |
|----------|-------------|
| `analyzer.Load(dir, patterns...)` | load Go packages for analysis |
| `rules.CheckDomainIsolation(pkgs, module, root, opts...)` | cross-domain boundary checks |
| `rules.CheckLayerDirection(pkgs, module, root, opts...)` | intra-domain direction checks |
| `rules.CheckNaming(pkgs, opts...)` | naming convention checks |
| `rules.CheckStructure(root, opts...)` | filesystem structure checks |
| `rules.AnalyzeBlastRadius(pkgs, module, root, opts...)` | coupling outlier detection |
| `rules.CheckInterfacePattern(pkgs, opts...)` | interface pattern best practices |
| `rules.RunAll(pkgs, module, root, opts...)` | run the recommended built-in rule bundle |
| `report.AssertNoViolations(t, violations)` | fail test on Error violations |
| `report.BuildJSONReport(violations)` | build a machine-readable JSON-friendly report |
| `report.MarshalJSONReport(violations)` | marshal a machine-readable JSON report |
| `report.WriteJSONReport(w, violations)` | write a machine-readable JSON report |
| `scaffold.ArchitectureTest(preset, opts)` | generate a preset-specific `architecture_test.go` template |
| `rules.DDD()` | DDD architecture model (default) |
| `rules.CleanArch()` | Clean Architecture model |
| `rules.Layered()` | Spring-style layered model |
| `rules.Hexagonal()` | Ports & Adapters model |
| `rules.ModularMonolith()` | Module-based layered model |
| `rules.ConsumerWorker()` | Consumer/Worker flat-layout model |
| `rules.Batch()` | Batch flat-layout model |
| `rules.EventPipeline()` | Event-sourcing / CQRS flat-layout model |
| `rules.CheckTypePatterns(pkgs, opts...)` | AST-based type pattern enforcement |
| `rules.NewModel(opts...)` | custom model builder |
| `rules.WithModel(m)` | apply custom model to checks |
| `rules.WithSeverity(rules.Warning)` | downgrade to warnings |
| `rules.WithExclude("path/...")` | skip a subtree |

## Machine-readable JSON Output

For CI, bots, or AI remediation loops, you can emit the same violations as JSON:

```go
import "github.com/NamhaeSusan/go-arch-guard/report"

data, err := report.MarshalJSONReport(violations)
if err != nil {
    return err
}
fmt.Println(string(data))
```

## Claude Code Plugin

```text
/plugin marketplace add NamhaeSusan/go-arch-guard
/plugin install go-arch-guard@go-arch-guard-marketplace
```

## External Import Hygiene

`go-arch-guard` checks **project-internal** imports only. External dependency hygiene should be enforced via AI tool instructions and code review. See the [DDD external import constraints](README.ko.md#외부-import-위생--이-라이브러리가-아닌-ai-도구-지침으로-강제) for a copy-paste template.

## License

MIT
