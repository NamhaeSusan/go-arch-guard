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
internal/
├── domain/
│   ├── order/                          ← one domain = one vertical slice
│   │   ├── alias.go                    ← the ONLY external entry point
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
├── router/                              ← HTTP route registration + DI wiring
│   ├── router.go                        ← gin engine setup
│   ├── routes.go                        ← route definitions
│   └── error_handler.go                 ← domain error → HTTP status mapping
│
└── pkg/                                 ← shared utilities (anyone can import)
    ├── middleware/                       ← auth, rate limiting
    └── transport/http/                  ← shared response/error helpers
```

### alias.go — The Domain's Public API

Each domain exposes exactly **one entry point**: `alias.go` at the domain root. It uses Go type aliases to re-export only what external consumers need:

```go
// internal/domain/order/alias.go
package order

import "mymodule/internal/domain/order/app"

type Service = app.Service
```

**Why:** Outside code imports `order.Service`, never `order/app.Service` or `order/core/model.Order`. If you refactor the internals, the alias stays stable.

**Rule:** `alias.go` can only import `app/`. Direct imports of `core/model/`, `core/repo/`, etc. are violations.

### core/ — The Dependency-Free Center

```
core/
├── model/    ← entities (depends on NOTHING)
├── repo/     ← interfaces (depends on model only)
└── svc/      ← pure logic (depends on model only)
```

**Key constraint:** `core/svc/` depends on `core/model/` only — never `core/repo/`. The service layer doesn't know *how* data is stored, only *what* the data looks like.

`core/repo/` defines interfaces. `infra/persistence/` implements them. This is the dependency inversion boundary.

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

**Rule:** Orchestration can only import domain **aliases** (root packages). Importing `domain/user/core/model` directly is a violation.

### Dependency Direction (Full Picture)

```
                    ┌─────────────────────────────────┐
                    │          router/                 │
                    │   (DI wiring + route setup)      │
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
            │   alias.go ──→ app/                      │
            └──────────────────────────────────────────┘
                               │
                    ┌──────────▼──────────────────────┐
                    │           pkg/                   │
                    │   (shared utilities — anyone     │
                    │    can import, imports nothing)   │
                    └─────────────────────────────────┘
```

---

## Install

```bash
go get github.com/kimtaeyun/go-arch-guard
```

## Quick Start

Create `architecture_test.go` in your project root:

```go
package myproject_test

import (
    "testing"

    "github.com/kimtaeyun/go-arch-guard/analyzer"
    "github.com/kimtaeyun/go-arch-guard/report"
    "github.com/kimtaeyun/go-arch-guard/rules"
)

func TestArchitecture(t *testing.T) {
    pkgs, err := analyzer.Load(".", "internal/...")
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

**Purpose:** Domains must not know about each other. Cross-domain coordination goes through `orchestration/` only.

#### Import Matrix

| from | to | allowed? |
|------|----|----------|
| same domain | same domain | Yes |
| anyone | `pkg/` | Yes |
| `pkg/` | any domain | **No** — `isolation.pkg-imports-domain` |
| `orchestration/` | domain alias (root package) | Yes |
| `orchestration/` | domain sub-package | **No** — `isolation.orchestration-deep-import` |
| `router/` | domain alias (root package) | Yes |
| `router/` | domain sub-package | **No** — `isolation.router-deep-import` |
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

// ✅ anyone imports shared utilities
import "mymodule/internal/pkg/db"
```

---

### Layer Direction (`rules.CheckLayerDirection`)

**Purpose:** Within a single domain, enforce the dependency direction. Inner layers must not depend on outer layers.

#### Allowed Imports

| from | allowed to import |
|------|-------------------|
| `""` (alias.go) | `app` |
| `handler` | `app` |
| `app` | `core/model`, `core/repo`, `core/svc` |
| `core/svc` | `core/model` |
| `core/repo` | `core/model` |
| `core` | `core/model` |
| `infra` | `core/repo`, `core/model` |
| `event` | `core/model` |
| `core/model` | (nothing) |

Same-sublayer imports (e.g., `handler/http` → `handler/grpc`) are always allowed.

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

// ❌ alias.go imports core/model directly (should only import app)
import "mymodule/internal/domain/order/core/model"   // layer.direction
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
| Banned packages | `util`, `common`, `misc`, `helper`, `shared` under `internal/` | `structure.banned-package` |
| Legacy packages | `handler`, `app`, `infra` at `internal/` top level | `structure.legacy-package` |
| Middleware placement | `middleware/` must be under `pkg/` | `structure.middleware-placement` |
| Domain model required | each domain must have `model.go` or `core/model/` | `structure.domain-model-required` |
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
// Exclude DI wiring layer from isolation checks
rules.CheckDomainIsolation(pkgs, module, root,
    rules.WithExclude("mymodule/internal/router", "mymodule/internal/router/..."))

// Exclude specific domains from naming checks
rules.CheckNaming(pkgs, rules.WithExclude("internal/domain/auth/..."))
```

Pattern `...` matches all sub-paths.

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

### Handler-Only Domain

```
internal/domain/dashboard/
├── alias.go
└── handler/http/handler.go
```

---

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
| `rules.WithExclude("path/...")` | Skip matching packages |

---

## Violation Rules Reference

| Rule | Function | Description |
|------|----------|-------------|
| `isolation.cross-domain` | CheckDomainIsolation | Domain A imports domain B |
| `isolation.orchestration-deep-import` | CheckDomainIsolation | Orchestration imports domain sub-package |
| `isolation.router-deep-import` | CheckDomainIsolation | Router imports domain sub-package |
| `isolation.pkg-imports-domain` | CheckDomainIsolation | pkg/ imports a domain |
| `layer.direction` | CheckLayerDirection | Wrong layer direction within domain |
| `naming.no-stutter` | CheckNaming | Type name repeats package name |
| `naming.no-impl-suffix` | CheckNaming | Type ends with "Impl" |
| `naming.snake-case-file` | CheckNaming | Filename not snake_case |
| `naming.repo-file-interface` | CheckNaming | repo/ file missing matching interface |
| `naming.no-layer-suffix` | CheckNaming | Filename has redundant layer suffix |
| `structure.banned-package` | CheckStructure | Package is util/common/misc/helper/shared |
| `structure.legacy-package` | CheckStructure | Legacy handler/app/infra at top level |
| `structure.middleware-placement` | CheckStructure | Middleware not in pkg/ |
| `structure.domain-model-required` | CheckStructure | Domain missing model.go or core/model/ |
| `structure.dto-placement` | CheckStructure | dto.go in domain/ or infra/ |

## License

MIT
