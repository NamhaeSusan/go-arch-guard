# Vertical Slice Rules Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `CheckVerticalSlice()` and `CheckVerticalSliceInternal()` to go-arch-guard for validating vertical slice (domain-per-directory) architecture.

**Architecture:** Two new functions in `rules/` package. `CheckVerticalSlice` enforces HARD cross-domain isolation rules. `CheckVerticalSliceInternal` enforces SOFT intra-domain layer direction rules. Both use existing `Violation`, `Config`, `Option` types. New testdata directories (`testdata/vertical-valid`, `testdata/vertical-invalid`) for integration tests.

**Tech Stack:** Go, golang.org/x/tools/go/packages

---

## File Structure

```
rules/
├── vertical.go              ← CheckVerticalSlice (HARD: cross-domain isolation)
├── vertical_internal.go     ← CheckVerticalSliceInternal (SOFT: intra-domain direction)
├── vertical_test.go         ← unit tests
├── rule.go                  ← existing (reuse Violation, Config, Option)
├── dependency.go            ← existing (reuse findImportFile, findImportLine)

testdata/
├── vertical-valid/
│   ├── go.mod
│   └── internal/
│       ├── order/
│       │   ├── alias.go
│       │   ├── contract/dto.go
│       │   ├── app/public.go
│       │   ├── app/usecase/create_order.go
│       │   ├── model/order.go
│       │   ├── repo/order_repo.go
│       │   ├── domain/order_logic.go
│       │   ├── handler/http/order_handler.go
│       │   └── infra/persistence/order_store.go
│       ├── user/
│       │   ├── alias.go
│       │   ├── contract/dto.go
│       │   ├── model/user.go
│       │   └── app/public.go
│       └── shared/
│           └── db.go
│
├── vertical-invalid/
│   ├── go.mod
│   └── internal/
│       ├── order/
│       │   ├── alias.go
│       │   ├── contract/dto.go
│       │   ├── app/usecase/create_order.go
│       │   ├── model/order.go
│       │   ├── domain/bad_import.go      ← imports user/ directly (HARD violation)
│       │   ├── handler/http/bad_handler.go ← imports user/model (HARD violation)
│       │   └── infra/persistence/bad_store.go ← imports app/ (SOFT violation)
│       ├── user/
│       │   ├── alias.go
│       │   ├── contract/dto.go
│       │   └── model/user.go
│       └── shared/
│           └── db.go

integration_test.go           ← add vertical slice integration tests
```

## Vertical Slice Domain Detection

A package is inside a vertical slice domain if its path matches:
`{module}/internal/{domain}/...` where `{domain}` is NOT `shared`.

Within a domain, the **sub-layer** is identified by the first path segment after the domain:
- `internal/order/app/usecase/...` → domain=`order`, sublayer=`app`
- `internal/order/model/...` → domain=`order`, sublayer=`model`
- `internal/order/handler/...` → domain=`order`, sublayer=`handler`

## Rules

### HARD: Cross-domain isolation (`CheckVerticalSlice`)

For every import from package A to package B where both are inside `internal/`:
1. If A and B are in the **same domain** → always allowed
2. If A is in `shared/` → **never import a domain** (shared→domain forbidden)
3. If B is in `shared/` → always allowed (anyone can import shared)
4. If A and B are in **different domains**:
   - If A is in `{domain}/app/usecase/...` AND B is in `{otherDomain}/` (alias) or `{otherDomain}/contract/...` → **allowed**
   - Otherwise → **violation** (`vertical.cross-domain-isolation`)

### SOFT: Intra-domain layer direction (`CheckVerticalSliceInternal`)

Within a single domain, enforce layer direction:

| from | allowed to |
|------|-----------|
| handler | app, contract |
| app | domain, policy, model, repo, event, contract |
| domain | model, event |
| policy | model, event |
| infra | model, repo, event, contract |
| model | event |
| repo | model |
| event | (nothing) |
| contract | model (optional) |

Same-sublayer imports are always allowed (e.g., handler/http → handler/grpc).

---

### Task 1: Create testdata for vertical-valid project

