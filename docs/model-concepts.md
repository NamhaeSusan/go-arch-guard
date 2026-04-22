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

**Why:** `CheckLayerDirection` uses this list to detect unknown sublayers. In domain layouts, a package that lands in a directory not named in `Sublayers` triggers `layer.unknown-sublayer`. In flat layouts, unknown top-level directories under `internal/` are classified as `kindUnclassified` and silently skipped — no violation is emitted. For flat custom models, keep `Sublayers` and `InternalTopLevel` aligned manually so that `CheckStructure` (which does enforce `InternalTopLevel`) catches accidental additions.

**Example (DDD):**

```go
rules.WithSublayers([]string{
    "handler", "app", "core", "core/model",
    "core/repo", "core/svc", "event", "infra",
})
```

Slash notation (`core/model`) represents a subdirectory inside `core/`. The rule matches on the suffix of the package path, so `internal/domain/order/core/model` matches `core/model`.

**Common mistake (domain layouts):** adding a new directory (e.g. `cache/`) inside a domain without adding it here. Every package in that directory will fail with `layer.unknown-sublayer`.

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

**Gotcha — missing key vs. empty slice:** these are not equivalent.

- `"foo": {}` — explicit deny-all: `foo` may import nothing (except `SharedDir`). Any cross-layer import from `foo` triggers `layer.direction`.
- missing key for `foo` — direction enforcement is **disabled** for `foo`: the rule skips it entirely (`if !known { continue }`). Imports from `foo` are not checked.

Always add an explicit entry for every layer you want enforced. A layer absent from `Direction` is effectively ungoverned, which is a silent trap when building custom models.

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

### `AppDir`

**What:** the name of the top-level directory under `internal/` that acts as the composition root (DI wiring, container setup).

**Why:** packages under `AppDir` are classified as `kindApp`. The isolation rule grants them unrestricted import privileges — they are the composition root and must be able to wire together any combination of domain, orchestration, and shared packages.

**Set to `""`** to disable. When non-empty, `NewModel` adds it to `InternalTopLevel` automatically.

**Default in DDD:** `"app"` (`internal/app/`).

**Example:**

```go
rules.WithAppDir("container")  // internal/container/
rules.WithAppDir("")           // disable composition-root privilege
```

---

### `ServerDir`

**What:** the name of the top-level directory under `internal/` that groups transport layers. Any subdirectory under `ServerDir` is treated as a protocol-specific transport (e.g. `server/http`, `server/grpc`). No protocol whitelist — any subdirectory counts.

**Why:** packages under `ServerDir/<proto>/` are classified as `kindTransport`. The isolation rule restricts them: they may only import `AppDir` (composition root), `SharedDir` (shared utilities), or other transport packages. Imports of domain packages, orchestration, or unclassified internal packages from transport layers trigger violations.

Rule IDs emitted for transport source violations:
- `isolation.transport-imports-domain` — transport imports domain (root or sub-package) directly
- `isolation.transport-imports-orchestration` — transport imports orchestration directly
- `isolation.transport-imports-unclassified` — transport imports an unclassified internal package (e.g. `internal/config`, `internal/bootstrap`)

All three enforce the pattern where HTTP/gRPC handlers depend on the app container (or shared pkg), not on arbitrary internal code.

**Set to `""`** to disable. When non-empty, `NewModel` adds it to `InternalTopLevel` automatically.

**Default in DDD:** `"server"` (`internal/server/`).

**Example:**

```go
rules.WithServerDir("transport")  // internal/transport/http/, internal/transport/grpc/
rules.WithServerDir("")           // disable transport isolation
```

---

### `InternalTopLevel`

**What:** the set of directory names allowed directly under `internal/`.

**Why:** `CheckStructure` emits `structure.internal-top-level` for any directory under `internal/` that is not in this set. This catches typos and accidental package placement (e.g. someone creates `internal/utils/` by mistake).

**You should not set this manually.** `NewModel(...)` derives it automatically after applying all options:

- **Domain layout** (`DomainDir != ""`): `InternalTopLevel = {DomainDir, OrchestrationDir, SharedDir}` (non-empty values only).
- **Flat layout** (`DomainDir == ""`): `InternalTopLevel` is populated from every entry in `Sublayers`, plus `OrchestrationDir` and `SharedDir` when non-empty.

This means flat custom models built with `NewModel` + `WithSublayers(...)` automatically allow those directories under `internal/` without any manual patching.

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

### `PortLayers` and `ContractLayers`

**What:** explicit lists of sublayer names that the library treats as port layers and contract layers respectively.

**PortLayers:** sublayers that only declare interfaces — pure contracts for external dependencies (repository interfaces, gateway interfaces). No implementations, no concrete types.

