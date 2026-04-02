# go-arch-guard

[![CI](https://github.com/NamhaeSusan/go-arch-guard/actions/workflows/ci.yml/badge.svg)](https://github.com/NamhaeSusan/go-arch-guard/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/NamhaeSusan/go-arch-guard/branch/main/graph/badge.svg)](https://codecov.io/gh/NamhaeSusan/go-arch-guard)
[![Go Report Card](https://goreportcard.com/badge/github.com/NamhaeSusan/go-arch-guard)](https://goreportcard.com/report/github.com/NamhaeSusan/go-arch-guard)

[한국어](README.ko.md)

Architecture guardrails for Go projects via `go test`, built for AI coding agents and fast-moving teams.

Define isolation, layer-direction, structure, naming, and blast-radius rules, then fail regular tests when the project shape drifts. Ships with **DDD**, **Clean Architecture**, **Layered**, **Hexagonal**, **Modular Monolith**, **Consumer/Worker**, and **Batch** presets, and supports fully custom architecture models. No CLI to learn. No separate config format. Just Go tests.

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
`PresetHexagonal`, `PresetModularMonolith`, `PresetConsumerWorker`, `PresetBatch`.

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
m := rules.CleanArch() // or Layered(), Hexagonal(), ModularMonolith(), ConsumerWorker(), Batch()
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

## Architecture Models

### Built-in Presets

| Preset | Sublayers | Direction | Alias Required | Model Required |
|--------|-----------|-----------|:-:|:-:|
| `DDD()` | handler, app, core/model, core/repo, core/svc, event, infra | handler→app→core/\*, infra→core/repo | Yes | Yes |
| `CleanArch()` | handler, usecase, entity, gateway, infra | handler→usecase→entity+gateway, infra→gateway | No | No |
| `Layered()` | handler, service, repository, model | handler→service→repository+model, repository→model | No | No |
| `Hexagonal()` | handler, usecase, port, domain, adapter | handler→usecase→port+domain, adapter→port+domain | No | No |
| `ModularMonolith()` | api, application, core, infrastructure | api→application→core, infrastructure→core | No | No |
| `ConsumerWorker()` | worker, service, store, model | worker→service→store→model | No | No |
| `Batch()` | job, service, store, model | job→service→store→model | No | No |

### DDD Layout

```text
internal/
├── domain/
│   └── order/
│       ├── alias.go              # public surface (required)
│       ├── handler/http/         # inbound adapters
│       ├── app/                  # application service
│       ├── core/
│       │   ├── model/            # domain model (required)
│       │   ├── repo/             # repository interface
│       │   └── svc/              # domain service interface
│       ├── event/                # domain events
│       └── infra/persistence/    # outbound adapters
├── orchestration/                # cross-domain coordination
└── pkg/                          # shared utilities
```

DDD layer direction:

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

### Clean Architecture Layout

```text
internal/
├── domain/
│   └── product/
│       ├── handler/              # interface adapters (controllers)
│       ├── usecase/              # application business rules
│       ├── entity/               # enterprise business rules
│       ├── gateway/              # data access interfaces
│       └── infra/                # frameworks & drivers
├── orchestration/
└── pkg/
```

Clean Architecture layer direction:

| from | allowed to import |
|------|-------------------|
| `handler` | `usecase` |
| `usecase` | `entity`, `gateway` |
| `entity` | nothing |
| `gateway` | `entity` |
| `infra` | `gateway`, `entity` |

### Layered (Spring-style) Layout

```text
internal/
├── domain/
│   └── order/
│       ├── handler/              # HTTP/gRPC handlers
│       ├── service/              # business logic
│       ├── repository/           # data access
│       └── model/                # domain models
├── orchestration/
└── pkg/
```

Layered direction:

| from | allowed to import |
|------|-------------------|
| `handler` | `service` |
| `service` | `repository`, `model` |
| `repository` | `model` |
| `model` | nothing |

### Hexagonal (Ports & Adapters) Layout

```text
internal/
├── domain/
│   └── order/
│       ├── handler/              # driving adapters (HTTP, gRPC)
│       ├── usecase/              # application logic
│       ├── port/                 # interfaces (inbound + outbound)
│       ├── domain/               # entities, value objects
│       └── adapter/              # driven adapters (DB, messaging)
├── orchestration/
└── pkg/
```

Hexagonal direction:

| from | allowed to import |
|------|-------------------|
| `handler` | `usecase` |
| `usecase` | `port`, `domain` |
| `port` | `domain` |
| `domain` | nothing |
| `adapter` | `port`, `domain` |

### Modular Monolith Layout

```text
internal/
├── domain/
│   └── order/
│       ├── api/                  # module public interface
│       ├── application/          # use cases
│       ├── core/                 # entities, value objects
│       └── infrastructure/       # DB, external services
├── orchestration/
└── pkg/
```

Modular Monolith direction:

| from | allowed to import |
|------|-------------------|
| `api` | `application` |
| `application` | `core` |
| `core` | nothing |
| `infrastructure` | `core` |

### Consumer/Worker Layout (Flat)

Unlike domain-centric presets, the Consumer/Worker preset uses a **flat layout** —
layers live directly under `internal/` with no `domain/` directory.

```text
internal/
├── worker/            # worker_order.go, worker_payment.go
├── service/           # business logic
├── store/             # persistence (DB, external APIs)
├── model/             # data structures
└── pkg/               # shared infra (consumer lib, logging)
    └── consumer/
```

Consumer/Worker direction:

| from | allowed to import |
|------|-------------------|
| `worker` | `service`, `model` |
| `service` | `store`, `model` |
| `store` | `model` |
| `model` | nothing |

All layers may import `pkg/` except `model` (restricted).

**Type pattern enforcement:** Files matching `worker_*.go` in `worker/` must define
a corresponding exported type with a `Process` method:
- `worker_order.go` → must define `OrderWorker` with `Process` method
- `worker_payment.go` → must define `PaymentWorker` with `Process` method

Domain isolation rules are not applicable and are skipped entirely.

### Batch Layout (Flat)

The Batch preset uses the same flat layout as Consumer/Worker, with `job/` as the
entry-point layer for cron/scheduler-triggered batch processing.

```text
internal/
├── job/               # job_expire_files.go, job_cleanup_trash.go
├── service/           # business logic
├── store/             # persistence (DB, external APIs)
├── model/             # data structures
└── pkg/               # shared infra (batchutil, logging)
```

Batch direction:

| from | allowed to import |
|------|-------------------|
| `job` | `service`, `model` |
| `service` | `store`, `model` |
| `store` | `model` |
| `model` | nothing |

All layers may import `pkg/` except `model` (restricted).

**Type pattern enforcement:** Files matching `job_*.go` in `job/` must define
a corresponding exported type with a `Run` method:
- `job_expire_files.go` → must define `ExpireFilesJob` with `Run` method
- `job_cleanup_trash.go` → must define `CleanupTrashJob` with `Run` method

Domain isolation rules are not applicable and are skipped entirely.

### Custom Model

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

## Isolation Rules

`rules.CheckDomainIsolation(pkgs, module, root, opts...)`

Blocks cross-domain imports and forces external access through domain root packages.

| Rule | Meaning |
|------|---------|
| `isolation.cross-domain` | domain A must not import domain B |
| `isolation.cmd-deep-import` | `cmd/` must only import domain roots, not sub-packages |
| `isolation.orchestration-deep-import` | orchestration must only import domain roots |
| `isolation.pkg-imports-domain` | shared pkg must not import any domain |
| `isolation.pkg-imports-orchestration` | shared pkg must not import orchestration |
| `isolation.domain-imports-orchestration` | domains must not import orchestration |
| `isolation.internal-imports-orchestration` | non-cmd/orchestration packages must not import orchestration |
| `isolation.internal-imports-domain` | unregistered internal packages must not import domains |

Import matrix:

| from → to | domain root | domain sub-pkg | orchestration | shared pkg |
|-----------|:-:|:-:|:-:|:-:|
| **same domain** | Yes | Yes | No | Yes |
| **other domain** | No | No | No | Yes |
| **orchestration** | Yes | No | Yes | Yes |
| **cmd** | Yes | No | Yes | Yes |
| **shared pkg** | No | No | No | Yes |

## Layer Direction Rules

`rules.CheckLayerDirection(pkgs, module, root, opts...)`

Enforces allowed intra-domain dependency direction. The direction matrix is defined by the architecture model.

| Rule | Meaning |
|------|---------|
| `layer.direction` | import violates the allowed layer direction |
| `layer.inner-imports-pkg` | inner layer imports shared pkg (controlled by `PkgRestricted`) |
| `layer.unknown-sublayer` | unknown sublayer found in domain |

Notes:

- same-sublayer imports are always allowed
- the domain root package is not checked
- direction matrix is fully customizable via `WithDirection`

## Structure Rules

`rules.CheckStructure(root, opts...)`

| Rule | Meaning |
|------|---------|
| `structure.internal-top-level` | only allowed top-level packages under `internal/` |
| `structure.banned-package` | banned package names (default: `util`, `common`, `misc`, `helper`, `shared`, `services`) |
| `structure.legacy-package` | legacy packages that should be migrated |
| `structure.misplaced-layer` | `app`/`handler`/`infra` outside domain slices |
| `structure.middleware-placement` | `middleware/` must live in shared pkg |
| `structure.domain-root-alias-required` | domain root must define alias file (DDD only) |
| `structure.domain-root-alias-package` | alias file package name must match directory |
| `structure.domain-root-alias-only` | domain root may only contain alias file |
| `structure.domain-alias-no-interface` | alias file must not re-export interfaces |
| `structure.domain-model-required` | domain must have model directory (DDD only) |
| `structure.dto-placement` | DTO files only in handler/app |

## Naming Rules

`rules.CheckNaming(pkgs, opts...)`

| Rule | Meaning |
|------|---------|
| `naming.no-stutter` | exported type repeats the package name |
| `naming.no-impl-suffix` | exported type ends with `Impl` |
| `naming.snake-case-file` | file name is not snake_case |
| `naming.repo-file-interface` | file in repo/ lacks matching interface |
| `naming.no-layer-suffix` | file name redundantly repeats the layer name |
| `naming.domain-interface-repo-only` | domain interface outside repo sublayer (DDD only) |
| `naming.no-handmock` | test file defines hand-rolled mock/fake/stub |
| `naming.worker-type-mismatch` | `worker_*.go` file must define matching type (ConsumerWorker only) |
| `naming.worker-missing-process` | worker type must have `Process` method (ConsumerWorker only) |

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
