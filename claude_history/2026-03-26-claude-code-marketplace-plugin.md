# Claude Code marketplace plugin wrapper

## Summary

Added a Claude Code marketplace/plugin wrapper around the existing `go-arch-guard` skill while keeping the root `SKILL.md` workflow intact.

## Packaging

- Added repository marketplace manifest at `.claude-plugin/marketplace.json`
- Added plugin manifest at `plugins/go-arch-guard/.claude-plugin/plugin.json`
- Packaged the skill for Claude Code at `plugins/go-arch-guard/skills/go-arch-guard/SKILL.md`

## Documentation

- `README.md` — documented marketplace install and local validation
- `SKILL.md` — added Claude Code plugin install path alongside `go get`
- `CLAUDE.md` — documented that the plugin-packaged skill copy must stay in sync with the root `SKILL.md`
- `docs/superpowers/specs/2026-03-26-claude-code-marketplace-plugin-design.md` — added design note
- `docs/superpowers/plans/2026-03-26-claude-code-marketplace-plugin.md` — added implementation plan

## Files Changed

- `.claude-plugin/marketplace.json`
- `plugins/go-arch-guard/.claude-plugin/plugin.json`
- `plugins/go-arch-guard/skills/go-arch-guard/SKILL.md`
- `README.md`
- `SKILL.md`
- `CLAUDE.md`
- `docs/superpowers/specs/2026-03-26-claude-code-marketplace-plugin-design.md`
- `docs/superpowers/plans/2026-03-26-claude-code-marketplace-plugin.md`
- `claude_history/2026-03-26-claude-code-marketplace-plugin.md`

## Verification

- `claude plugin validate .` — PASS
- `go test ./...` — PASS
- `make lint` — PASS (`go vet ./...`, `0 issues.`)
