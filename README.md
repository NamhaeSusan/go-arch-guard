# go-arch-guard

Architecture rule enforcement for Go projects via `go test`.

Define domain isolation, layer direction, naming conventions, and structural rules — then enforce them as regular test failures in CI. No CLI tools to learn, no config files to maintain. Just Go tests.

## Why

Architecture rules decay silently. A developer adds one cross-domain import, code review misses it, and six months later your "clean architecture" is spaghetti. go-arch-guard catches violations **at test time**, so `go test` fails before bad imports land in main.

---

## Target Architecture

go-arch-guard enforces a **domain-centric vertical slice** architecture. Each domain owns its complete stack — from HTTP handler to DB persistence. Cross-domain coordination happens through a dedicated orchestration layer.

### Project Layout

```
cmd/
└── api/
    ├── main.go                         ← process entry point
    ├── wire.go                         ← app-specific dependency wiring
    └── routes.go                       ← route registration via domain root APIs

internal/
├── domain/
│   ├── order/                          ← one domain = one vertical slice
│   │   ├── alias.go                    ← public API surface for the domain root package
│   │   │
│   │   ├── app/                        ← application service (facade)
│   │   │   └── service.go              ← coordinates core/* layers
│   │   │
│   │   ├── core/                       ← domain core (pure business logic)
│   │   │   ├── model/                  ← entities, value objects
│   │   │   │   └── order.go            ← NO dependencies on other layers
│   │   │   ├── repo/                   ← repository interfaces
│   │   │   │   └── repository.go       ← depends on model/ only
│   │   │   └── svc/                    ← domain services (stateless logic)
│   │   │       └── order.go            ← depends on model/ only
│   │   │
│   │   ├── event/                      ← domain events (immutable facts)
│   │   │   └── events.go               ← depends on model/ only
│   │   │
│   │   ├── handler/                    ← inbound adapter (driving)
│   │   │   └── http/
│   │   │       └── handler.go          ← calls app/ only, never core/ directly
│   │   │
│   │   └── infra/                      ← outbound adapter (driven)
│   │       └── persistence/
│   │           └── store.go            ← implements repo/ interfaces
│   │
│   └── user/
│       └── ...                          ← same structure, every domain is identical
│
├── orchestration/                       ← cross-domain coordination
│   ├── handler/http/                    ← cross-domain API endpoints
│   │   └── handler.go
│   ├── create_order.go                  ← imports domain aliases ONLY
│   └── draft_submit.go
│
└── pkg/                                 ← shared utilities (domain-unaware)
    ├── middleware/                       ← auth, rate limiting
    └── transport/http/                  ← shared response/error helpers
```

### alias.go — The Domain Root's Public Surface

Each domain exposes exactly **one package** to the outside: the domain root package. `alias.go` defines the public surface for that root package by re-exporting only what external consumers need:

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

**Why:** Outside code imports `order.Service` or `order.Handler`, never `order/app.Service` or `order/handler/http.Handler`. If you refactor the internals, the root package stays stable.

**Rule:** Outside code can import only the domain root package. Deep imports such as `domain/order/handler/http` or `domain/order/core/model` are violations. `alias.go` itself is the publication file, not a layer-direction target.

### core/ — The Dependency-Free Center

```
core/
├── model/    ← entities (depends on NOTHING)
├── repo/     ← interfaces (depends on model only)
└── svc/      ← pure logic (depends on model only)
```

**Key constraint:** `core/svc/` depends on `core/model/` only — never `core/repo/`. The service layer doesn't know *how* data is stored, only *what* the data looks like.

`core/repo/` defines interfaces. `infra/persistence/` implements them. This is the dependency inversion boundary.

Inner layers stay `internal/pkg`-free. `core`, `core/model`, `core/repo`, `core/svc`, and `event` must not import shared support packages directly.

### orchestration/ — Cross-Domain Flows

When a use case spans multiple domains (e.g., "submit a draft creates a review and notifies users"), the orchestration layer coordinates:

```go
// internal/orchestration/draft_submit.go
package orchestration

import (
    "mymodule/internal/domain/draft"    // alias only
    "mymodule/internal/domain/review"   // alias only
    "mymodule/internal/domain/user"     // alias only
)

type DraftSubmit struct {
    draftSvc  draft.Service
    reviewSvc review.Service
    userSvc   user.Service
}
```

