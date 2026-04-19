# Model Concepts

This guide explains what each field in a `rules.Model` actually means and when you would change it. It targets teams who want to build a custom model via `rules.NewModel(...)` — not maintain presets, and not read the full godoc reference.

For the option list, see [README § Custom Model Options](../README.md#custom-model-options).
For layout diagrams of each preset, see [Preset Details](presets.md).

---

## What is a Model?

A `Model` is a struct that describes your project's intended architecture shape. Rules read the model and compare it against the actual code on disk and in the import graph. If they diverge, a rule emits a violation.

Presets (`DDD()`, `CleanArch()`, `ConsumerWorker()`, etc.) are pre-built Models with defaults tuned for common shapes. `NewModel(...)` lets teams describe shapes that don't match any preset, or that match one closely enough to start there and override a few fields.

The library is neutral infrastructure. It enforces whatever shape you commit to. If your team doesn't have a convention on something, leave that field at its default — or use a preset and only override what you know you want different.

---

## The fields, explained

### `Sublayers`

**What:** the complete list of layer names that may exist inside a domain (for domain layouts) or directly under `internal/` (for flat layouts).

**Why:** `CheckLayerDirection` uses this list to detect unknown sublayers. A package that lands in a directory not named in `Sublayers` triggers `layer.unknown-sublayer` — the rule cannot reason about a layer it doesn't know about.

**Example (DDD):**

```go
rules.WithSublayers([]string{
    "handler", "app", "core", "core/model",
    "core/repo", "core/svc", "event", "infra",
})
```

Slash notation (`core/model`) represents a subdirectory inside `core/`. The rule matches on the suffix of the package path, so `internal/domain/order/core/model` matches `core/model`.

**Common mistake:** adding a new directory (e.g. `cache/`) without adding it here. Every package in that directory will fail with `layer.unknown-sublayer`.

---

### `Direction`

**What:** a map from layer name to the list of layers that layer is allowed to import.

**Why:** prevents architectural inversion. A violation means a package in layer A is importing a package in layer B, but `Direction["A"]` does not contain `"B"`.

**Example (DDD):**

```go
rules.WithDirection(map[string][]string{
    "handler":    {"app"},
    "app":        {"core/model", "core/repo", "core/svc", "event"},
    "core":       {"core/model"},
    "core/model": {},
    "core/repo":  {"core/model"},
    "core/svc":   {"core/model"},
    "event":      {"core/model"},
    "infra":      {"core/repo", "core/model", "event"},
})
```

**Gotcha:** if a layer is in `Sublayers` but has no entry in `Direction`, the rule treats it as "may import nothing." An empty slice `{}` is explicit; a missing key behaves the same way. Add an explicit empty entry for clarity.

**Gotcha:** `Direction` only governs cross-layer imports *within* the same domain. Cross-domain imports are handled by `CheckDomainIsolation`, not `CheckLayerDirection`.

---

### `DomainDir`

**What:** the name of the top-level directory under `internal/` that groups per-domain packages.

**Why:** distinguishes domain-layout presets from flat-layout presets.

- Domain layout: `internal/{DomainDir}/{domain-name}/{sublayer}/`
  Example: `internal/domain/order/app/`
- Flat layout: `internal/{sublayer}/`

**Set to `""`** for flat layout (ConsumerWorker, Batch, EventPipeline). The structure and isolation rules skip domain-specific checks entirely when `DomainDir` is empty.

**Example:**

```go
rules.WithDomainDir("module")  // internal/module/{name}/{sublayer}/
rules.WithDomainDir("")        // flat layout, internal/{sublayer}/
```

---

### `OrchestrationDir`

**What:** the name of the top-level directory under `internal/` for cross-domain coordination code.

**Why:** `CheckDomainIsolation` allows orchestration packages to import domain roots. Code outside `DomainDir` and `OrchestrationDir` that imports a domain root triggers `isolation.orchestration-deep-import` or similar.

**Set to `""`** for flat layouts where orchestration is not a concept.

**Example:**

```go
rules.WithOrchestrationDir("workflow")  // internal/workflow/
```

---

### `SharedDir`

**What:** the name of the top-level directory under `internal/` for neutral utilities (logging, tracing, DB connection pools, etc.) that many layers import.

**Why:** the isolation rule does not flag imports of packages under `SharedDir` as cross-domain leaks. `PkgRestricted` gates which layers may import shared packages.

**Example:**

```go
rules.WithSharedDir("lib")  // internal/lib/
rules.WithSharedDir("pkg")  // internal/pkg/ (DDD default)
```

---

### `InternalTopLevel`

**What:** the set of directory names allowed directly under `internal/`.

**Why:** `CheckStructure` emits `structure.internal-top-level` for any directory under `internal/` that is not in this set. This catches typos and accidental package placement (e.g. someone creates `internal/utils/` by mistake).

**You should not set this manually.** `NewModel(...)` derives it automatically from `DomainDir`, `OrchestrationDir`, and `SharedDir` after applying all options:

```
InternalTopLevel = {DomainDir, OrchestrationDir, SharedDir} (non-empty values only)
```

For flat layouts where `DomainDir == ""` and `OrchestrationDir == ""`, only `SharedDir` (and the sublayer directories themselves, which you add to `InternalTopLevel` manually) are allowed. If you use a flat layout with `NewModel`, set `InternalTopLevel` explicitly via a raw field assignment or list the sublayer dirs in a custom option, because `NewModel` only includes `SharedDir` automatically.

---

### `PkgRestricted`

**What:** the set of sublayer names that must not import the shared `pkg/` directory.

**Why:** prevents core/domain code from taking on infrastructure dependencies hidden in a shared package. In DDD, `core/model` and friends should be dependency-free; marking them restricted means the rule flags `import "myapp/internal/pkg/..."` inside those layers.

**Example (DDD):**

```go
rules.WithPkgRestricted(map[string]bool{
    "core":       true,
    "core/model": true,
    "core/repo":  true,
    "core/svc":   true,
    "event":      true,
})
```

Set to `nil` or an empty map to allow all layers to import `pkg/`.

---

### Port and contract detection

The library uses two internal concepts — **port sublayers** and **contract sublayers** — that are derived from `Sublayers` by convention, not by a separate Model field.

**Port sublayer:** a sublayer whose basename is `repo` or `gateway`. These are pure interface-declaration layers (no implementations, just contracts for external dependencies).

**Contract sublayer:** port sublayers plus any sublayer whose basename is `svc` (service interface). Used by re-export checks that reason about "interfaces consumed by callers."

**Implication for custom models:** if your port layer uses a different name (e.g. `store`, `boundary`, `port`), the built-in port detection will not recognize it. The rules that rely on port detection (`interface.single-per-package`, `interface.max-methods`, etc.) use a basename suffix fallback and will fall back to `core/repo` if nothing matches. You can work around this by naming your port layer with basename `repo` or `gateway`.

This is an intentional simplification: the library matches on a well-known basename convention rather than requiring a new Model field for every semantic concept.

---

### `RequireAlias` and `AliasFileName`

**What:** `RequireAlias` gates whether each domain root must contain an alias file (default `alias.go`). `AliasFileName` sets the expected filename.

**Why:** the alias file pattern restricts the surface a domain exposes. Only types re-exported through `alias.go` are available to `cmd/` and `orchestration/`. Without the check, callers can import arbitrary sub-packages directly.

**When to disable:** flat-layout presets (ConsumerWorker, Batch, EventPipeline) have no domain directories, so `RequireAlias: false` is correct. Domain presets default to `false` too, except DDD.

```go
rules.WithRequireAlias(false)
rules.WithAliasFileName("exports.go")  // if your team uses a different name
```

---

### `RequireModel` and `ModelPath`

**What:** `RequireModel` gates whether each domain must contain a model directory. `ModelPath` sets the expected path (relative to the domain root).

**Why:** in DDD, every domain must have a `core/model` package (the canonical domain model). The check prevents creating a domain stub without a real model.

**When to disable:** most presets set this to `false`. Enable only if your team's convention requires a model directory in every domain.

```go
rules.WithRequireModel(true)
rules.WithModelPath("core/model")  // default for DDD
```

---

### `DTOAllowedLayers`

**What:** the sublayers where Data Transfer Objects (structs that carry data across boundaries, e.g. HTTP request/response bodies) are allowed.

**Why:** DTOs often carry JSON tags and validation annotations that do not belong in domain model types. Restricting DTO placement prevents leaking infrastructure concerns into the domain.

**Example:**

```go
rules.WithDTOAllowedLayers([]string{"handler", "app"})
```

---

### `BannedPkgNames` and `LegacyPkgNames`

**What:** `BannedPkgNames` is a list of package names (not paths) that the naming rule rejects under `internal/`. `LegacyPkgNames` is a list that triggers a warning instead of an error, signaling migration candidates.

**Defaults:**

```go
BannedPkgNames:  []string{"util", "common", "misc", "helper", "shared", "services"}
LegacyPkgNames:  []string{"router", "bootstrap"}
```

These names are common in codebases that haven't committed to a layer structure. The rule is a guardrail against vague naming during vibe coding, not a semantic check.

**To extend the list:**

```go
rules.WithBannedPkgNames(append(rules.DefaultBannedPkgNames, "managers", "handlers"))
```

---

### `LayerDirNames`

**What:** directory names that the naming rule treats as "layer-like" for purposes like checking that a file in a `handler/` directory doesn't define types that look like repository implementations.

**Why:** the naming check needs to know which directories represent layers so it can apply layer-appropriate naming conventions without false-positiving on project-specific directories.

You rarely need to change this. It becomes relevant only when you add a new sublayer with a name the default set doesn't include.

---

### `InterfacePatternExclude`

**What:** sublayers to skip when checking the interface-pattern rule (`interface.exported-impl`, `interface.single-per-package`).

**Why:** not every layer follows the "unexported struct, exported interface" pattern. `handler` and `model`/`entity` layers typically have exported structs by design (HTTP handlers, domain entities). Excluding them avoids false positives.

**Example:**

```go
rules.WithInterfacePatternExclude(map[string]bool{
    "handler":    true,
    "app":        true,
    "core/model": true,
})
```

---

### `TypePatterns`

**What:** AST-based conventions that require files matching a naming prefix to define a matching exported type with a specific method.

**Why:** flat-layout presets like ConsumerWorker and Batch rely on file-level conventions (`worker_order.go` must define `OrderWorker` with `Process`). `TypePatterns` encodes those conventions so the rule can check them mechanically.

**Example:**

```go
rules.TypePattern{
    Dir:           "worker",
    FilePrefix:    "worker",
    TypeSuffix:    "Worker",
    RequireMethod: "Process",
}
```

This is an advanced option. Most teams don't need it unless they adopt a flat-layout preset and add new entry-point types.

---

## Building a custom Model — worked example

Suppose you have a `messaging` service with this layout:

```text
internal/
├── publisher/    # sends messages to a broker
├── subscriber/   # receives messages from a broker
├── broker/       # broker client abstraction (pure interfaces)
├── dto/          # message payload types
└── pkg/          # shared infra (logging, config)
```

Direction:
- `publisher` and `subscriber` import `broker` and `dto`
- `broker` imports nothing (pure interface layer)
- `dto` is allowed everywhere

```go
m := rules.NewModel(
    // Flat layout: no domain grouping
    rules.WithDomainDir(""),
    rules.WithOrchestrationDir(""),
    rules.WithSharedDir("pkg"),

    // All layer names, including pkg-adjacent ones
    rules.WithSublayers([]string{"publisher", "subscriber", "broker", "dto"}),

    rules.WithDirection(map[string][]string{
        "publisher":  {"broker", "dto"},
        "subscriber": {"broker", "dto"},
        "broker":     {},
        "dto":        {},
    }),

    // broker and dto should not pull from pkg/
    rules.WithPkgRestricted(map[string]bool{
        "broker": true,
        "dto":    true,
    }),

    // No alias or model requirements — flat layout
    rules.WithRequireAlias(false),
    rules.WithRequireModel(false),

    // DTO types are plain structs, exclude from interface pattern check
    rules.WithInterfacePatternExclude(map[string]bool{
        "dto":       true,
        "publisher": true,
        "subscriber": true,
    }),

    // Banned names apply globally
    rules.WithBannedPkgNames([]string{"util", "common", "misc", "helper", "shared", "services"}),
)

opts := []rules.Option{rules.WithModel(m)}
```

Because `DomainDir` is `""`, `NewModel` sets `InternalTopLevel` to `{"pkg": true}` only. You need to explicitly allow the sublayer directories:

```go
// After NewModel, patch InternalTopLevel if using flat layout
m.InternalTopLevel = map[string]bool{
    "publisher":  true,
    "subscriber": true,
    "broker":     true,
    "dto":        true,
    "pkg":        true,
}
```

> This is a known rough edge with flat layouts. `NewModel` auto-populates `InternalTopLevel` from the three directory fields, not from `Sublayers`. For flat layouts, patch it manually after construction.

---

## Inheritance behavior

`NewModel(...)` always starts from `DDD()` defaults, then applies your options in order.

Key consequence: **options you don't set keep their DDD values.** For example:
- `BannedPkgNames` stays as the DDD default unless you call `WithBannedPkgNames`.
- `RequireAlias` stays `true` unless you call `WithRequireAlias(false)`.
- `Direction` stays as the DDD direction map unless you call `WithDirection`.

When building a model that is substantially different from DDD, set every relevant option explicitly. Partial overrides that assume DDD defaults can produce surprising violations.

`InternalTopLevel` is always recomputed by `NewModel` from `DomainDir`, `OrchestrationDir`, and `SharedDir` after all options are applied. Any manual assignment to `InternalTopLevel` inside an option will be overwritten. Set it manually *after* `NewModel` returns if you need fine-grained control (flat layouts especially).

---

## When to use a preset vs a custom Model

| Situation | Recommendation |
|-----------|----------------|
| Your layout matches DDD, CleanArch, Layered, Hexagonal, ModularMonolith, ConsumerWorker, Batch, or EventPipeline | Use the preset directly |
| Your layout matches a preset but you rename one or two layers | `NewModel` with `WithSublayers` + `WithDirection` overrides |
| Your layout matches a preset but you change directory names (`domain/` → `module/`) | `NewModel` with `WithDomainDir` + relevant overrides |
| Your architecture is a variant of a preset with extra sublayers | `NewModel` starting from DDD, add layers to `Sublayers` and `Direction` |
| Your architecture is fundamentally different (flat layout, non-standard grouping) | `NewModel` with explicit everything; patch `InternalTopLevel` after construction |

The eight presets cover most Go service shapes. Start there and narrow down which fields you actually need to override before writing a full custom model.