**Files:**
- Create: `testdata/vertical-valid/go.mod`
- Create: `testdata/vertical-valid/internal/order/alias.go`
- Create: `testdata/vertical-valid/internal/order/contract/dto.go`
- Create: `testdata/vertical-valid/internal/order/app/public.go`
- Create: `testdata/vertical-valid/internal/order/app/usecase/create_order.go`
- Create: `testdata/vertical-valid/internal/order/model/order.go`
- Create: `testdata/vertical-valid/internal/order/repo/order_repo.go`
- Create: `testdata/vertical-valid/internal/order/domain/order_logic.go`
- Create: `testdata/vertical-valid/internal/order/handler/http/order_handler.go`
- Create: `testdata/vertical-valid/internal/order/infra/persistence/order_store.go`
- Create: `testdata/vertical-valid/internal/user/alias.go`
- Create: `testdata/vertical-valid/internal/user/contract/dto.go`
- Create: `testdata/vertical-valid/internal/user/model/user.go`
- Create: `testdata/vertical-valid/internal/user/app/public.go`
- Create: `testdata/vertical-valid/internal/shared/db.go`

- [ ] **Step 1: Create go.mod**

```
module github.com/kimtaeyun/testproject-vertical

go 1.25.0
```

- [ ] **Step 2: Create user domain (depended-upon domain)**

`user/model/user.go`:
```go
package model

type User struct {
    ID    string
    Name  string
    Email string
}
```

`user/contract/dto.go`:
```go
package contract

type UserResponse struct {
    ID   string
    Name string
}
```

`user/app/public.go`:
```go
package app

import "github.com/kimtaeyun/testproject-vertical/internal/user/model"

type Service struct{}

func (s *Service) GetUser(id string) *model.User {
    return &model.User{ID: id}
}
```

`user/alias.go`:
```go
package user

import "github.com/kimtaeyun/testproject-vertical/internal/user/app"

type Service = app.Service
```

- [ ] **Step 3: Create order domain with valid cross-domain import**

`order/model/order.go`:
```go
package model

type Order struct {
    ID     string
    UserID string
    Amount int
}
```

`order/contract/dto.go`:
```go
package contract

type CreateOrderRequest struct {
    UserID string
    Amount int
}
```

`order/repo/order_repo.go`:
```go
package repo

import "github.com/kimtaeyun/testproject-vertical/internal/order/model"

type Repository interface {
    Save(order *model.Order) error
}
```

`order/domain/order_logic.go`:
```go
package domain

import "github.com/kimtaeyun/testproject-vertical/internal/order/model"

func Validate(o *model.Order) error {
    if o.Amount <= 0 {
        return fmt.Errorf("invalid amount")
    }
    return nil
}
```

`order/app/usecase/create_order.go` — **valid cross-domain: usecase imports user alias + contract**:
```go
package usecase

import (
    "github.com/kimtaeyun/testproject-vertical/internal/order/model"
    "github.com/kimtaeyun/testproject-vertical/internal/user"
    userContract "github.com/kimtaeyun/testproject-vertical/internal/user/contract"
)

type CreateOrder struct {
    userSvc *user.Service
}

func (uc *CreateOrder) Execute(userID string, amount int) (*model.Order, error) {
    _ = userContract.UserResponse{} // valid: importing other domain's contract
    return &model.Order{ID: "1", UserID: userID, Amount: amount}, nil
}
```

`order/app/public.go`:
```go
package app

type Service struct{}
```

`order/handler/http/order_handler.go`:
```go
package http

import "github.com/kimtaeyun/testproject-vertical/internal/order/app"

func Handle(svc *app.Service) {}
```

`order/infra/persistence/order_store.go`:
```go
package persistence

import "github.com/kimtaeyun/testproject-vertical/internal/order/model"

type Store struct{}

func (s *Store) Save(o *model.Order) error { return nil }
```

`order/alias.go`:
```go
package order

import "github.com/kimtaeyun/testproject-vertical/internal/order/app"

type Service = app.Service
```

- [ ] **Step 4: Create shared package**

`shared/db.go`:
```go
package shared

type DB struct{}
```

- [ ] **Step 5: Verify project compiles**

```bash
cd testdata/vertical-valid && go build ./...
```

- [ ] **Step 6: Commit**

```bash
git add testdata/vertical-valid/
git commit -m "test: add vertical-valid testdata project"
```

---

### Task 2: Create testdata for vertical-invalid project

**Files:**
- Create: `testdata/vertical-invalid/go.mod`
- Create: same domain structure as vertical-valid, plus violation files

- [ ] **Step 1: Create go.mod and basic structure**

Copy the same user domain + shared from vertical-valid.

- [ ] **Step 2: Create order domain with violations**

`order/domain/bad_import.go` — **HARD violation: domain directly imports another domain's model**:
```go
package domain

import _ "github.com/kimtaeyun/testproject-vertical-invalid/internal/user/model"
```

`order/handler/http/bad_handler.go` — **HARD violation: handler imports another domain**:
```go
package http

import _ "github.com/kimtaeyun/testproject-vertical-invalid/internal/user/model"

func BadHandle() {}
```

