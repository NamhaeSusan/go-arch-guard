# go-arch-guard

Architecture rule enforcement for Go projects via `go test`.

Define domain isolation, layer direction, naming conventions, and structural rules — then enforce them as regular test failures in CI. No CLI tools to learn, no config files to maintain. Just Go tests.

## Why

Architecture rules decay silently. A developer adds one cross-domain import, code review misses it, and six months later your "clean architecture" is spaghetti. go-arch-guard catches violations **at test time**, so `go test` fails before bad imports land in main.

## Target Architecture

go-arch-guard enforces a **domain-centric vertical slice** architecture:

```
internal/
├── domain/
│   ├── order/                    ← each domain owns its full stack
│   │   ├── alias.go             ← external entry point (re-exports app/)
│   │   ├── app/service.go       ← facade (coordinates core/*)
│   │   ├── core/
│   │   │   ├── model/order.go   ← entities (no dependencies)
│   │   │   ├── repo/repository.go ← interfaces (depends on model only)
│   │   │   └── svc/order.go     ← pure logic (depends on model only)
│   │   ├── event/events.go      ← domain events (immutable)
│   │   ├── handler/http/        ← HTTP handler (calls app/ only)
│   │   └── infra/persistence/   ← DB implementation (implements repo/)
│   └── user/
│       └── ...                   ← same structure
├── saga/                         ← cross-domain orchestration
│   ├── handler/http/             ← cross-domain API endpoints
│   ├── create_order.go           ← imports domain aliases only
│   └── draft_submit.go
└── pkg/                          ← shared utilities (anyone can import)
```

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

**Purpose:** Domains must not know about each other. Cross-domain coordination goes through `saga/` only.

#### Import Matrix

| from | to | allowed? |
|------|----|----------|
| same domain | same domain | Yes |
| anyone | `pkg/` | Yes |
| `pkg/` | any domain | **No** — `isolation.pkg-imports-domain` |
| `saga/` | domain alias (root package) | Yes |
| `saga/` | domain sub-package | **No** — `isolation.saga-deep-import` |
| `saga/handler/` | saga internals | Yes |
| domain A | domain B | **No** — `isolation.cross-domain` |

#### Examples

```go
// ✅ saga imports domain via alias (root package)
import "mymodule/internal/domain/user"        // alias.go
import "mymodule/internal/domain/order"       // alias.go

// ❌ saga imports domain internals
import "mymodule/internal/domain/user/core/model"  // isolation.saga-deep-import

// ❌ order handler imports user domain
import "mymodule/internal/domain/user/app"    // isolation.cross-domain

// ✅ anyone imports shared utilities
import "mymodule/internal/pkg/db"
```

#### Violation Output

```
[ERROR] violation: domain "order" imports domain "user" sub-package directly
  (file: internal/domain/order/handler/http/handler.go:5,
   rule: isolation.cross-domain,
   fix: use saga/ for cross-domain coordination)
```

---

### Layer Direction (`rules.CheckLayerDirection`)

**Purpose:** Within a single domain, enforce the dependency direction. Inner layers must not depend on outer layers.

#### Dependency DAG

```
handler/  ──→  app/  ──→  core/svc/  ──→  core/model/
                 │                              ↑
                 ├──→  core/repo/  ─────────────┘
                 │
infra/  ────→  core/repo/  ──→  core/model/

event/  ────→  core/model/

alias.go  ──→  app/
```

#### Allowed Imports Table

| from | allowed to import |
|------|-------------------|
| `""` (alias.go) | `app` |
| `handler` | `app` |
| `app` | `core/model`, `core/repo`, `core/svc` |
| `core/svc` | `core/model` |
| `core/repo` | `core/model` |
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
import "mymodule/internal/domain/order/core/repo"  // layer.direction

// ❌ handler imports infra directly
import "mymodule/internal/domain/order/infra/persistence"  // layer.direction

