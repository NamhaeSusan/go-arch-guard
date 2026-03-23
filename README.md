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

Run with `go test`.

## Rules

### Naming (`rules.CheckNaming`)

- **No stutter** — `user.UserService` → rename to `user.Service`
- **No `Impl` suffix** — `ServiceImpl` is banned
- **Snake case files** — `userService.go` → `user_service.go`
- **Repo file interface** — `repo/user.go` must contain `type User interface`
- **No layer suffix** — `svc/install_svc.go` → rename to `svc/install.go`

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
rules.CheckDomainIsolation(pkgs, module, root, rules.WithSeverity(rules.Warning))
```

Use `WithExclude` to skip specific paths:

```go
rules.CheckNaming(pkgs, rules.WithExclude("internal/legacy/..."))
```

## License

MIT