`order/infra/persistence/bad_store.go` — **SOFT violation: infra imports app (wrong direction within domain)**:
```go
package persistence

import _ "github.com/kimtaeyun/testproject-vertical-invalid/internal/order/app"

type BadStore struct{}
```

- [ ] **Step 3: Verify project compiles**

```bash
cd testdata/vertical-invalid && go build ./...
```

- [ ] **Step 4: Commit**

```bash
git add testdata/vertical-invalid/
git commit -m "test: add vertical-invalid testdata project"
```

---

### Task 3: Implement CheckVerticalSlice (HARD rule)

**Files:**
- Create: `rules/vertical.go`
- Create: `rules/vertical_test.go`

- [ ] **Step 1: Write failing test**

`rules/vertical_test.go`:
```go
package rules_test

import (
    "testing"
    "github.com/kimtaeyun/go-arch-guard/analyzer"
    "github.com/kimtaeyun/go-arch-guard/rules"
)

func TestCheckVerticalSlice(t *testing.T) {
    t.Run("valid project has no violations", func(t *testing.T) {
        pkgs, err := analyzer.Load("../testdata/vertical-valid", "internal/...")
        if err != nil {
            t.Fatal(err)
        }
        violations := rules.CheckVerticalSlice(pkgs, "github.com/kimtaeyun/testproject-vertical", "../testdata/vertical-valid")
        if len(violations) > 0 {
            for _, v := range violations {
                t.Log(v.String())
            }
            t.Errorf("expected no violations, got %d", len(violations))
        }
    })

    t.Run("detects cross-domain violation from domain layer", func(t *testing.T) {
        pkgs, err := analyzer.Load("../testdata/vertical-invalid", "internal/...")
        if err != nil {
            t.Fatal(err)
        }
        violations := rules.CheckVerticalSlice(pkgs, "github.com/kimtaeyun/testproject-vertical-invalid", "../testdata/vertical-invalid")
        found := findViolation(violations, "vertical.cross-domain-isolation")
        if found == nil {
            t.Error("expected cross-domain-isolation violation")
        }
    })

    t.Run("allows usecase to import other domain alias and contract", func(t *testing.T) {
        pkgs, err := analyzer.Load("../testdata/vertical-valid", "internal/...")
        if err != nil {
            t.Fatal(err)
        }
        violations := rules.CheckVerticalSlice(pkgs, "github.com/kimtaeyun/testproject-vertical", "../testdata/vertical-valid")
        for _, v := range violations {
            if v.Rule == "vertical.cross-domain-isolation" {
                t.Errorf("usecase cross-domain import should be allowed, got: %s", v.String())
            }
        }
    })
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./rules/ -run TestCheckVerticalSlice -v
```
Expected: FAIL (CheckVerticalSlice not defined)

- [ ] **Step 3: Implement CheckVerticalSlice**

`rules/vertical.go`:

Key logic:
1. `identifyDomain(pkgPath, internalPrefix)` — returns domain name (first segment after `internal/`), or `"shared"`, or `""`
2. `identifySublayer(pkgPath, internalPrefix, domain)` — returns sublayer (`app`, `model`, `domain`, `handler`, `infra`, etc.)
3. `isUsecasePath(pkgPath, internalPrefix, domain)` — checks if path is `{domain}/app/usecase/...`
4. `isAliasOrContract(importPath, internalPrefix, domain)` — checks if import is `{domain}` (root = alias.go) or `{domain}/contract/...`

For each package → for each import:
- Skip if import is not in `internal/`
- Skip if same domain
- Skip if importing `shared/`
- If source is `shared/` and target is a domain → violation (`vertical.shared-imports-domain`)
- If source is in `app/usecase/` AND target is other domain's alias or contract → allowed
- Otherwise → violation (`vertical.cross-domain-isolation`)

- [ ] **Step 4: Run test to verify it passes**