// ❌ alias.go imports core/model directly (should only import app)
import "mymodule/internal/domain/order/core/model"  // layer.direction
```

#### Violation Output

```
[ERROR] violation: "core/svc" imports "core/repo" — not allowed
  (file: internal/domain/order/core/svc/order.go:4,
   rule: layer.direction,
   fix: core/svc may only import core/model)
```

---

### Naming (`rules.CheckNaming`)

| Rule | Bad | Good | Violation |
|------|-----|------|-----------|
| No stutter | `user.UserService` | `user.Service` | `naming.no-stutter` |
| No Impl suffix | `ServiceImpl` | `Service` (unexported) | `naming.no-impl-suffix` |
| Snake case files | `userService.go` | `user_service.go` | `naming.snake-case-file` |
| Repo file interface | `repo/user.go` without `User` interface | `repo/user.go` with `type User interface` | `naming.repo-file-interface` |
| No layer suffix | `svc/install_svc.go` | `svc/install.go` | `naming.no-layer-suffix` |

#### Repo File Interface

Files in `repo/` directories must contain an exported interface matching the filename:

```
repo/user.go        → must contain: type User interface { ... }
repo/order_item.go  → must contain: type OrderItem interface { ... }
repo/repository.go  → must contain: type Repository interface { ... }
```

#### No Layer Suffix

Filenames must not repeat the layer name. The directory already tells you the layer:

```
svc/install_svc.go       ❌  →  svc/install.go       ✅
repo/user_repo.go        ❌  →  repo/user.go         ✅
handler/order_handler.go ❌  →  handler/order.go     ✅
model/user_model.go      ❌  →  model/user.go        ✅
```

Banned suffixes: `_svc`, `_service`, `_repo`, `_repository`, `_handler`, `_controller`, `_model`, `_entity`, `_store`, `_persistence`

---

### Structure (`rules.CheckStructure`)

| Rule | Description | Violation |
|------|-------------|-----------|
| Banned packages | `util`, `common`, `misc`, `helper`, `shared` under `internal/` | `structure.banned-package` |
| Domain model required | each `internal/domain/<name>/` must have `model.go` | `structure.domain-model-required` |
| DTO placement | `dto.go` must not be in `domain/` or `infra/` | `structure.dto-placement` |

---

## Options

### Severity

By default, violations are `Error` (test fails). Use `Warning` for gradual adoption:

```go
// Warnings: violations are logged but test passes
rules.CheckDomainIsolation(pkgs, module, root, rules.WithSeverity(rules.Warning))

// Errors (default): violations fail the test
rules.CheckDomainIsolation(pkgs, module, root)
```

### Exclude Paths

Skip specific paths during migration:

```go
// Exclude legacy packages from isolation checks
rules.CheckDomainIsolation(pkgs, module, root,
    rules.WithExclude("internal/legacy/..."))

// Exclude specific domains from naming checks
rules.CheckNaming(pkgs, rules.WithExclude("internal/domain/auth/..."))
```

Pattern `...` matches all sub-paths (e.g., `internal/legacy/...` matches `internal/legacy/foo/bar`).

---

## Gradual Adoption

### Step 1: Warning mode (see what breaks)

```go
func TestArchitecture(t *testing.T) {
    pkgs, _ := analyzer.Load(".", "internal/...")

    violations := rules.CheckDomainIsolation(pkgs, module, ".",
        rules.WithSeverity(rules.Warning))
    report.AssertNoViolations(t, violations)  // passes, but logs violations
}
```

### Step 2: Exclude legacy, enforce new code

```go
rules.CheckDomainIsolation(pkgs, module, ".",
    rules.WithExclude("internal/old_handler/...", "internal/old_infra/..."))
