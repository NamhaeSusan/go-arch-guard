# go-arch-guard

[![CI](https://github.com/NamhaeSusan/go-arch-guard/actions/workflows/ci.yml/badge.svg)](https://github.com/NamhaeSusan/go-arch-guard/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/NamhaeSusan/go-arch-guard/branch/main/graph/badge.svg)](https://codecov.io/gh/NamhaeSusan/go-arch-guard)
[![Go Report Card](https://goreportcard.com/badge/github.com/NamhaeSusan/go-arch-guard)](https://goreportcard.com/report/github.com/NamhaeSusan/go-arch-guard)

[한국어](README.ko.md)

Architecture guardrails for Go projects via `go test`.

Define isolation, layer-direction, structure, naming, and blast-radius rules, then fail regular tests when the project shape drifts. Ships with **DDD** and **Clean Architecture** presets, and supports fully custom architecture models. No CLI to learn. No separate config format. Just Go tests.

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

## Quick Start

### DDD (default)

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

### Clean Architecture

```go
func TestArchitecture(t *testing.T) {
    pkgs, err := analyzer.Load(".", "internal/...", "cmd/...")
    if err != nil {
        t.Log(err)
    }
    if len(pkgs) == 0 {
        t.Fatalf("no packages loaded: %v", err)
    }

    m := rules.CleanArch()
    opts := []rules.Option{rules.WithModel(m)}

    t.Run("domain isolation", func(t *testing.T) {
        report.AssertNoViolations(t, rules.CheckDomainIsolation(pkgs, "", "", opts...))
    })
    t.Run("layer direction", func(t *testing.T) {
        report.AssertNoViolations(t, rules.CheckLayerDirection(pkgs, "", "", opts...))
    })
    t.Run("naming", func(t *testing.T) {
        report.AssertNoViolations(t, rules.CheckNaming(pkgs, opts...))
    })
    t.Run("structure", func(t *testing.T) {
        report.AssertNoViolations(t, rules.CheckStructure(".", opts...))
    })
}
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
| `report.AssertNoViolations(t, violations)` | fail test on Error violations |
| `rules.DDD()` | DDD architecture model (default) |
| `rules.CleanArch()` | Clean Architecture model |
| `rules.NewModel(opts...)` | custom model builder |
| `rules.WithModel(m)` | apply custom model to checks |
| `rules.WithSeverity(rules.Warning)` | downgrade to warnings |
| `rules.WithExclude("path/...")` | skip a subtree |

## Claude Code Plugin

```text
/plugin marketplace add NamhaeSusan/go-arch-guard
/plugin install go-arch-guard@go-arch-guard-marketplace
```

## External Import Hygiene

`go-arch-guard` checks **project-internal** imports only. External dependency hygiene should be enforced via AI tool instructions and code review. See the [DDD external import constraints](README.ko.md#외부-import-위생--이-라이브러리가-아닌-ai-도구-지침으로-강제) for a copy-paste template.

## License

MIT