```bash
go test ./rules/ -run TestCheckVerticalSlice -v
```
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add rules/vertical.go rules/vertical_test.go
git commit -m "feat: add CheckVerticalSlice for cross-domain isolation"
```

---

### Task 4: Implement CheckVerticalSliceInternal (SOFT rule)

**Files:**
- Create: `rules/vertical_internal.go`
- Modify: `rules/vertical_test.go`

- [ ] **Step 1: Write failing test**

Add to `rules/vertical_test.go`:
```go
func TestCheckVerticalSliceInternal(t *testing.T) {
    t.Run("valid project has no violations", func(t *testing.T) {
        pkgs, err := analyzer.Load("../testdata/vertical-valid", "internal/...")
        if err != nil {
            t.Fatal(err)
        }
        violations := rules.CheckVerticalSliceInternal(pkgs, "github.com/kimtaeyun/testproject-vertical", "../testdata/vertical-valid")
        if len(violations) > 0 {
            for _, v := range violations {
                t.Log(v.String())
            }
            t.Errorf("expected no violations, got %d", len(violations))
        }
    })

    t.Run("detects infra importing app", func(t *testing.T) {
        pkgs, err := analyzer.Load("../testdata/vertical-invalid", "internal/...")
        if err != nil {
            t.Fatal(err)
        }
        violations := rules.CheckVerticalSliceInternal(pkgs, "github.com/kimtaeyun/testproject-vertical-invalid", "../testdata/vertical-invalid")
        found := findViolation(violations, "vertical.internal-layer-direction")
        if found == nil {
            t.Error("expected internal-layer-direction violation")
        }
    })
}
```

- [ ] **Step 2: Run test to verify it fails**

- [ ] **Step 3: Implement CheckVerticalSliceInternal**

`rules/vertical_internal.go`:

Define sublayer allowed imports table:
```go
var verticalAllowed = map[string][]string{
    "handler":  {"app", "contract"},
    "app":      {"domain", "policy", "model", "repo", "event", "contract"},
    "domain":   {"model", "event"},
    "policy":   {"model", "event"},
    "infra":    {"model", "repo", "event", "contract"},
    "model":    {"event"},
    "repo":     {"model"},
    "event":    {},
    "contract": {"model"},
}
```

For each package → for each import within the same domain:
- Identify source sublayer and destination sublayer
- If same sublayer → allowed
- If destination is in the allowed list → allowed
- Otherwise → violation (`vertical.internal-layer-direction`)

- [ ] **Step 4: Run test to verify it passes**

```bash
go test ./rules/ -run TestCheckVerticalSliceInternal -v
```

- [ ] **Step 5: Commit**

```bash
git add rules/vertical_internal.go rules/vertical_test.go
git commit -m "feat: add CheckVerticalSliceInternal for intra-domain layer direction"
```

---

### Task 5: Integration tests + README update

**Files:**
- Modify: `integration_test.go`
- Modify: `README.md`

- [ ] **Step 1: Add vertical slice integration tests**

```go
func TestIntegration_VerticalValid(t *testing.T) {
    pkgs, err := analyzer.Load("testdata/vertical-valid", "internal/...")
    if err != nil {
        t.Fatal(err)
    }
    t.Run("cross-domain", func(t *testing.T) {
        report.AssertNoViolations(t, rules.CheckVerticalSlice(pkgs, "github.com/kimtaeyun/testproject-vertical", "testdata/vertical-valid"))
    })
    t.Run("internal-direction", func(t *testing.T) {
        report.AssertNoViolations(t, rules.CheckVerticalSliceInternal(pkgs, "github.com/kimtaeyun/testproject-vertical", "testdata/vertical-valid"))
    })
}

func TestIntegration_VerticalInvalid(t *testing.T) {
    pkgs, err := analyzer.Load("testdata/vertical-invalid", "internal/...")
    if err != nil {
        t.Fatal(err)
    }
    t.Run("cross-domain violations found", func(t *testing.T) {
        violations := rules.CheckVerticalSlice(pkgs, "github.com/kimtaeyun/testproject-vertical-invalid", "testdata/vertical-invalid")
        if len(violations) == 0 {
            t.Error("expected cross-domain violations")
        }
    })
    t.Run("internal-direction violations found", func(t *testing.T) {
        violations := rules.CheckVerticalSliceInternal(pkgs, "github.com/kimtaeyun/testproject-vertical-invalid", "testdata/vertical-invalid")
        if len(violations) == 0 {
            t.Error("expected internal-direction violations")
        }
    })
}
```

- [ ] **Step 2: Run all tests**

```bash
go test ./... -count=1 -v
```

- [ ] **Step 3: Update README.md**

Add a "Vertical Slice" section after the existing rules docs:

```markdown
## Vertical Slice Rules

For projects using domain-per-directory (vertical slice) architecture.

### Cross-Domain Isolation (`rules.CheckVerticalSlice`) — HARD

Only `app/usecase/` can import other domains, and only via `alias.go` or `contract/`.

### Intra-Domain Direction (`rules.CheckVerticalSliceInternal`) — SOFT

Enforces layer direction within a domain: handler → app → domain/policy → model.
```

- [ ] **Step 4: Commit**

```bash
git add integration_test.go README.md
git commit -m "feat: add vertical slice integration tests and docs"
```

- [ ] **Step 5: Run full test suite and push**

```bash
go test ./... -count=1
git push
```