**ContractLayers:** a broader set that includes port layers plus service-interface layers (`svc`-like layers). Semantically `ContractLayers ⊇ PortLayers`. Helpers union the two lists at check time, so you do not need to repeat port names in `ContractLayers`.

**Which rules use these:** rules that check alias re-export patterns (`structure.alias-*`) consult `matchContractSublayer` to identify interface-only layers. Port detection is also used by `CheckStructure`'s alias checks to determine whether a domain export surface is exposing implementation details.

Note: `interface.single-per-package` and `interface.exported-impl` are NOT gated by port/contract detection — they run on all packages not listed in `InterfacePatternExclude`.

**Defaults:**

| Preset | `PortLayers` | `ContractLayers` |
|--------|-------------|-----------------|
| `DDD()` | `["core/repo"]` | `["core/repo", "core/svc"]` |
| `CleanArch()` | `["gateway"]` | `["gateway"]` |
| All others | `[]` (basename fallback) | `[]` (basename fallback) |

**Basename fallback:** when both `PortLayers` and `ContractLayers` are empty, the helpers fall back to hardcoded basename matching — sublayers whose last path component is `repo` or `gateway` are treated as ports; add `svc` for contracts. This preserves backward compatibility for presets and custom models that don't set these fields.

**When to set:** if your custom model uses a non-standard name for port layers (e.g. `store`, `boundary`, `outbound`), set `PortLayers` explicitly so the library recognizes them. If you want the basename fallback despite inheriting DDD defaults, clear both with `WithPortLayers(nil)` and `WithContractLayers(nil)`.

```go
// Custom model with a "store" port layer
m := rules.NewModel(
    rules.WithSublayers([]string{"handler", "service", "store", "model"}),
    rules.WithPortLayers([]string{"store"}),
    rules.WithContractLayers([]string{"store"}), // no svc-equivalent here
    // ... other options
)

// Clear to force basename fallback (undoes DDD defaults)
m := rules.NewModel(
    rules.WithPortLayers(nil),
    rules.WithContractLayers(nil),
)
```

**Gotcha:** `WithSublayers` does NOT clear `PortLayers`/`ContractLayers`. If you replace sublayers with a completely different set, the inherited DDD port/contract lists may no longer match any sublayer in your model — effectively becoming dead config. Set them explicitly whenever you call `WithSublayers` with a non-DDD list.

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

**To override the list** (the defaults are unexported; copy and extend as needed):

```go
rules.WithBannedPkgNames([]string{
    "util", "common", "misc", "helper", "shared", "services", // defaults
    "managers", "handlers", // team additions
})
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

Because `DomainDir` is `""`, `NewModel` automatically promotes every entry in `Sublayers` into `InternalTopLevel`, plus `SharedDir`. The resulting `InternalTopLevel` will be `{publisher, subscriber, broker, dto, pkg}` — no manual patching needed.

---

## Inheritance behavior

`NewModel(...)` always starts from `DDD()` defaults, then applies your options in order.

Key consequence: **options you don't set keep their DDD values.** For example:
- `BannedPkgNames` stays as the DDD default unless you call `WithBannedPkgNames`.
- `RequireAlias` stays `true` unless you call `WithRequireAlias(false)`.
- `Direction` stays as the DDD direction map unless you call `WithDirection`.

When building a model that is substantially different from DDD, set every relevant option explicitly. Partial overrides that assume DDD defaults can produce surprising violations.

`InternalTopLevel` is always recomputed by `NewModel` after all options are applied. Any manual assignment to `InternalTopLevel` inside an option will be overwritten. For domain layouts the set is derived from `DomainDir`, `OrchestrationDir`, and `SharedDir`. For flat layouts it is derived from `Sublayers` plus `OrchestrationDir` and `SharedDir`. If you need a directory in `InternalTopLevel` that doesn't come from those sources, set it on the returned model directly after `NewModel` returns.

---

## When to use a preset vs a custom Model

| Situation | Recommendation |
|-----------|----------------|
| Your layout matches DDD, CleanArch, Layered, Hexagonal, ModularMonolith, ConsumerWorker, Batch, or EventPipeline | Use the preset directly |
| Your layout matches a preset but you rename one or two layers | `NewModel` with `WithSublayers` + `WithDirection` overrides |
| Your layout matches a preset but you change directory names (`domain/` → `module/`) | `NewModel` with `WithDomainDir` + relevant overrides |
| Your architecture is a variant of a preset with extra sublayers | `NewModel` starting from DDD, add layers to `Sublayers` and `Direction` |
| Your architecture is fundamentally different (flat layout, non-standard grouping) | `NewModel` with explicit everything; flat layouts auto-populate `InternalTopLevel` from `Sublayers` |

The eight presets cover most Go service shapes. Start there and narrow down which fields you actually need to override before writing a full custom model.
