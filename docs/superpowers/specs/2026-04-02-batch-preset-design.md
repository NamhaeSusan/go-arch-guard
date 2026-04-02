# Batch Preset

**Date:** 2026-04-02
**Status:** Approved

## Summary

Add a `Batch()` flat-layout preset for cron/scheduler-triggered batch job projects.
Structurally identical to ConsumerWorker — the only differences are the entry-point
directory name (`job/` instead of `worker/`) and the TypePattern
(`job_xxx.go` → `XxxJob.Run` instead of `worker_xxx.go` → `XxxWorker.Process`).

## Target Structure

```
cmd/batch/main.go
internal/
  job/               # job_expire_files.go, job_cleanup_trash.go
  service/           # business logic, orchestration
  store/             # persistence (DB, external APIs)
  model/             # data structures
  pkg/               # shared infra (batchutil, logging, etc.)
```

## Layer Direction

```
job     → service, model, pkg
service → store, model, pkg
store   → model, pkg
model   → (none)
pkg     → (none)
```

## Preset-Specific Rules

### 1. Job Type Pattern (AST-based, reuses CheckTypePatterns)

In `job/`, files matching `job_*.go` must:
- `job_abc.go` → define exported type `AbcJob`
- `AbcJob` → have a `Run` method

Rule IDs (same as ConsumerWorker, different context):
- `naming.worker-type-mismatch` — file `job_abc.go` does not define `AbcJob`
- `naming.worker-missing-process` — `AbcJob` has no `Run` method

Note: Rule IDs are generic (`worker-type-mismatch` / `worker-missing-process`)
because they come from the shared `CheckTypePatterns` function. The message text
includes the actual expected type/method names so the developer sees `AbcJob`
and `Run`, not `Worker` or `Process`.

### 2. Existing Rules — same behavior as ConsumerWorker

All flat-layout rule behavior is inherited: `structure.internal-top-level`,
`layer.direction`, `layer.inner-imports-pkg`, naming rules, blast radius.
Domain isolation is skipped (`DomainDir == ""`).

## Batch() Factory

```go
func Batch() Model {
    return Model{
        Sublayers: []string{"job", "service", "store", "model"},
        Direction: map[string][]string{
            "job":     {"service", "model"},
            "service": {"store", "model"},
            "store":   {"model"},
            "model":   {},
        },
        PkgRestricted:    map[string]bool{"model": true},
        InternalTopLevel: map[string]bool{
            "job": true, "service": true,
            "store": true, "model": true, "pkg": true,
        },
        DomainDir:        "",
        OrchestrationDir: "",
        SharedDir:        "pkg",
        RequireAlias:     false,
        AliasFileName:    "",
        RequireModel:     false,
        ModelPath:        "model",
        DTOAllowedLayers: []string{"job", "service"},
        BannedPkgNames:   []string{"util", "common", "misc", "helper", "shared", "services"},
        LegacyPkgNames:   []string{"router", "bootstrap"},
        LayerDirNames: map[string]bool{
            "job": true, "service": true,
            "store": true, "model": true,
        },
        TypePatterns: []TypePattern{
            {Dir: "job", FilePrefix: "job", TypeSuffix: "Job", RequireMethod: "Run"},
        },
    }
}
```

## Scaffold

Add `PresetBatch` (`"batch"`) to scaffold.go.
Template generates test calling `rules.Batch()`.

## Scope

Since the flat-layout infrastructure (CheckLayerDirection flat branch,
CheckDomainIsolation skip, CheckStructure guard, CheckTypePatterns,
naming exemption) was already built for ConsumerWorker, this preset
only needs:

| File | Action |
|---|---|
| `rules/model.go` | Add `Batch()` factory |
| `rules/model_test.go` | Add `TestBatch_ReturnsValidModel`, add to consistency table |
| `scaffold/scaffold.go` | Add `PresetBatch` |
| `scaffold/scaffold_test.go` | Add test |
| `integration_test.go` | Add valid + violation integration tests |
| `README.md`, `README.ko.md` | Add Batch preset docs |
| `plugins/.../SKILL.md` | Update skill reference |
