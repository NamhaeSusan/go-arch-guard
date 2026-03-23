# Rule Hardening Design

**Date:** 2026-03-23

## Goal

Close the gaps in architecture enforcement so the rule set becomes more explicit and harder to bypass.

## Approved Scope

1. Domain packages may import only packages inside the same domain or `internal/pkg/...`.
2. Any domain sublayer not explicitly modeled by the library must be rejected.
3. Every domain root must contain `alias.go`.
4. Missing regression coverage around the existing structure and isolation rules must be added.

## Design

### 1. Domain to Orchestration Reverse Dependency

`CheckDomainIsolation` currently restricts how `orchestration` imports domains, but it does not reject the reverse dependency. This should become an explicit violation because orchestration is an outer coordination layer.

- Add a new isolation rule for domain packages importing orchestration.
- Keep `pkg -> orchestration` as its own existing violation because it conveys a different architectural intent.

### 2. Unknown Domain Sublayers

`CheckLayerDirection` should operate on a closed set of known sublayers. If a package lives under `internal/domain/<name>/<unknown>/...`, the library should emit a dedicated violation instead of silently skipping it.

- Keep the current direction table for known sublayers.
- Add a dedicated rule for unrecognized sublayers so users can distinguish “bad direction” from “undefined architecture surface”.

### 3. Domain Root Public Surface

The current structure rule prevents extra non-test Go files in the domain root, but it does not require `alias.go` to exist. That weakens the contract that the domain root is the only public API surface.

- Add a dedicated structure rule when `internal/domain/<name>/alias.go` is missing.
- Preserve the existing “alias-only” rule for extra root files.

### 4. Test Coverage

Add focused regression tests for:

- `domain -> orchestration` reverse dependency
- unknown domain sublayer rejection
- missing `alias.go`
- existing but under-tested structure rules:
  - `structure.domain-model-required`
  - `structure.dto-placement`
- existing but under-tested isolation rule:
  - `isolation.pkg-imports-domain`

## Rule IDs

- `isolation.domain-imports-orchestration`
- `layer.unknown-sublayer`
- `structure.domain-root-alias-required`

## Files Expected To Change

- `rules/isolation.go`
- `rules/layer.go`
- `rules/structure.go`
- `rules/isolation_test.go`
- `rules/layer_test.go`
- `rules/structure_test.go`
- `integration_test.go`
- `README.md`
- `testdata/invalid/...`

## Verification

- Targeted red/green test runs for each rule group
- Full `go test ./...`
- `make lint`
