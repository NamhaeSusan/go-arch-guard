# New Architecture Model Presets

**Date:** 2026-03-31
**Status:** Approved

## Summary

Add three new built-in architecture model presets to go-arch-guard:
`Layered()`, `Hexagonal()`, and `ModularMonolith()`.

These complement the existing `DDD()` and `CleanArch()` presets,
covering the most common Go server project structures in practice.

## Presets

### Layered() — Spring-style 4-tier

```
handler → service → {repository, model}
repository → model
model → {}
```

| Field | Value |
|---|---|
| Sublayers | handler, service, repository, model |
| PkgRestricted | model |
| RequireAlias | false |
| RequireModel | false |
| ModelPath | model |
| DTOAllowedLayers | handler, service |

### Hexagonal() — Ports & Adapters

```
handler → usecase → {port, domain}
adapter → {port, domain}
port → domain
domain → {}
```

| Field | Value |
|---|---|
| Sublayers | handler, usecase, port, domain, adapter |
| PkgRestricted | domain |
| RequireAlias | false |
| RequireModel | false |
| ModelPath | domain |
| DTOAllowedLayers | handler, usecase |

### ModularMonolith() — Module-based 4-tier

```
api → application → domain
infrastructure → domain
domain → {}
```

| Field | Value |
|---|---|
| Sublayers | api, application, domain, infrastructure |
| PkgRestricted | domain |
| RequireAlias | false |
| RequireModel | false |
| ModelPath | domain |
| DTOAllowedLayers | api, application |

## Common Settings (all three)

- DomainDir: "domain"
- OrchestrationDir: "orchestration"
- SharedDir: "pkg"
- AliasFileName: "alias.go"
- BannedPkgNames: util, common, misc, helper, shared, services
- LegacyPkgNames: router, bootstrap
- InternalTopLevel: {domain, orchestration, pkg}

## LayerDirNames per preset

Each preset includes its own sublayer names plus shared names
(controller, entity, store, persistence, domain, service).

- **Layered:** handler, service, repository, model
- **Hexagonal:** handler, usecase, port, domain, adapter
- **ModularMonolith:** api, application, domain, infrastructure

## Scope

- Add factory functions in `rules/model.go`
- Add unit tests in `rules/model_test.go` (consistency validation)
- Add integration tests in `integration_test.go` (valid + violation cases)
- Update `TestModelConsistency` to cover all five presets
- Update README.md, README.ko.md, SKILL.md
