# go-arch-guard

[![CI](https://github.com/NamhaeSusan/go-arch-guard/actions/workflows/ci.yml/badge.svg)](https://github.com/NamhaeSusan/go-arch-guard/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/NamhaeSusan/go-arch-guard/branch/main/graph/badge.svg)](https://codecov.io/gh/NamhaeSusan/go-arch-guard)
[![Go Report Card](https://goreportcard.com/badge/github.com/NamhaeSusan/go-arch-guard)](https://goreportcard.com/report/github.com/NamhaeSusan/go-arch-guard)

[한국어](README.ko.md)

Architecture guardrails for Go projects via `go test`, built for AI coding agents and fast-moving teams.

Define isolation, layer-direction, structure, naming, and blast radius rules, then fail regular tests when the project shape drifts. Ships with **DDD**, **Clean Architecture**, **Layered**, **Hexagonal**, **Modular Monolith**, **Consumer/Worker**, **Batch**, and **Event-Driven Pipeline** presets, and supports fully custom architecture models. No CLI to learn. No separate config format. Just Go tests.

AI-agent-friendly by default:

- `scaffold.ArchitectureTest(...)` generates a ready-to-copy `architecture_test.go`
- `core.Run(ctx, presets.RecommendedDDD())` runs the recommended rule bundle in one call
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
arch := presets.DDD()
ctx := core.NewContext(pkgs, "", "", arch, nil)
violations := core.Run(ctx, presets.RecommendedDDD())

report.AssertNoViolations(t, violations)
```

Use the matching architecture and recommended ruleset for other presets, for example
`presets.Hexagonal()` with `presets.RecommendedHexagonal()`.

Passing empty strings for `module` and `root` asks `core.NewContext` to derive
both values from loaded package module metadata. If package metadata is not
available, the values stay empty and layout-dependent rules may report
`meta.layout-not-supported` instead of guessing a project root.

### Per-rule control (DDD example)

For finer control over individual checks, compose a `core.RuleSet` manually:

```go
func TestArchitecture(t *testing.T) {
    pkgs, err := analyzer.Load(".", "internal/...", "cmd/...")
    if err != nil {
        t.Log(err)
    }
    if len(pkgs) == 0 {
        t.Fatalf("no packages loaded: %v", err)
    }

    arch := presets.DDD()
    ctx := core.NewContext(pkgs, "", "", arch, nil)
    ruleset := core.NewRuleSet(
        dependency.NewIsolation(),
        dependency.NewLayerDirection(),
        naming.NewNoStutter(),
        structural.NewAlias(),
    )

    report.AssertNoViolations(t, core.Run(ctx, ruleset))
}
```

### Custom Architecture

```go
arch := core.Architecture{
    Layers: core.LayerModel{
        Sublayers: []string{"api", "logic", "data"},
        Direction: map[string][]string{
            "api":   {"logic"},
            "logic": {"data"},
            "data":  {},
        },
        InternalTopLevel: map[string]bool{
            "module": true,
            "lib":    true,
        },
    },
    Layout: core.LayoutModel{
        DomainDir: "module",
        SharedDir: "lib",
    },
    Naming: core.NamingPolicy{
        BannedPkgNames: []string{"util", "common", "misc", "helper", "shared", "services"},
        LegacyPkgNames: []string{"router", "bootstrap"},
        AliasFileName:  "alias.go",
    },
}
if err := arch.Validate(); err != nil {
    t.Fatal(err)
}
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
     rule: dependency.cross-domain,
     fix: use orchestration/ for cross-domain orchestration or move shared types to pkg/)
--- FAIL: TestArchitecture/domain_isolation
```

Pass empty strings for `module` and `root` to derive them from loaded package module metadata. If metadata is unavailable, the values remain empty rather than being guessed; layout-dependent rules may emit `meta.layout-not-supported`. If a rule panics, `core.Run` emits `meta.rule-panic` with Error severity and continues running the remaining rules.

## Presets

| Preset | Type | Sublayers | Direction |
|--------|------|-----------|-----------|
| `DDD()` | Domain | handler, app, core/model, core/repo, core/svc, event, infra | handler->app->core/\*, infra->core/repo+core/model+event |
| `CleanArch()` | Domain | handler, usecase, entity, gateway, infra | handler->usecase->entity+gateway, infra->gateway+entity |
| `Layered()` | Domain | handler, service, repository, model | handler->service->repository+model |
| `Hexagonal()` | Domain | handler, usecase, port, domain, adapter | handler->usecase->port+domain, adapter->port+domain |
| `ModularMonolith()` | Domain | api, application, core, infrastructure | api->application->core, infrastructure->core |
| `ConsumerWorker()` | Flat | worker, service, store, model | worker→service+model, service→store+model, store→model |
| `Batch()` | Flat | job, service, store, model | job→service+model, service→store+model, store→model |
| `EventPipeline()` | Flat | command, aggregate, event, projection, eventstore, readstore, model | command→aggregate+eventstore+model, aggregate→event+model, projection→event+readstore+model |

Domain presets use `internal/domain/{name}/{layer}/` layout.
Flat presets use `internal/{layer}/` layout (no domain directory).

See [preset details](docs/presets.md) for full layout diagrams and direction tables.

### Architecture Fields

Build a custom architecture directly with a `core.Architecture` literal:

```go
arch := core.Architecture{
    Layers: core.LayerModel{
        Sublayers: []string{"api", "logic", "data"},
        Direction: map[string][]string{
            "api":   {"logic"},
            "logic": {"data"},
            "data":  {},
        },
        InternalTopLevel: map[string]bool{"module": true, "workflow": true, "lib": true},
    },
    Layout: core.LayoutModel{
        DomainDir:        "module",
        OrchestrationDir: "workflow",
        SharedDir:        "lib",
    },
    Structure: core.StructurePolicy{
        RequireAlias: false,
        RequireModel: false,
    },
}
```

For conceptual explanations of each field — what it means, why it exists, and
when to set it — see [Model Concepts](docs/model-concepts.md).

Architecture fields:

| Field | Description |
|-------|-------------|
| `LayerModel.Sublayers` | authoritative layer-path vocabulary (`"core/repo"`); referenced by direction, port, and contract rules |
| `LayerModel.Direction` | allowed import direction matrix (keys must be in `Sublayers`) |
| `LayerModel.PortLayers` | pure interface layers such as repo, gateway, or Hexagonal port (must be in `Sublayers`; `port` is not inferred unless listed) |
| `LayerModel.ContractLayers` | contract layers; must be a superset of `PortLayers` |
| `LayerModel.PkgRestricted` | sublayers that must not import shared packages |
| `LayerModel.InternalTopLevel` | allowed top-level directories under the package root |
| `LayerModel.LayerDirNames` | layer **basenames** (`"repo"`) recognized by file/directory placement rules; intentionally NOT required to appear in `Sublayers` |
| `LayoutModel.InternalRoot` | project-relative package-root directory; defaults to `"internal"` when empty (set to `"packages"`, `"src"`, etc. for non-default layouts) |
| `LayoutModel.DomainDir` | top-level directory name for domains; empty for flat layouts |
| `LayoutModel.OrchestrationDir` | top-level directory name for orchestration |
| `LayoutModel.SharedDir` | top-level directory name for shared packages |
| `LayoutModel.AppDir` | top-level composition-root directory |
| `LayoutModel.ServerDir` | top-level transport directory |
| `NamingPolicy.BannedPkgNames` | package names banned under `internal/` |
| `NamingPolicy.LegacyPkgNames` | package names that trigger migration warnings |
| `NamingPolicy.AliasFileName` | domain alias filename |
| `StructurePolicy.RequireAlias` | whether domain roots must define an alias file |
| `StructurePolicy.RequireModel` | whether domains must have a model directory |
| `StructurePolicy.ModelPath` | path to the domain model directory |
| `StructurePolicy.TypePatterns` | AST naming/structure patterns for flat layouts |
| `StructurePolicy.InterfacePatternExclude` | layers skipped by interface pattern checks |

`core.Validate(arch)` and `arch.Validate()` enforce direction completeness,
layer references, and `PortLayers ⊆ ContractLayers`.

```go
arch := presets.DDD()
arch.Layout.DomainDir = "module"
arch.Layout.SharedDir = "lib"
arch.Layers.Sublayers = []string{"api", "logic", "data"}
arch.Layers.Direction = map[string][]string{
        "api":   {"logic"},
        "logic": {"data"},
        "data":  {},
}
arch.Structure.RequireAlias = false
arch.Structure.RequireModel = false
```

### Custom package root (non-`internal/` layouts)

Projects that keep their packages under `packages/`, `src/`, or any other directory
instead of the canonical `internal/` can set `Layout.InternalRoot`:

```go
arch := presets.DDD()
arch.Layout.InternalRoot = "packages" // packages/domain/order/...
```

Empty `InternalRoot` is normalized to `"internal"` at construction, so existing
configurations keep working unchanged. Layout-dependent rules emit
`meta.layout-not-supported` (Warning) when no `<root>/<InternalRoot>/` directory
is found, instead of silently producing zero violations.

`scaffold.ArchitectureTest` honors the same field via
`ArchitectureTestOptions.InternalRoot` so generated `architecture_test.go`
matches the project's actual layout.

## Isolation Rules

`dependency.NewIsolation()`

Prevents domains from leaking into each other. Without isolation, a change in domain A
can silently break domain B --- the most common source of unintended coupling in DDD projects.

### `dependency.cross-domain`

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

### `dependency.cmd-deep-import`

`cmd/` must only import domain root packages (alias), not sub-packages.

```go
// cmd/server/main.go
import _ "myapp/internal/domain/order/app"  // too deep

import _ "myapp/internal/domain/order"  // domain root only
```

### `dependency.orchestration-deep-import`

Orchestration must only import domain roots, keeping the coupling surface minimal.

```go
// internal/orchestration/checkout.go
import _ "myapp/internal/domain/order/app"  // too deep

import _ "myapp/internal/domain/order"  // domain root only
```

### `dependency.pkg-imports-domain`

Shared `pkg/` must not import any domain --- it should be domain-agnostic.

```go
// internal/pkg/logger/logger.go
import _ "myapp/internal/domain/order"  // violation: pkg depends on domain
```

### `dependency.pkg-imports-orchestration`

Shared `pkg/` must not import orchestration.

### `dependency.domain-imports-orchestration`

Domains must not import orchestration --- orchestration coordinates domains, not the reverse.

### `dependency.stray-imports-orchestration`

Only `cmd/` and orchestration itself may depend on orchestration.

### `dependency.stray-imports-domain`

Non-domain internal packages (other than orchestration/cmd/pkg/app/transport) must not import domains.

### `dependency.transport-imports-domain`

Transport packages (`internal/server/<proto>/`) must not import domain sub-packages directly.
They should go through the composition root (`internal/app/`) instead.

```go
// internal/server/http/handler.go
import _ "myapp/internal/domain/order/core/model"  // violation: transport imports domain directly
import _ "myapp/internal/app"                       // correct: go through composition root
```

### `dependency.transport-imports-orchestration`

Transport packages must not import orchestration directly.

### `dependency.transport-imports-unclassified`

Transport packages must not import unclassified internal packages (e.g. `internal/config`, `internal/bootstrap`).
Anything transport depends on must be routed through `internal/app/` (the composition root) or `internal/pkg/`.

```go
// internal/server/http/server.go
import _ "myapp/internal/config"  // violation: transport imports unclassified package
```

**Import matrix (DDD with app/server):**

| from | domain root | domain sub-pkg | orchestration | shared pkg | app | transport |
|------|:-:|:-:|:-:|:-:|:-:|:-:|
| **same domain** | Yes | Yes | No | Yes | No | No |
| **other domain** | No | No | No | Yes | No | No |
| **orchestration** | Yes | No | Yes | Yes | No | No |
| **cmd** | Yes | No | Yes | Yes | No | No |
| **shared pkg** | No | No | No | Yes | No | No |
| **app (composition root)** | Yes | Yes | Yes | Yes | Yes | No |
| **transport** | No | No | No | Yes | Yes | Yes |

> **Flat-layout presets** (ConsumerWorker, Batch, EventPipeline): isolation rules are
> skipped entirely --- there are no domains to isolate.

## Layer Direction Rules

`dependency.NewLayerDirection()`

Prevents reverse dependencies between layers. Without direction enforcement,
inner layers (model, entity) gradually accumulate imports from outer layers,
making them impossible to extract or test independently.

### `dependency.invalid-import-direction`

Imports must follow the allowed direction defined by the preset's direction matrix.

```go
// DDD preset: core/svc may only import core/model
package svc // internal/domain/order/core/svc/

import _ "myapp/internal/domain/order/app"  // reverse direction

import _ "myapp/internal/domain/order/core/model"  // allowed
```

### `dependency.inner-imports-pkg`

Inner layers marked as `PkgRestricted` must not import shared `pkg/`.
This keeps core domain logic free of infrastructure concerns.

```go
// DDD: core/model is PkgRestricted
package model // internal/domain/order/core/model/

import _ "myapp/internal/pkg/logger"  // model must be self-contained
```

### `dependency.unknown-sublayer`

Detects directories under a domain that don't match any recognized sublayer name.

```
internal/domain/order/utils/   "utils" is not a recognized sublayer
```

> **Flat-layout presets**: layers are checked at `internal/` top level instead of within domains.

## Structure Rules

`structural.NewInternalTopLevel()`, `structural.NewBannedPackage()`,
`structural.NewLayerPlacement()`, `structural.NewAlias()`, and
`structural.NewModelRequired()`

Enforces filesystem layout conventions that prevent structural drift during vibe coding.

### `structural.internal-top-level`

Only allowed directories may exist at the `internal/` top level.

```
// DDD: only domain/, orchestration/, pkg/ allowed
internal/
  domain/          allowed
  orchestration/   allowed
  pkg/             allowed
  config/          not in allowed list
```

### `structural.banned-package-name`

Blocks vague package names that become dumping grounds.

Banned by default: `util`, `common`, `misc`, `helper`, `shared`, `services`

```
internal/domain/order/app/util/   "util" is banned
```

### `structural.legacy-package`

Flags package names that should be migrated: `router`, `bootstrap`. Default severity is Error; downgrade with `WithSeverityOverride("structure.legacy-package", core.Warning)` during migration windows.

### `structural.misplaced-layer`

Layer directories (`app`, `handler`, `infra`) must only exist inside domain slices,
not floating at the internal/ top level.

### `structural.domain-alias-exists` (DDD only)

Each domain root must define an `alias.go` file as its public API surface.

### `structural.domain-alias-package`

The alias file's package name must match the directory name.

### `structural.domain-alias-exclusive`

Domain root directories may only contain `alias.go` --- all other code goes in sublayers.

### `structural.domain-alias-no-interface`

Alias files must not directly define interfaces --- this leaks cross-domain contracts.

### `structural.domain-alias-contract-reexport`

Alias files must not re-export types from contract sublayers (repo/svc) --- this creates hidden cross-domain dependencies.

### `structural.domain-model-required` (DDD only)

Each domain must have a `core/model/` directory with at least one Go file.

## Naming Rules

`naming.NewNoStutter()`, `naming.NewImplSuffix()`,
`naming.NewSnakeCaseFiles()`, `naming.NewNoLayerSuffix()`,
and `naming.NewTypePattern()`

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

### `structural.repo-file-interface-missing`

Files in a configured port layer (`core/repo/`, `gateway/`, Hexagonal `port/`,
etc.) must contain an interface matching the filename.

```go
// order.go in a port layer must define:
type Order interface { ... }  // matches filename
```

### `structural.repo-file-extra-interface`

Each file in a configured port layer must define exactly one interface. Extra
interfaces should be split into their own files.

```go
// core/repo/review.go or port/review.go
type Review interface { Find() }   // correct
type Helper interface { Assist() } // violation: move to helper.go
```

### `interfaces.too-many-methods`

Emitted by `interfaces.NewTooManyMethods()` (separate rule, not part of `interfaces.NewPattern`). Default cap is 10; override with `interfaces.WithMaxMethods(n)`. Every recommended bundle includes this rule with the default cap.

```go
// repo/review.go
type Review interface {
    // 11 methods --- violation (max 10)
}
```

To use a custom cap, build a custom RuleSet instead of using a recommended
bundle. `RuleSet.Without("interfaces.too-many-methods")` filters that violation
ID, so appending another `NewTooManyMethods` after `Without` would still hide
the custom rule's violations.

```go
ruleset := core.NewRuleSet(
    dependency.NewIsolation(),
    dependency.NewLayerDirection(),
    structural.NewAlias(),
    interfaces.NewPattern(),
    interfaces.NewTooManyMethods(interfaces.WithMaxMethods(7)),
)
```

`interfaces.WithMaxMethods` is a TooManyMethods option only — passing it to `NewPattern` / `NewContainer` / `NewCrossDomainAnonymous` is silently ignored.

### `naming.no-layer-suffix`

Filenames must not redundantly repeat the layer name.

```
// inside service/ directory:
order_service.go  "_service" suffix is redundant
order.go          correct
```

### `structural.interface-placement`

Repository-port interfaces — names ending in `Repository` or `Repo` by default
— must be defined in the architecture's configured port layer (`core/repo/` for
DDD, `gateway/` for Clean Architecture, `port/` for Hexagonal), not scattered
across layers. Consumer-defined interfaces (the Go idiom where a package
declares the small interface it consumes) are allowed anywhere they are used:
`handler/`, `app/`, `usecase/`, `svc/`, etc.

Also flags `type X = otherdomain.Repo` aliases that re-export a repository
interface from a port layer — cross-domain coordination belongs in
`orchestration/`.

Pass `structural.WithRepoPortSuffixes("Gateway", "Adapter", ...)` to match a
different vocabulary; default is `["Repository", "Repo"]`.

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

## Test Policy Rules

`testpolicy.NewNoHandMock()`

Constraints on what test files may contain. Lives in its own package because
testing concerns are orthogonal to naming or structural conventions.

### `testpolicy.no-handmock`

Test files must not define hand-rolled mock/fake/stub structs with methods.
Use a mock generator (e.g. mockery) and import the generated types from a
dedicated mocks package instead.

## Interface Pattern Rules

`interfaces.NewPattern()`, `interfaces.NewTooManyMethods()`,
`interfaces.NewContainer()`, and `interfaces.NewCrossDomainAnonymous()`

Enforces Go interface best practices: private implementation, `New()`-only constructor,
interface return type, and single interface per package.

### `interfaces.exported-impl`

Exported structs must not implement interfaces --- make implementation types unexported
to prevent consumers from depending on the concrete type.

```go
type RepositoryImpl struct{ db *sql.DB }  // exported struct implements interface
type repository struct{ db *sql.DB }      // unexported --- correct
```

### `interfaces.constructor-name`

Constructors must be named `New`, not `NewXxx` variants. This enforces a consistent
factory pattern across all packages.

```go
func NewRepository(db *sql.DB) Repository  // NewXxx not allowed
func New(db *sql.DB) Repository            // correct
```

### `interfaces.constructor-returns-interface`

`New()` must return an interface, not a concrete type. This ensures callers depend
on the contract, not the implementation.

```go
func New(db *sql.DB) *repository  // returns concrete type
func New(db *sql.DB) Repository   // returns interface --- correct
```

### `interfaces.single-per-package`

At most one exported interface per package (Warning). Multiple interfaces in one package
typically signal that the package has too many responsibilities.

Excluded layers per preset (entry points, model, event, pkg) are controlled by `InterfacePatternExclude`.

### `interfaces.cross-domain-anonymous`

Detects anonymous interfaces declared outside of their referenced domain — and outside the designated orchestration layer — whose method signatures touch types from another domain. Default severity is **Error**.

This rule enforces the convention that **cross-domain abstractions are owned by the orchestration package**, not by arbitrary wiring code. A `cmd/` (or `internal/pkg/`) package that declares an inline anonymous interface over a domain type is creating a parallel uncontrolled cross-domain surface; that adapter/abstraction belongs in `internal/orchestration/`.

```go
// flagged: cmd/ declares inline interface that abstracts a domain type
package main

import "example.com/p/internal/domain/user"

type adapter struct {
    repo interface {                                          // ← cross-domain anonymous in cmd/
        GetByID(ctx context.Context, id string) (*user.User, error)
    }
}
```

```go
// not flagged: same shape but inside the orchestration layer where
// cross-domain coordination is by design
package orchestration

import "example.com/p/internal/domain/user"

type userInfoAdapter struct {
    repo interface {                                          // ← anonymous, but orchestration is exempt
        GetByID(ctx context.Context, id string) (*user.User, error)
    }
}
```

The fix for a flagged occurrence is to **move the adapter into the orchestration package** and have wiring code call orchestration constructors instead of declaring its own interfaces.

Skipped:
- Test files (`_test.go`) where mock/fake fixtures naturally use this shape
- Empty interfaces (`interface{}`) and interfaces without method declarations
- Embedded interface types (e.g. `interface { io.Reader }`)
- Same-domain references (anonymous interface inside `internal/domain/X` referencing `internal/domain/X` types)
- Packages inside `internal/<OrchestrationDir>/` — orchestration is the designated cross-domain coordination layer
- Models with no `DomainDir` (flat layouts like ConsumerWorker, Batch, EventPipeline)

### `interfaces.container-only`

Detects interfaces declared in a package that are used **only as struct field types** —
never as a function parameter or return type. Default severity is **Warning**.

This is a vibe-coding smell: the interface is being used as a value container rather than
as an abstraction. A common cause is a wiring layer that needs to hold a value whose
concrete type is not exposed (e.g. an `alias.go` re-exports the constructor but not the
type), so the developer declares a local interface just to give the field a type.

```go
// flagged: container-only — never used as parameter or return
type userRepo interface {
    GetByID(id string) string
}

type holder struct {
    r userRepo  // only usage
}
```

```go
// not flagged: legitimate consumer-defined interface
type userRepo interface {
    GetByID(id string) string
}

func newHolder(r userRepo) *holder {  // used as parameter → real abstraction
    return &holder{r: r}
}
```

Skipped:
- Test files (`_test.go`) where mock/fake fixtures naturally use this shape
- Type aliases (`type Foo = pkg.Foo`)
- Embedded fields (anonymous embedding) in structs
- Interfaces that are not used at all (different smell category — out of scope)

The rule does **not** prescribe a fix. It only points at the smell. Two common resolutions:
1. Re-export the concrete type from `alias.go` so the field can hold it directly.
2. Rewrite the wiring so the value is a local variable inside one function instead of a struct field shared between functions.

Severity can be upgraded to Error via `interfaces.WithSeverity(core.Error)` if a project wants to enforce
the smell as a hard rule.

## Orchestration Rules

`orchestration.NewLogicBudget()`

Opt-in advisory rule for packages under the configured orchestration directory
such as `internal/orchestration`. It flags functions whose branch count,
statement count, or cyclomatic complexity exceeds configurable budgets.
Default severity is **Warning** with budgets `maxBranches=8`,
`maxStatements=40`, and `maxCyclomatic=10`.

Simple `if err != nil { return err }` and `fmt.Errorf("%w", err)` branches are
discounted by default so ordinary Go error flow does not hide the real signal:
orchestration functions making business decisions or accumulating too much
coordination code. For `if err := call(); err != nil { return err }`, the
branch itself is discounted, but the `call()` init statement still counts.

```go
ruleset := core.NewRuleSet(orchestration.NewLogicBudget(
    orchestration.WithMaxBranches(6),
    orchestration.WithMaxStatements(30),
    orchestration.WithMaxCyclomatic(8),
))
```

Use `orchestration.WithCountErrorBranches()` for stricter accounting,
`orchestration.WithIgnoredFunctions(...)` for known exceptional functions, and
`orchestration.WithIgnoredPaths(...)` for subtrees such as transport handlers
that a team wants to govern separately.

## Blast Radius

`dependency.NewBlastRadius()`

Surfaces internal packages with abnormally high coupling via IQR-based statistical outlier detection. Default severity is Warning. Skips projects with fewer than 5 internal packages.

| Rule | Meaning |
|------|---------|
| `dependency.high-coupling` | package has statistically outlying transitive dependents |

| Metric | Definition |
|--------|-----------|
| Ca (Afferent Coupling) | packages that import this package |
| Ce (Efferent Coupling) | packages this package imports |
| Instability | Ce / (Ca + Ce) |
| Transitive Dependents | full reverse-reachable set via BFS |

## Tx Boundary

### `tx.New` (opt-in)

Gates where transactions may **start** and prevents transaction types from
**leaking** into function signatures outside an allowed layer. Fully opt-in —
does nothing unless you configure it. Add it to a `core.RuleSet` only when
your project has known transaction start symbols or transaction types.

```go
ruleset := presets.RecommendedDDD().With(tx.New(tx.Config{
    StartSymbols: []string{
        "database/sql.(*DB).BeginTx",
        "database/sql.(*DB).Begin",
    },
    Types:         []string{"database/sql.Tx"},
    AllowedLayers: []string{"app"}, // default when empty
}))
```

Emitted rule IDs: `tx.start-outside-allowed-layer`, `tx.type-in-signature`.

## Setter Pattern

### `types.NewNoSetter`

Flags exported setter methods (`Set*` on pointer receivers with at least one parameter) to steer custom types toward explicit constructor parameters.

**Recommended fix**: add the dependency as an explicit parameter on the constructor (`NewService(..., dep)`). Reserve the `With`-pattern option for dependencies that are truly optional and combine with many others — setters are rarely the right answer for that either.

- Fluent builders (methods returning the receiver type) are exempt.
- Test files and packages under `testdata/` or `mocks/` are auto-excluded.
- Default severity: Warning. Use `types.WithSeverity(core.Error)` for strict enforcement.

```go
// Default: Warning severity
report.AssertNoViolations(t, core.Run(ctx, core.NewRuleSet(types.NewNoSetter())))

// Strict: Error severity
report.AssertNoViolations(t, core.Run(ctx, core.NewRuleSet(types.NewNoSetter(types.WithSeverity(core.Error)))))
```

Emitted rule ID: `types.no-setter`.

## Options

### Severity

```go
// Log a specific violation without failing the test
core.Run(ctx, presets.RecommendedDDD(),
    core.WithSeverityOverride("dependency.cross-domain", core.Warning))
```

### Exclude Paths

```go
// Skip subtrees during migration
ctx := core.NewContext(pkgs, "", "", presets.DDD(), []string{"internal/legacy/..."})
violations := core.Run(ctx, presets.RecommendedDDD())
```

Patterns are project-relative paths with forward slashes. `...` matches the root and all descendants. Equivalent shapes (`/internal/foo`, `internal/foo`, `internal/foo/`, `./internal/foo`) all match the same paths after normalization.

### Meta Violations

The runner emits a small set of `meta.*` violations to surface environmental issues
without blocking builds by default. They are deduped by `(Rule, Message)` pair so distinct
diagnostics survive even when multiple rules emit the same ID.

| ID | Severity | When |
|---|---|---|
| `meta.no-matching-packages` | Warning | the configured project module does not match any loaded package |
| `meta.layout-not-supported` | Warning | a layout-dependent rule is run against a project without a recognized package root (`<root>/<InternalRoot>/`) |
| `meta.rule-disabled-by-config` | Warning | a rule (or one of its sub-checks) is registered in the ruleset but Architecture config disables it — e.g. `Structure.RequireAlias=false` for `structural.alias`, `Layout.DomainDir=""` for `dependency.isolation`, empty `tx.Config` for `tx.boundary` |
| `meta.rule-panic` | Error | a rule's `Check` panicked; the panic is captured and other rules continue to run |
| `meta.unknown-violation-id` | per rule | a rule emits a violation ID it didn't declare in `Spec().Violations` |

Promote any of them to a hard failure with `core.WithSeverityOverride(...)`, or filter them out with `RuleSet.Without(...)`.

## TUI Viewer

Visualize your project's package structure and dependencies in an interactive terminal UI.

```bash
go run github.com/NamhaeSusan/go-arch-guard/cmd/tui .
```

Pick a non-DDD preset with `--preset` (one of `ddd`, `cleanarch`, `layered`, `hexagonal`, `modular-monolith`, `consumer-worker`, `batch`, `event-pipeline`):

```bash
go run github.com/NamhaeSusan/go-arch-guard/cmd/tui --preset hexagonal .
```

Features: health-status tree coloring, imports/reverse dependencies/coupling metrics, violation details, search/filter (`/`), keyboard navigation.

## API Reference

| Function | Description |
|----------|-------------|
| `analyzer.Load(dir, patterns...)` | load Go packages for analysis |
| `core.NewContext(pkgs, module, root, arch, exclude)` | build the immutable analysis context |
| `core.Run(ctx, ruleset, opts...)` | execute a ruleset and return `[]core.Violation`; rule panics become `meta.rule-panic` Error violations |
| `core.RuleSet` | immutable collection of rules plus violation filters |
| `core.NewRuleSet(ruleValues...)` | create an immutable ruleset |
| `(rs).With(ruleValues...)` / `(rs).Without(ids...)` | append rules (nil entries are silently dropped) or filter violation IDs |
| `core.WithSeverityOverride(violationID, sev)` | override effective severity for one violation ID |
| `report.AssertNoViolations(t, violations)` | fail test on Error violations |
| `report.BuildJSONReport(violations)` | build a machine-readable JSON-friendly report |
| `report.MarshalJSONReport(violations)` | marshal a machine-readable JSON report |
| `report.WriteJSONReport(w, violations)` | write a machine-readable JSON report |
| `scaffold.ArchitectureTest(preset, opts)` | generate a preset-specific `architecture_test.go` template (`opts.InternalRoot` overrides the canonical `internal/`) |
| `presets.DDD()` / `presets.RecommendedDDD()` | DDD architecture and recommended ruleset |
| `presets.CleanArch()` / `presets.RecommendedCleanArch()` | Clean Architecture architecture and ruleset |
| `presets.Layered()` / `presets.RecommendedLayered()` | layered architecture and ruleset |
| `presets.Hexagonal()` / `presets.RecommendedHexagonal()` | Ports & Adapters architecture and ruleset |
| `presets.ModularMonolith()` / `presets.RecommendedModularMonolith()` | modular monolith architecture and ruleset |
| `presets.ConsumerWorker()` / `presets.RecommendedConsumerWorker()` | Consumer/Worker flat-layout architecture and ruleset |
| `presets.Batch()` / `presets.RecommendedBatch()` | Batch flat-layout architecture and ruleset |
| `presets.EventPipeline()` / `presets.RecommendedEventPipeline()` | event-sourcing / CQRS architecture and ruleset |
| `dependency.NewIsolation()` / `NewLayerDirection()` / `NewBlastRadius()` | dependency rules |
| `orchestration.NewLogicBudget()` | opt-in orchestration complexity budget rule |
| `naming.NewNoStutter()` / `NewImplSuffix()` / `NewSnakeCaseFiles()` / `NewNoLayerSuffix()` / `NewTypePattern()` | naming rules |
| `structural.NewAlias()` / `NewLayerPlacement()` / `NewBannedPackage()` / `NewModelRequired()` / `NewInternalTopLevel()` / `NewRepoFileInterface()` | structure rules |
| `structural.WithRepoPortSuffixes(...)` | option for `structural.NewRepoFileInterface` setting repository-port interface name suffixes. Default is `Repository`, `Repo`; blank suffixes are ignored. |
| `interfaces.NewPattern()` / `NewContainer()` / `NewCrossDomainAnonymous()` / `NewTooManyMethods()` | interface rules |
| `testpolicy.NewNoHandMock()` | test policy rules |
| `interfaces.WithMaxMethods(n)` | option for `interfaces.NewTooManyMethods` setting the per-interface method cap. Default 10 when the option is omitted; n ≤ 0 also falls back to 10. Silently ignored if passed to other interfaces.New*() rules. |
| `tx.New(tx.Config{...})` | transaction boundary enforcement (opt-in) |
| `types.NewNoSetter()` | setter rule (immutability for value types) |

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
