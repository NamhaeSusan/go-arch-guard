# Skill version bump action

## Summary

Added GitHub Actions automation that bumps the Claude Code plugin version when the skill definition changes on `main`.

## Behavior

- Watches `SKILL.md`
- Watches `plugins/go-arch-guard/skills/go-arch-guard/SKILL.md`
- Bumps `plugins/go-arch-guard/.claude-plugin/plugin.json`
- Uses repository version rule:
  - `0.0.1` start
  - patch bump by default
  - `0.0.99 -> 0.1.0`
  - reject `0.99.99`

## Files Changed

- `scripts/bump_plugin_version.py`
- `scripts/test_bump_plugin_version.py`
- `.github/workflows/plugin-version-bump.yml`
- `.github/workflows/ci.yml`
- `plugins/go-arch-guard/.claude-plugin/plugin.json`
- `README.md`
- `CLAUDE.md`
- `docs/superpowers/specs/2026-03-26-skill-version-bump-action-design.md`
- `docs/superpowers/plans/2026-03-26-skill-version-bump-action.md`
- `claude_history/2026-03-26-skill-version-bump-action.md`

## Verification

- `python3 -m unittest scripts/test_bump_plugin_version.py -v` — PASS
- `claude plugin validate .` — PASS
- `go test ./...` — PASS
- `make lint` — PASS
- `actionlint` — not run locally because it is not installed in this environment
