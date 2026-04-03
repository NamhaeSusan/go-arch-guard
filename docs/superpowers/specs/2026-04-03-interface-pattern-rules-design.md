# Interface Pattern Rules

**Date:** 2026-04-03
**Status:** Approved

## Summary

Add 3 new rules enforcing Go interface best practices:
one exported interface per package, private implementation, and `New()` returns interface.

## Rules

### 1. `interface.exported-impl`

If a package contains an exported interface, any struct implementing that interface
must be unexported.

```go
// ✅
type Repository interface { FindByID(id int64) (*Order, error) }
type repository struct { db *sql.DB }  // private

// ❌
type Repository interface { FindByID(id int64) (*Order, error) }
type RepositoryImpl struct { db *sql.DB }  // exported impl
```

**Detection:** For each exported interface in a package, scan all struct types.
If an exported struct implements the interface (has all methods), emit violation.

### 2. `interface.constructor-name`

Constructor functions must be named exactly `New`. `NewXxx` variants are not allowed.

```go
// ✅
func New(db *sql.DB) Repository { return &repository{db: db} }

// ❌
func NewRepository(db *sql.DB) Repository { ... }
func NewOrderRepository(db *sql.DB) Repository { ... }
```

**Detection:** Any exported function whose name starts with "New" and has length > 3
(i.e., `NewX...`) is a violation. Only exact `New` is allowed.

### 3. `interface.constructor-returns-interface`

The `New()` function must return an exported interface, not a concrete type.

```go
// ✅
func New(db *sql.DB) Repository { return &repository{db: db} }

// ❌
func New(db *sql.DB) *repository { return &repository{db: db} }
```

**Detection:** Find `New` function in package. If its first return type is
a pointer or struct (not an interface), emit violation.

## Excluded Layers

These rules skip packages in layers where interfaces don't make sense.
Controlled by `InterfacePatternExclude map[string]bool` field on Model.

Default excludes for all presets: `model`, `event`, `pkg`.

For flat-layout presets, additional excludes per preset:
- ConsumerWorker: `worker` (entry point, no interface needed)
- Batch: `job` (entry point)
- EventPipeline: `command`, `aggregate` (have TypePattern instead)

For domain-centric presets, additional excludes:
- DDD: `handler`, `app`, `core/model`, `event`
- Others: `handler`, `model`/`entity`/`domain`/`core` (the model layer of each preset)

### Single Interface Per Package

If a package contains more than one exported interface, emit a violation:

Rule: `interface.single-per-package`
Message: "package has N exported interfaces, expected at most 1; split into separate packages"
Severity: Warning (not Error — this is a guideline, not a hard gate)

## Model Changes

Add field to `Model`:

```go
InterfacePatternExclude map[string]bool  // layers to skip for interface pattern checks
```

Add to each preset factory. Example for DDD:

```go
InterfacePatternExclude: map[string]bool{
    "handler": true, "app": true, "core/model": true, "event": true,
},
```

## New Rule Function

```go
func CheckInterfacePattern(pkgs []*packages.Package, opts ...Option) []Violation
```

Added to `RunAll` between `CheckTypePatterns` and `AnalyzeBlastRadius`.

## Implementation File

`rules/interface_pattern.go` — contains:
- `CheckInterfacePattern` — main entry
- `checkExportedImpl` — rule 1
- `checkConstructorName` — rule 2
- `checkConstructorReturnsInterface` — rule 3
- `checkSingleInterfacePerPackage` — rule 4 (warning)

## Scope

| File | Action |
|---|---|
| `rules/model.go` | Add `InterfacePatternExclude` field, update all preset factories |
| `rules/interface_pattern.go` | Create — 4 rule checks |
| `rules/interface_pattern_test.go` | Create — unit tests |
| `rules/run_all.go` | Wire into RunAll |
| `integration_test.go` | Add integration tests |
| `README.md`, `README.ko.md` | Add rule docs |
| `plugins/.../SKILL.md` files | Update rule tables |
