# Architecture Concepts

This guide explains the `core.Architecture` model used by presets and rule implementations.
It is for teams that want to understand or hand-build an architecture instead
of using a preset unchanged.

## Overview

`core.Architecture` has four sub-models:

```go
type Architecture struct {
    Layers    core.LayerModel
    Layout    core.LayoutModel
    Naming    core.NamingPolicy
    Structure core.StructurePolicy
}
```

Presets such as `presets.DDD()` return ready-made `core.Architecture` values.
Rules read architecture through `ctx.Arch()` after you build a context:

```go
arch := presets.DDD()
ctx := core.NewContext(pkgs, "", "", arch, nil)
violations := core.Run(ctx, presets.RecommendedDDD())
```

## LayerModel

`LayerModel` owns the layer vocabulary. `Sublayers` is the single source of
truth: every field that names a layer must refer to a value in `Sublayers`.

| Field | Meaning |
|-------|---------|
| `Sublayers` | Authoritative list of layer names. Domain presets classify paths under each domain using this list; flat presets classify top-level `internal/` directories. |
| `Direction` | Allowed import matrix. Every sublayer must have a key, even if the allowed target list is empty. |
| `PortLayers` | Pure interface layers such as `core/repo`, `gateway`, or `port`. |
| `ContractLayers` | Layers that expose contracts. Must include every entry in `PortLayers`. |
| `PkgRestricted` | Layers that must not import the shared package tree. |
| `InternalTopLevel` | Top-level directories allowed under `internal/`. Structure rules use this to catch accidental directories. |
| `LayerDirNames` | Basename hints used by placement rules to identify layer-like directories. |

Example:

```go
layers := core.LayerModel{
    Sublayers: []string{"api", "logic", "data"},
    Direction: map[string][]string{
        "api":   {"logic"},
        "logic": {"data"},
        "data":  {},
    },
    PkgRestricted: map[string]bool{
        "data": true,
    },
    InternalTopLevel: map[string]bool{
        "module": true,
        "pkg":    true,
    },
    LayerDirNames: map[string]bool{
        "api": true, "logic": true, "data": true,
    },
}
```

## LayoutModel

`LayoutModel` describes the `internal/` directory topology. Empty fields
disable the corresponding classification.

| Field | Meaning |
|-------|---------|
| `DomainDir` | Directory that contains domain slices, for example `internal/domain/{name}/`. Empty means flat layout. |
| `OrchestrationDir` | Directory for cross-domain coordination, for example `internal/orchestration/`. |
| `SharedDir` | Shared internal package tree, usually `internal/pkg/`. |
| `AppDir` | Composition root such as `internal/app/`. Rules can allow this layer to wire everything. |
| `ServerDir` | Transport root such as `internal/server/http/` or `internal/server/grpc/`. |

Domain layout example:

```go
layout := core.LayoutModel{
    DomainDir:        "domain",
    OrchestrationDir: "orchestration",
    SharedDir:        "pkg",
    AppDir:           "app",
    ServerDir:        "server",
}
```

Flat layout example:

```go
layout := core.LayoutModel{
    DomainDir: "",
    SharedDir: "pkg",
}
```

## NamingPolicy

`NamingPolicy` carries naming-only conventions. It does not duplicate layer
names; layer-aware rules read from `LayerModel.Sublayers`.

| Field | Meaning |
|-------|---------|
| `BannedPkgNames` | Package names that should not appear under `internal/`, such as `util` or `common`. |
| `LegacyPkgNames` | Package names that produce migration warnings, such as `router` or `bootstrap`. |
| `AliasFileName` | Filename for domain root alias files, usually `alias.go`. |

## StructurePolicy

`StructurePolicy` carries placement and structural conventions.

| Field | Meaning |
|-------|---------|
| `RequireAlias` | Whether each domain root must define the alias file. |
| `RequireModel` | Whether each domain must contain a model directory. |
| `ModelPath` | Model directory path inside a domain, for example `core/model`. |
| `DTOAllowedLayers` | Sublayers where DTO files are allowed. Every value must appear in `LayerModel.Sublayers`. |
| `TypePatterns` | File/type/method conventions for flat layouts. |
| `InterfacePatternExclude` | Sublayers skipped by interface pattern checks. Every key must appear in `LayerModel.Sublayers`. |

Type pattern example:

```go
structure := core.StructurePolicy{
    TypePatterns: []core.TypePattern{
        {
            Dir:           "worker",
            FilePrefix:    "worker_",
            TypeSuffix:    "Worker",
            RequireMethod: "Process",
        },
    },
}
```

## Validation

`Architecture.Validate()` and `core.Validate(arch)` enforce the structural
invariants that rules rely on:

- every `Sublayers` entry has a `Direction` key
- every `Direction` source and target exists in `Sublayers`
- `PortLayers`, `ContractLayers`, `PkgRestricted`, `DTOAllowedLayers`, and
  `InterfacePatternExclude` reference only known layers
- `PortLayers` is a subset of `ContractLayers`

`core.Run` validates the architecture before running rules and panics if it is
invalid. Presets validate themselves when constructed so preset mistakes fail
early.

## Build Patterns

Prefer a preset when your layout matches one of the built-ins:

```go
arch := presets.DDD()
ctx := core.NewContext(pkgs, "", "", arch, nil)
report.AssertNoViolations(t, core.Run(ctx, presets.RecommendedDDD()))
```

Hand-build an architecture when your project shape is genuinely custom:

```go
arch := core.Architecture{
    Layers: core.LayerModel{
        Sublayers: []string{"publisher", "subscriber", "broker", "dto"},
        Direction: map[string][]string{
            "publisher":  {"broker", "dto"},
            "subscriber": {"broker", "dto"},
            "broker":     {"dto"},
            "dto":        {},
        },
        PkgRestricted: map[string]bool{
            "dto": true,
        },
        InternalTopLevel: map[string]bool{
            "publisher":  true,
            "subscriber": true,
            "broker":     true,
            "dto":        true,
            "pkg":        true,
        },
    },
    Layout: core.LayoutModel{
        DomainDir: "",
        SharedDir: "pkg",
    },
    Naming: core.NamingPolicy{
        BannedPkgNames: []string{"util", "common", "misc", "helper", "shared", "services"},
        LegacyPkgNames: []string{"router", "bootstrap"},
        AliasFileName:  "alias.go",
    },
    Structure: core.StructurePolicy{
        RequireAlias:     false,
        RequireModel:     false,
        DTOAllowedLayers: []string{"publisher", "subscriber"},
    },
}
if err := core.Validate(arch); err != nil {
    t.Fatal(err)
}
```

Use `presets.Recommended...()` as the matching rule bundle for a preset. For a
custom architecture, start with `core.NewRuleSet(...)` and include only the
rules whose assumptions match your layout.