**Rule:** Orchestration can only import domain **aliases** (root packages). Importing `domain/user/core/model` directly is a violation. Outside of `cmd/...`, no other non-orchestration package may depend on `internal/orchestration/...`.

### cmd/ — Composition Root

`cmd/...` is the only place that should wire applications together: create services, build handlers through domain root APIs, register routes, and start processes.

**Rule:** `cmd/...` can import domain root packages, but it must not import domain sub-packages directly.

### Dependency Direction (Full Picture)

```
                    ┌─────────────────────────────────┐
                    │            cmd/                  │
                    │  (entrypoint + wiring + routes)  │
                    └──────────┬──────────────────────┘
                               │ creates
                    ┌──────────▼──────────────────────┐
                    │      orchestration/              │
                    │   (cross-domain coordination)    │
                    └──────────┬──────────────────────┘
                               │ imports alias only
            ┌──────────────────▼──────────────────────┐
            │            domain/{name}/                │
            │                                          │
            │   handler/ ──→ app/ ──→ core/svc/        │
            │                 │          │             │
            │                 ├──→ core/repo/ ──┐      │
            │                 │                 │      │
            │   infra/ ──→ core/repo/ ──→ core/model/  │
            │                                          │
            │   event/ ──→ core/model/                 │
            │                                          │
            │   alias.go ──→ selected internal APIs    │
            └──────────────────────────────────────────┘
                               │
                    ┌──────────▼──────────────────────┐
                    │        internal/pkg/             │
                    │   (shared utilities, imports      │
                    │    neither domains nor            │
                    │    orchestration)                 │
                    └─────────────────────────────────┘
```

---

## Install

```bash
go get github.com/NamhaeSusan/go-arch-guard
```

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
    pkgs, err := analyzer.Load(".", "internal/...", "cmd/...")
    if err != nil {
        t.Fatal(err)
    }

    t.Run("domain isolation", func(t *testing.T) {
        report.AssertNoViolations(t, rules.CheckDomainIsolation(pkgs, "github.com/yourmodule", "."))
    })
    t.Run("layer direction", func(t *testing.T) {
        report.AssertNoViolations(t, rules.CheckLayerDirection(pkgs, "github.com/yourmodule", "."))
    })
    t.Run("naming", func(t *testing.T) {
        report.AssertNoViolations(t, rules.CheckNaming(pkgs))
    })
    t.Run("structure", func(t *testing.T) {
        report.AssertNoViolations(t, rules.CheckStructure("."))
    })
}
```

Run: `go test -run TestArchitecture -v`

---

## Rules

### Domain Isolation (`rules.CheckDomainIsolation`)

**Purpose:** Domains must not know about each other, and external packages must depend on a domain through its root package only.

#### Import Matrix

| from | to | allowed? |
|------|----|----------|
| same domain | same domain | Yes |
| anyone | `internal/pkg/` | Yes |
| `internal/pkg/` | any domain | **No** — `isolation.pkg-imports-domain` |
| `internal/pkg/` | `orchestration/` | **No** — `isolation.pkg-imports-orchestration` |
| `orchestration/` | domain alias (root package) | Yes |
| `orchestration/` | domain sub-package | **No** — `isolation.orchestration-deep-import` |
| `cmd/` | `orchestration/` | Yes |
| `cmd/` | domain alias (root package) | Yes |
| `cmd/` | domain sub-package | **No** — `isolation.cmd-deep-import` |
| domain | `orchestration/` | **No** — `isolation.domain-imports-orchestration` |
| other internal package | `orchestration/` | **No** — `isolation.internal-imports-orchestration` |
| other internal package | any domain | **No** — `isolation.internal-imports-domain` |
| domain A | domain B | **No** — `isolation.cross-domain` |

#### Examples

```go
// ✅ orchestration imports domain via alias
import "mymodule/internal/domain/user"         // alias.go
import "mymodule/internal/domain/order"        // alias.go

// ❌ orchestration imports domain internals
import "mymodule/internal/domain/user/core/model"   // isolation.orchestration-deep-import

// ❌ order handler imports user domain
import "mymodule/internal/domain/user/app"          // isolation.cross-domain

// ❌ config imports a domain directly
import "mymodule/internal/domain/user"              // isolation.internal-imports-domain

// ❌ shared package imports orchestration
import "mymodule/internal/orchestration"            // isolation.pkg-imports-orchestration

