# Event-Driven Pipeline Preset

**Date:** 2026-04-03
**Status:** Approved

## Summary

Add an `EventPipeline()` flat-layout preset for full event-sourcing / CQRS projects.
This is the most complex flat-layout preset with 7 sublayers, a non-linear direction
graph, and two TypePattern entries.

## Target Structure

```
cmd/pipeline/main.go
internal/
  command/       # command_create_order.go → CreateOrderCommand.Execute()
  aggregate/     # aggregate_order.go → OrderAggregate.Apply()
  event/         # event definitions (OrderCreated, OrderShipped, etc.)
  projection/    # read model builders
  eventstore/    # append-only event storage
  readstore/     # projection CRUD storage
  model/         # shared types
  pkg/           # shared infra
```

## Layer Direction

```
command    → aggregate, eventstore, model, pkg
aggregate  → event, model, pkg
event      → model
projection → event, readstore, model, pkg
eventstore → event, model, pkg
readstore  → model, pkg
model      → (none)
pkg        → (none)
```

## Preset-Specific Rules

### TypePatterns (2 entries)

| Dir | FilePrefix | TypeSuffix | RequireMethod |
|-----|-----------|-----------|--------------|
| `command` | `command` | `Command` | `Execute` |
| `aggregate` | `aggregate` | `Aggregate` | `Apply` |

Examples:
- `command/command_create_order.go` → must define `CreateOrderCommand` with `Execute` method
- `aggregate/aggregate_order.go` → must define `OrderAggregate` with `Apply` method

### PkgRestricted

`model` and `event` — neither may import `pkg/`.

### Existing Rules

Same flat-layout behavior as ConsumerWorker/Batch: `structure.internal-top-level`,
`layer.direction`, `layer.inner-imports-pkg`, naming rules, blast radius.
Domain isolation skipped (`DomainDir == ""`).

## EventPipeline() Factory

```go
func EventPipeline() Model {
    return Model{
        Sublayers: []string{
            "command", "aggregate", "event", "projection",
            "eventstore", "readstore", "model",
        },
        Direction: map[string][]string{
            "command":    {"aggregate", "eventstore", "model"},
            "aggregate":  {"event", "model"},
            "event":      {"model"},
            "projection": {"event", "readstore", "model"},
            "eventstore": {"event", "model"},
            "readstore":  {"model"},
            "model":      {},
        },
        PkgRestricted: map[string]bool{"model": true, "event": true},
        InternalTopLevel: map[string]bool{
            "command": true, "aggregate": true, "event": true,
            "projection": true, "eventstore": true, "readstore": true,
            "model": true, "pkg": true,
        },
        DomainDir:        "",
        OrchestrationDir: "",
        SharedDir:        "pkg",
        RequireAlias:     false,
        AliasFileName:    "",
        RequireModel:     false,
        ModelPath:        "model",
        DTOAllowedLayers: []string{"command", "projection"},
        BannedPkgNames:   []string{"util", "common", "misc", "helper", "shared", "services"},
        LegacyPkgNames:   []string{"router", "bootstrap"},
        LayerDirNames: map[string]bool{
            "command": true, "aggregate": true, "event": true,
            "projection": true, "eventstore": true, "readstore": true,
            "model": true,
        },
        TypePatterns: []TypePattern{
            {Dir: "command", FilePrefix: "command", TypeSuffix: "Command", RequireMethod: "Execute"},
            {Dir: "aggregate", FilePrefix: "aggregate", TypeSuffix: "Aggregate", RequireMethod: "Apply"},
        },
    }
}
```

## Scaffold

Add `PresetEventPipeline` (`"event-pipeline"`) to scaffold.go.

## Scope

| File | Action |
|---|---|
| `rules/model.go` | Add `EventPipeline()` factory |
| `rules/model_test.go` | Add test + consistency table entry |
| `scaffold/scaffold.go` | Add `PresetEventPipeline` |
| `scaffold/scaffold_test.go` | Add test |
| `integration_test.go` | Add valid + violation integration tests |
| `skill_test.go` | Add `TestSkill_EventPipelineSetup` |
| `plugins/.../go-arch-guard-event-pipeline/SKILL.md` | Create preset skill |
| `plugins/.../plugin.json` | Bump version 0.0.9 → 0.0.10 |
| `README.md`, `README.ko.md` | Add preset docs |
| `plugins/.../go-arch-guard/SKILL.md` | Update main skill reference |