```

### Step 3: Remove excludes one by one

As you migrate legacy code, remove exclude patterns until everything is enforced.

---

## Domain Structure Reference

### Full Domain (e.g., review)

```
internal/domain/review/
├── alias.go                  ← re-exports app.Service
├── app/
│   └── service.go            ← public facade, coordinates core/*
├── core/
│   ├── model/
│   │   ├── review.go         ← entities
│   │   └── view.go           ← read models
│   ├── repo/
│   │   └── repository.go     ← repository interface
│   └── svc/
│       └── review.go         ← pure business logic (no I/O)
├── event/
│   └── events.go             ← domain events (skeleton)
├── handler/
│   └── http/
│       ├── handler.go        ← HTTP handler + RegisterRoutes()
│       └── response.go       ← HTTP DTOs
└── infra/
    └── persistence/
        └── repository.go     ← repo interface implementation
```

### Thin Domain (e.g., audit)

Only what's needed:

```
internal/domain/audit/
├── alias.go
├── core/
│   ├── model/audit.go
│   └── repo/repository.go
└── infra/
    └── persistence/repository.go
```

### Handler-Only Domain (e.g., dashboard)

No service layer — calls saga/ for cross-domain queries:

```
internal/domain/dashboard/
├── alias.go
└── handler/http/handler.go
```

### Saga (Cross-Domain Orchestration)

```
internal/saga/
├── handler/http/handler.go      ← cross-domain API endpoints
├── draft_submit.go              ← draft → review conversion
├── contract_create.go           ← review → contract creation
└── dashboard_query.go           ← multi-domain data assembly
```

Each saga struct depends on multiple domain aliases:

```go
type DraftSubmit struct {
    draftSvc  draft.Service    // via alias.go
    reviewSvc review.Service   // via alias.go
    userSvc   user.Service     // via alias.go
}
```

---

## API Reference

### `analyzer.Load(dir, patterns...)`

Loads Go packages for analysis.

```go
pkgs, err := analyzer.Load(".", "internal/...")
```

### `rules.CheckDomainIsolation(pkgs, module, root, opts...)`

Checks cross-domain isolation. Returns `[]Violation`.

### `rules.CheckLayerDirection(pkgs, module, root, opts...)`

Checks intra-domain layer direction. Returns `[]Violation`.

### `rules.CheckNaming(pkgs, opts...)`

Checks naming conventions. Returns `[]Violation`.

### `rules.CheckStructure(root, opts...)`

Checks directory structure. Returns `[]Violation`.

### `report.AssertNoViolations(t, violations)`

Logs all violations. Fails the test only if any `Error`-level violations exist. `Warning`-level violations are logged but don't fail.

### Options

| Option | Description |
|--------|-------------|
| `rules.WithSeverity(rules.Warning)` | Set all violations to Warning (test passes) |
| `rules.WithSeverity(rules.Error)` | Set all violations to Error (default, test fails) |
| `rules.WithExclude("path/...")` | Skip packages matching the pattern |

---

## Violation Rules Reference

| Rule | Function | Severity | Description |
|------|----------|----------|-------------|
| `isolation.cross-domain` | CheckDomainIsolation | Error | Domain A imports domain B |
| `isolation.saga-deep-import` | CheckDomainIsolation | Error | Saga imports domain sub-package instead of alias |
| `isolation.pkg-imports-domain` | CheckDomainIsolation | Error | pkg/ imports a domain |
| `layer.direction` | CheckLayerDirection | Error | Intra-domain layer imports in wrong direction |
| `naming.no-stutter` | CheckNaming | Error | Type name repeats package name |
| `naming.no-impl-suffix` | CheckNaming | Error | Type ends with "Impl" |
| `naming.snake-case-file` | CheckNaming | Error | Filename is not snake_case |
| `naming.repo-file-interface` | CheckNaming | Error | repo/ file missing matching interface |
| `naming.no-layer-suffix` | CheckNaming | Error | Filename has redundant layer suffix |
| `structure.banned-package` | CheckStructure | Error | Package name is util/common/misc/helper/shared |
| `structure.domain-model-required` | CheckStructure | Error | Domain missing model.go |
| `structure.dto-placement` | CheckStructure | Error | dto.go in domain/ or infra/ |

## License

MIT
