# Architecture Hardening Design

**Date:** 2026-03-23

## Goal

Close the remaining architecture escape hatches so the documented vertical-slice model matches the enforced rule set.

## Approved Decisions

1. `internal/orchestration` is a protected outer layer. Only `cmd/...` and `internal/orchestration/...` may import it.
2. `core`, `core/model`, `core/repo`, `core/svc`, and `event` are inner layers and must not import `internal/pkg/...`.
3. This tightening ships as the default behavior and reports `Error` violations.
4. `event` remains a first-class sublayer.
5. `WithExclude` should be normalized around project-relative paths.
6. Structure rules should reject recursive legacy/banned package names and strengthen domain root/model checks.

## Design

### 1. Orchestration Import Whitelist

`internal/orchestration` is the cross-domain coordination layer. Allowing arbitrary `internal/*` packages to import it creates hidden composition roots and weakens the role of `cmd/...`.

- Keep dedicated violations for `pkg -> orchestration` and `domain -> orchestration`.
- Add a broader rule for other internal packages importing orchestration.
- Allow:
  - `cmd/... -> internal/orchestration/...`
  - `internal/orchestration/... -> internal/orchestration/...`
- Reject:
  - any other `internal/* -> internal/orchestration/...`

### 2. Inner Layers Must Stay `pkg`-Free

The current implementation treats `internal/pkg/...` as universally safe. That makes it an escape hatch for domain logic and conflicts with the documented “dependency-free center”.

- Add a new layer rule for `core`, `core/model`, `core/repo`, `core/svc`, and `event` importing `internal/pkg/...`.
- Keep `handler`, `app`, `infra`, and domain root `alias.go` allowed to import `internal/pkg/...`.
- This is a hard rule, not an opt-in mode.

### 3. `event` Dependency Story

`event` remains in v1, so its allowed imports must be explicit.

- Allow `event -> core/model`
- Allow `app -> event`
- Allow `infra -> event`
- Reject `core/* -> event`
- Reject `handler -> event`

### 4. Exclude Path Semantics

The current API mixes package-path matching and file-path matching. That makes adoption error-prone.

- Normalize matching around project-relative paths in docs and tests.
- Preserve compatibility with module-qualified paths where practical so existing users do not break immediately.

### 5. Structure Hardening

Filesystem checks should close obvious naming/location bypasses.

- Ban `util`, `common`, `misc`, `helper`, `shared` recursively under `internal/`.
- Treat `router` and `bootstrap` as legacy names recursively under `internal/`.
- Treat `app`, `handler`, and `infra` as location-sensitive directories:
  - allowed only under `internal/domain/<name>/...`
  - `handler` also allowed under `internal/orchestration/...`
  - elsewhere they are legacy violations
- Strengthen domain root validation:
  - `alias.go` must exist
  - `alias.go` package name must match the domain name
- Strengthen model validation:
  - `core/model` must contain at least one non-test `.go` file when used

## Rule IDs

- `isolation.internal-imports-orchestration`
- `layer.inner-imports-pkg`
- `structure.domain-root-alias-package`

## Files Expected To Change

- `rules/isolation.go`
- `rules/layer.go`
- `rules/structure.go`
- `rules/rule.go`
- `rules/helpers.go`
- `rules/isolation_test.go`
- `rules/layer_test.go`
- `rules/structure_test.go`
- `integration_test.go`
- `README.md`
- `testdata/invalid/...`

## Verification

- Focused red/green runs for updated rule groups
- Full `go test ./...`
- Full `make lint`
