# Domain Import Whitelist Design

**Date:** 2026-03-23

## Goal

Tighten domain import rules so only the intended composition and orchestration layers may depend on domain root packages, while keeping `internal/pkg` domain-unaware.

## Decisions

1. `internal/orchestration` may import domain root packages only.
2. `cmd/...` may import domain root packages only.
3. `internal/pkg/...` must never import domains or orchestration.
4. Other non-domain `internal/*` packages must not import domains at all.
5. Top-level `internal/router` and `internal/bootstrap` are not part of v1. Their responsibilities move to:
   - `cmd/...` for app-specific wiring and route registration
   - `internal/pkg/...` for shared transport and middleware helpers
6. Violation file paths should be project-relative across all rule checks.
7. `analyzer.Load` should not require full type information because the current rule set is syntax/import based.

## Implications

- Whitelist policy becomes explicit: domain-aware code lives in `internal/domain`, `internal/orchestration`, and `cmd`.
- `internal/pkg` becomes a strict support layer with no business knowledge.
- Packages like `internal/config` may exist, but they must not import domains directly.
- README should remove `router/bootstrap` as privileged layers and explain `cmd` as the composition root.

## Expected Rule Shape

- `isolation.cross-domain`
- `isolation.orchestration-deep-import`
- `isolation.cmd-deep-import`
- `isolation.pkg-imports-domain`
- `isolation.internal-imports-domain`
- `structure.legacy-package` should cover `router` and `bootstrap`

## API Quality Targets

- Naming violations should report relative paths.
- Snake-case and other file-based naming checks should report relative paths.
- `analyzer.Load` should load package syntax/import info without type-check-only failures blocking analysis.