// ❌ domain imports orchestration
import "mymodule/internal/orchestration"            // isolation.domain-imports-orchestration

// ❌ config imports orchestration directly
import "mymodule/internal/orchestration"            // isolation.internal-imports-orchestration

// ✅ anyone imports shared utilities
import "mymodule/internal/pkg/db"
```

---

### Layer Direction (`rules.CheckLayerDirection`)

**Purpose:** Within a single domain, enforce the dependency direction. Inner layers must not depend on outer layers.

#### Allowed Imports

| from | allowed to import |
|------|-------------------|
| `handler` | `app` |
| `app` | `core/model`, `core/repo`, `core/svc`, `event` |
| `core/svc` | `core/model` |
| `core/repo` | `core/model` |
| `core` | `core/model` |
| `infra` | `core/repo`, `core/model`, `event` |
| `event` | `core/model` |
| `core/model` | (nothing) |

`alias.go` is the root package's publication file and is not checked by `CheckLayerDirection`.

Same-sublayer imports (e.g., `handler/http` → `handler/grpc`) are always allowed.
Packages under `internal/domain/<name>/` must use one of the known sublayers only: `handler`, `app`, `core`, `core/model`, `core/repo`, `core/svc`, `event`, `infra`. Any other sublayer is rejected as `layer.unknown-sublayer`.
`core`, `core/model`, `core/repo`, `core/svc`, and `event` must not import `internal/pkg/...`; that is reported as `layer.inner-imports-pkg`.

#### Examples

```go
// ✅ app imports core layers
import "mymodule/internal/domain/order/core/model"
import "mymodule/internal/domain/order/core/repo"
import "mymodule/internal/domain/order/core/svc"

// ❌ core/svc imports core/repo (svc should only know model)
import "mymodule/internal/domain/order/core/repo"    // layer.direction

// ❌ handler imports infra directly
import "mymodule/internal/domain/order/infra/persistence"  // layer.direction

// ❌ event imports internal/pkg directly
import "mymodule/internal/pkg/clock"  // layer.inner-imports-pkg

```

---

### Naming (`rules.CheckNaming`)

| Rule | Bad | Good | Violation |
|------|-----|------|-----------|
| No stutter | `user.UserService` | `user.Service` | `naming.no-stutter` |
| No Impl suffix | `ServiceImpl` | `service` (unexported) | `naming.no-impl-suffix` |
| Snake case files | `userService.go` | `user_service.go` | `naming.snake-case-file` |
| Repo file interface | `repo/user.go` without `User` interface | `repo/user.go` with `type User interface` | `naming.repo-file-interface` |
| No layer suffix | `svc/install_svc.go` | `svc/install.go` | `naming.no-layer-suffix` |

---

### Structure (`rules.CheckStructure`)

| Rule | Description | Violation |
|------|-------------|-----------|
| Banned packages | `util`, `common`, `misc`, `helper`, `shared` anywhere under `internal/` | `structure.banned-package` |
| Legacy packages | `router`, `bootstrap`, or misplaced `app`/`handler`/`infra` anywhere under `internal/` | `structure.legacy-package` |
| Middleware placement | `middleware/` must be under `pkg/` | `structure.middleware-placement` |
| Domain root alias required | each domain root must define `alias.go` | `structure.domain-root-alias-required` |
| Domain root alias package | `alias.go` package name must match the domain root name | `structure.domain-root-alias-package` |
| Domain root alias only | each domain root may contain only `alias.go` as a non-test Go file | `structure.domain-root-alias-only` |
| Domain model required | each domain must have a non-empty `core/model/` | `structure.domain-model-required` |
| DTO placement | `dto.go` must not be in `domain/` or `infra/` | `structure.dto-placement` |

---

## Options

### Severity

```go
// Warnings: violations are logged but test passes
rules.CheckDomainIsolation(pkgs, module, root, rules.WithSeverity(rules.Warning))

// Errors (default): violations fail the test
rules.CheckDomainIsolation(pkgs, module, root)
```

### Exclude Paths

```go
// Exclude a legacy subtree from isolation checks
rules.CheckDomainIsolation(pkgs, module, root,
    rules.WithExclude("internal/legacy/..."))

