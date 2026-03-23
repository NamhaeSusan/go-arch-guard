# go-arch-guard

Architecture rule enforcement for Go projects via `go test`.

Define layer dependencies, naming conventions, and structural rules — then enforce them as regular test failures in CI.

## Install

```bash
go get github.com/kimtaeyun/go-arch-guard
```

## Usage

Write a test in your project:

```go
func TestArchitecture(t *testing.T) {
    pkgs, err := analyzer.Load(".", "internal/...")
    if err != nil {
        t.Fatal(err)
    }

    t.Run("dependency", func(t *testing.T) {
        report.AssertNoViolations(t, rules.CheckDependency(pkgs, "github.com/yourmodule", "."))
    })
    t.Run("naming", func(t *testing.T) {
        report.AssertNoViolations(t, rules.CheckNaming(pkgs))
    })
    t.Run("structure", func(t *testing.T) {
        report.AssertNoViolations(t, rules.CheckStructure("."))
    })
}
```

Run with `go test`.

## Rules

### Dependency (`rules.CheckDependency`)

Enforces layer dependency direction for `internal/` packages:

| Layer | Allowed imports |
|-------|----------------|
| handler | app, domain |
| app | domain |
| infra | domain |
| domain | (nothing) |

Also detects cross-domain imports within the domain layer.

### Naming (`rules.CheckNaming`)

- **No stutter** — `user.UserService` → rename to `user.Service`
- **No `Impl` suffix** — `ServiceImpl` is banned
- **Snake case files** — `userService.go` → `user_service.go`
- **Repo file interface** — `repo/user.go` must contain `type User interface`
- **No layer suffix** — `svc/install_svc.go` → rename to `svc/install.go`

### Vertical Slice (`rules.CheckVerticalSlice`)

Enforces cross-domain isolation for vertical slice architecture under `internal/`:

- Same domain imports → always allowed
- Import `shared/` → always allowed
- `shared/` importing a domain → violation (`vertical.shared-imports-domain`)
- Cross-domain import from `app/usecase/` to other domain's root (alias) or `port/` → allowed
- Any other cross-domain import → violation (`vertical.cross-domain-isolation`)

```go
rules.CheckVerticalSlice(pkgs, "github.com/yourmodule", ".")
```

### Vertical Slice Internal (`rules.CheckVerticalSliceInternal`)

Enforces intra-domain layer direction within a vertical slice. Cross-domain and `shared/` imports are skipped (handled by `CheckVerticalSlice`).

Allowed imports per sublayer:

| from | allowed to import |
|------|-------------------|
| handler | app, port |
| app | domain, policy, model, repo, event, port |
| domain | model, event |
| policy | model, event |
| infra | model, repo, event, port |
| model | event |
| repo | model |
| event | (nothing) |
| port | model |

Same-sublayer imports are always allowed. Violation rule: `vertical.internal-layer-direction`.

```go
rules.CheckVerticalSliceInternal(pkgs, "github.com/yourmodule", ".")
```

### Domain Isolation (`rules.CheckDomainIsolation`)

Enforces cross-domain isolation for domain-centric architecture under `internal/domain/`:

- Same domain imports → always allowed
- Import `pkg/` → always allowed
- `pkg/` importing a domain → violation (`isolation.pkg-imports-domain`)
- `saga/` (non-handler) importing domain alias → allowed
- `saga/` importing domain sub-package → violation (`isolation.saga-deep-import`)
- Cross-domain import → violation (`isolation.cross-domain`)

```go
rules.CheckDomainIsolation(pkgs, "github.com/yourmodule", ".")
```

### Layer Direction (`rules.CheckLayerDirection`)

Enforces intra-domain layer direction within a domain-centric architecture. Cross-domain, `pkg/`, and `saga/` imports are skipped (handled by `CheckDomainIsolation`).

Allowed imports per sublayer:

| from | allowed to import |
|------|-------------------|
| "" (alias) | app |
| handler | app |
| app | core/model, core/repo, core/svc |
| core/svc | core/model |
| core/repo | core/model |
| infra | core/repo, core/model |
| event | core/model |
| core/model | (nothing) |

Same-sublayer imports are always allowed. Violation rule: `layer.direction`.

```go
rules.CheckLayerDirection(pkgs, "github.com/yourmodule", ".")
```

### Structure (`rules.CheckStructure`)

- **Banned packages** — `util`, `common`, `misc`, `helper`, `shared`
- **Domain model required** — each `internal/domain/<name>/` must have `model.go`
- **DTO placement** — `dto.go` files must not be in `domain/` or `infra/`

## Options

Use `WithSeverity` to degrade violations to warnings (test passes but logs them):

```go
rules.CheckDependency(pkgs, module, root, rules.WithSeverity(rules.Warning))
```

Use `WithExclude` to skip specific paths:

```go
rules.CheckNaming(pkgs, rules.WithExclude("internal/legacy/..."))
```

## License

MIT
