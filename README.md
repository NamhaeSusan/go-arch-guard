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