// Exclude specific domains from naming checks
rules.CheckNaming(pkgs, rules.WithExclude("internal/domain/auth/..."))
```

Pattern `...` matches the root and all sub-paths. Project-relative paths are the preferred format; module-qualified paths are still accepted for backward compatibility.

---

## Gradual Adoption

### Step 1: Warning mode

```go
violations := rules.CheckDomainIsolation(pkgs, module, ".",
    rules.WithSeverity(rules.Warning))
report.AssertNoViolations(t, violations)  // passes, but logs violations
```

### Step 2: Exclude legacy, enforce new code

```go
rules.CheckDomainIsolation(pkgs, module, ".",
    rules.WithExclude("internal/old_handler/..."))
```

### Step 3: Migrate and remove excludes

As you migrate legacy code, remove exclude patterns until everything is enforced.

---

## Domain Variants

### Full Domain

```
internal/domain/review/
├── alias.go
├── app/service.go
├── core/
│   ├── model/review.go, view.go
│   ├── repo/repository.go
│   └── svc/review.go
├── event/events.go
├── handler/http/handler.go
└── infra/persistence/repository.go
```

### Thin Domain (no handler, no svc)

```
internal/domain/audit/
├── alias.go
├── core/
│   ├── model/audit.go
│   └── repo/repository.go
└── infra/persistence/repository.go
```

## API Reference

| Function | Description |
|----------|-------------|
| `analyzer.Load(dir, patterns...)` | Load Go packages for analysis |
| `rules.CheckDomainIsolation(pkgs, module, root, opts...)` | Cross-domain isolation |
| `rules.CheckLayerDirection(pkgs, module, root, opts...)` | Intra-domain layer direction |
| `rules.CheckNaming(pkgs, opts...)` | Naming conventions |
| `rules.CheckStructure(root, opts...)` | Directory structure |
| `report.AssertNoViolations(t, violations)` | Fail test on Error violations |
| `rules.WithSeverity(rules.Warning)` | Degrade violations to warnings |
| `rules.WithExclude("internal/path/...")` | Skip matching packages or files |

---

## Violation Rules Reference

| Rule | Function | Description |
|------|----------|-------------|
| `isolation.cross-domain` | CheckDomainIsolation | Domain A imports domain B |
| `isolation.internal-imports-domain` | CheckDomainIsolation | Unauthorized internal package imports a domain |
| `isolation.orchestration-deep-import` | CheckDomainIsolation | Orchestration imports domain sub-package |
| `isolation.cmd-deep-import` | CheckDomainIsolation | cmd/ imports domain sub-package |
| `isolation.pkg-imports-domain` | CheckDomainIsolation | pkg/ imports a domain |
| `isolation.pkg-imports-orchestration` | CheckDomainIsolation | pkg/ imports orchestration |
| `isolation.domain-imports-orchestration` | CheckDomainIsolation | Domain imports orchestration |
| `isolation.internal-imports-orchestration` | CheckDomainIsolation | Unauthorized internal package imports orchestration |
| `layer.direction` | CheckLayerDirection | Wrong layer direction within domain |
| `layer.inner-imports-pkg` | CheckLayerDirection | Inner domain layer imports internal/pkg |
| `layer.unknown-sublayer` | CheckLayerDirection | Package uses an unsupported domain sublayer |
| `naming.no-stutter` | CheckNaming | Type name repeats package name |
| `naming.no-impl-suffix` | CheckNaming | Type ends with "Impl" |
| `naming.snake-case-file` | CheckNaming | Filename not snake_case |
| `naming.repo-file-interface` | CheckNaming | repo/ file missing matching interface |
| `naming.no-layer-suffix` | CheckNaming | Filename has redundant layer suffix |
| `naming.handler-no-exported-interface` | CheckNaming | handler package defines exported interface |
| `structure.banned-package` | CheckStructure | Package uses a banned util/common/misc/helper/shared name |
| `structure.legacy-package` | CheckStructure | Package uses router/bootstrap or a misplaced app/handler/infra directory |
| `structure.middleware-placement` | CheckStructure | Middleware not in pkg/ |
| `structure.domain-root-alias-required` | CheckStructure | Domain root is missing alias.go |
| `structure.domain-root-alias-package` | CheckStructure | alias.go package name does not match the domain root |
| `structure.domain-root-alias-only` | CheckStructure | Domain root contains files other than alias.go |
| `structure.domain-model-required` | CheckStructure | Domain missing non-empty core/model/ |
| `structure.dto-placement` | CheckStructure | dto.go in domain/ or infra/ |

## License

MIT
