# go-arch-guard Design Spec

## Overview

A Go AST-based static analysis library that enforces DDD layered architecture rules. Designed as an independent Go module that target projects import and run via `go test`.

Inspired by [Channel Corp's approach](https://channel.io/ko/team/blog/articles/ai-native-ddd-refactoring-98c23cdb): "Don't write rules as docs. Enforce them as code."

- **Module path**: `github.com/kimtaeyun/go-arch-guard`
- **Go version**: 1.22+

## Target Project Architecture

DDD layered structure for mid-scale API servers (10K-50K LOC):

```
your-api-server/
├── internal/
│   ├── domain/              # Domain layer (pure business logic)
│   │   ├── user/
│   │   │   ├── model.go     # Entity, Value Object
│   │   │   ├── repo.go      # Repository interface
│   │   │   └── service.go   # Domain service
│   │   └── order/
│   │       ├── model.go
│   │       ├── repo.go
│   │       └── service.go
│   ├── app/                 # Application layer (use cases)
│   │   ├── user_service.go
│   │   └── dto.go           # Inter-service DTOs (Input/Output)
│   ├── infra/               # Infrastructure layer (implementations)
│   │   ├── postgres/
│   │   ├── redis/
│   │   └── external/
│   └── handler/             # Presentation layer
│       ├── http/
│       │   ├── user_handler.go
│       │   └── dto.go       # Request/Response DTOs
│       └── grpc/
├── pkg/
└── test/
    └── architecture/        # go-arch-guard rule tests
```

### Layer Dependency Direction

```
handler → app → domain ← infra
```

- `domain` depends on nothing within `internal/` (pure business logic)
- `infra` implements `domain` interfaces (depends on `domain` only)
- `handler` depends on `app` and `domain` (needs domain models for DTO conversion)
- `app` depends on `domain` only
- `pkg/` is a utility layer — any layer may import `pkg/`
- External libraries (outside the project module) are not checked

**Scope**: dependency rules only check imports between `internal/` sub-packages. Imports of `pkg/`, stdlib, and third-party modules are ignored.

### DTO Placement

DTO is identified by **filename** (`dto.go`, `*_dto.go`):

- `handler/*/dto.go` — Request/Response DTOs
- `app/dto.go` — Service-to-service DTOs (Input/Output)
- `domain/` — No DTO files allowed
- `infra/` — No DTO files allowed (internal conversion structs in non-DTO files are fine)

## go-arch-guard Module Structure

```
go-arch-guard/
├── analyzer/
│   ├── loader.go          # Load packages via go/packages
│   └── visitor.go         # Import extraction, type declaration walking
├── rules/
│   ├── rule.go            # Rule and Violation types
│   ├── dependency.go      # Layer dependency direction enforcement
│   ├── naming.go          # Naming convention checks
│   └── structure.go       # Directory/package structure checks
├── report/
│   └── report.go          # Violation formatting + test assertion helper
├── go.mod
└── rules_test.go          # Self-tests for go-arch-guard
```

### Core Types

```go
// analyzer/loader.go
// Load returns parsed packages for the given module-relative patterns.
// dir is the project root (where go.mod lives).
// patterns are module-relative paths like "internal/..." or "internal/domain/...".
func Load(dir string, patterns ...string) ([]*packages.Package, error)

// rules/rule.go
type Severity int
const (
    Error   Severity = iota // test fails (default)
    Warning                 // prints violations but test passes
)

type Option func(*config)
func WithSeverity(s Severity) Option      // set severity level
func WithExclude(patterns ...string) Option // exclude paths (glob), e.g. "internal/legacy/..."

type Violation struct {
    File     string   // relative file path from project root
    Line     int      // line number (0 if not applicable, e.g. structure rules)
    Rule     string   // rule identifier, e.g. "dependency.layer-direction"
    Message  string   // human-readable description
    Fix      string   // actionable fix suggestion
    Severity Severity // Error or Warning
}

func (v Violation) String() string
// Output format:
// [ERROR] violation: <Message> (file: <File>:<Line>, rule: <Rule>, fix: <Fix>)
// [WARNING] violation: <Message> (file: <File>:<Line>, rule: <Rule>, fix: <Fix>)

// Each rule package exposes a Check function:
// rules/dependency.go
func CheckDependency(pkgs []*packages.Package, projectModule string, projectRoot string, opts ...Option) []Violation
// rules/naming.go
func CheckNaming(pkgs []*packages.Package, opts ...Option) []Violation
// rules/structure.go
func CheckStructure(projectRoot string, opts ...Option) []Violation  // filesystem-based, no AST needed

// report/report.go
// AssertNoViolations prints all violations. Fails test only if any ERROR-level violations exist.
func AssertNoViolations(t testing.TB, violations []Violation)
```

### Core Flow

1. `analyzer.Load()` loads target project packages with `go/packages` (mode: `NeedName | NeedImports | NeedFiles | NeedSyntax | NeedTypes`)
2. Each `Check*` function inspects packages/filesystem and returns `[]Violation`
3. `report.AssertNoViolations()` fails the test with formatted violation messages

### Usage in Target Project

```go
// your-api-server/test/architecture/arch_test.go
package architecture_test

import (
    "testing"

    "github.com/kimtaeyun/go-arch-guard/analyzer"
    "github.com/kimtaeyun/go-arch-guard/report"
    "github.com/kimtaeyun/go-arch-guard/rules"
)

const projectRoot = "../.."
const projectModule = "github.com/kimtaeyun/your-api-server"

func TestArchitecture(t *testing.T) {
    pkgs, err := analyzer.Load(projectRoot, "internal/...")
    if err != nil {
        t.Fatal(err)
    }

    // Gradual enforcement: start with Warning, switch to Error after refactoring
    t.Run("dependency", func(t *testing.T) {
        report.AssertNoViolations(t, rules.CheckDependency(pkgs, projectModule, projectRoot,
            rules.WithSeverity(rules.Warning),             // won't fail test yet
            rules.WithExclude("internal/legacy/..."),      // skip legacy code
        ))
    })
    t.Run("naming", func(t *testing.T) {
        report.AssertNoViolations(t, rules.CheckNaming(pkgs))  // default: Error
    })
    t.Run("structure", func(t *testing.T) {
        report.AssertNoViolations(t, rules.CheckStructure(projectRoot))
    })
}
```

### Error Handling

- `analyzer.Load()` returns an error on package load failure — caller should `t.Fatal(err)`
- Individual rule checks are best-effort: if a file can't be parsed, skip it and continue
- All collectable violations are gathered before reporting

## Rules

### Dependency Rules (dependency.go)

Checks import statements in Go source files. Only inspects imports within the project module's `internal/` subtree.

| Rule ID | Description | Error Example |
|---------|-------------|---------------|
| `dependency.layer-direction` | Only `handler→{app,domain}`, `app→domain`, `infra→domain` allowed | `violation: "handler" imports "infra" directly (file: handler/http/user.go:5, rule: dependency.layer-direction, fix: handler must only depend on app or domain)` |
| `dependency.domain-purity` | domain must not import any other internal layer | `violation: "domain/user" imports "app" (file: domain/user/service.go:3, rule: dependency.domain-purity, fix: domain must not depend on any other layer)` |
| `dependency.domain-isolation` | No cross-domain imports (domain/A must not import domain/B) | `violation: domain "user" imports domain "order" directly (rule: dependency.domain-isolation, fix: use app layer to coordinate between domains)` |

### Naming Rules (naming.go)

Checks type declarations and filenames via AST.

| Rule ID | Description | Error Example |
|---------|-------------|---------------|
| `naming.no-stutter` | Exported type name must not start with package name (case-insensitive prefix match). `user.UserService` → stutter, `user.PowerUser` → OK | `violation: type "UserService" stutters with package "user" (file: domain/user/service.go:10, rule: naming.no-stutter, fix: rename to "Service")` |
| `naming.no-impl-suffix` | Exported type name must not end with `Impl` (all layers) | `violation: type "ServiceImpl" uses banned suffix "Impl" (file: infra/postgres/user.go:8, rule: naming.no-impl-suffix, fix: rename without Impl suffix)` |
| `naming.snake-case-file` | Go source filenames must be snake_case (lowercase + underscores + `_test` suffix allowed) | `violation: filename "userService.go" must be snake_case (rule: naming.snake-case-file, fix: rename to "user_service.go")` |

### Structure Rules (structure.go)

Filesystem-based checks — no AST needed.

| Rule ID | Description | Error Example |
|---------|-------------|---------------|
| `structure.banned-package` | Package names `util`, `common`, `misc`, `helper`, `shared` are banned under `internal/` | `violation: package "util" is banned (file: internal/util/, rule: structure.banned-package, fix: move to specific domain or pkg/)` |
| `structure.domain-model-required` | Each domain sub-package must contain `model.go` | `violation: domain "order" missing required file "model.go" (rule: structure.domain-model-required)` |
| `structure.dto-placement` | Files named `dto.go` or `*_dto.go` are only allowed in `handler/` and `app/` | `violation: "dto.go" found in domain/user/ (rule: structure.dto-placement, fix: DTOs belong in handler/ or app/)` |

## Non-Goals (v1)

- YAML/TOML config for rules (hardcoded for now)
- golangci-lint plugin integration
- Custom rule DSL
- Auto-fix capability
- Cross-module analysis
