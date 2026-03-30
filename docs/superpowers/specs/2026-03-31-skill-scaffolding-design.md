# go-arch-guard Skill — Scaffolding Feature

**Date:** 2026-03-31
**Status:** Approved

## Summary

Add scaffolding capability to the existing `go-arch-guard` SKILL.md.
When a user selects a preset, the skill guides creation of:

1. `internal/domain/`, `internal/orchestration/`, `internal/pkg/` directories
2. Preset-specific `architecture_test.go` at project root

## Scaffolding Output

### All presets (common)

```
internal/
├── domain/          # empty
├── orchestration/   # empty
└── pkg/             # empty
architecture_test.go # preset-specific
```

### architecture_test.go per preset

Each preset generates a complete test file with the appropriate model:

- **DDD** — no `WithModel` (default)
- **CleanArch** — `rules.WithModel(rules.CleanArch())`
- **Layered** — `rules.WithModel(rules.Layered())`
- **Hexagonal** — `rules.WithModel(rules.Hexagonal())`
- **ModularMonolith** — `rules.WithModel(rules.ModularMonolith())`

All templates include: CheckDomainIsolation, CheckLayerDirection,
CheckNaming, CheckStructure, AnalyzeBlastRadius.

## SKILL.md Changes

1. **description** — mention scaffolding capability
2. **When to Use** — add "new project scaffolding" trigger
3. **New section "Quick Init"** — scaffolding flow:
   - Check project state (go.mod, internal/, architecture_test.go)
   - New project → preset selection → scaffold
   - Existing project → existing guide flow
4. **Existing content preserved** — scaffolding is an additional path

## Skill Flow

```
User triggers skill
  ↓
Check project state:
  - go.mod exists?
  - internal/ exists?
  - architecture_test.go exists?
  ↓
New project → Quick Init (select preset → scaffold)
Existing project → existing guide (current SKILL.md content)
```

## Constraints

- `go get` requires user confirmation before execution
- Check for existing files before creating (no overwrites)
- Empty directories created without .gitkeep
- Domain subdirectories NOT created (user adds their own)
