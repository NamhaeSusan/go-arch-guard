# Design: Skill-Driven Plugin Version Bump Action

## Purpose

Automatically bump the Claude Code plugin version when the skill definition changes, so marketplace metadata stays in sync with meaningful skill updates.

## Scope

The automation targets changes to:

- `SKILL.md`
- `plugins/go-arch-guard/skills/go-arch-guard/SKILL.md`

When either file changes on `main`, the workflow updates:

- `plugins/go-arch-guard/.claude-plugin/plugin.json`

## Version Policy

- Initial version in the repository: `0.0.1`
- Normal bump: patch only
- Patch rollover: `0.0.99 -> 0.1.0`
- Minor rollover beyond two digits is rejected: `0.99.99` fails
- Each version component must stay within `0..99`

## Chosen Approach

Use a small Python script plus a GitHub Actions workflow.

- Python script handles version parsing, rollover, and JSON rewrite
- Workflow handles change detection, execution, and commit/push
- Unit tests cover the version policy independently from GitHub Actions

This keeps the rollover logic out of YAML and makes it easy to validate locally.

## Files

- `scripts/bump_plugin_version.py`
  - parse, validate, and bump semantic version under the repository rule set
- `scripts/test_bump_plugin_version.py`
  - unit tests for normal bump, rollover, overflow, and file rewrite
- `.github/workflows/plugin-version-bump.yml`
  - triggers on skill changes to `main`, runs the script, commits updated `plugin.json`
- `.github/workflows/ci.yml`
  - runs the Python unit tests
- `plugins/go-arch-guard/.claude-plugin/plugin.json`
  - reset initial version to `0.0.1`

## Safety

- Workflow triggers only when skill files change
- Commit occurs only if `plugin.json` changed
- Bot commit message is fixed and does not retrigger the same workflow because the workflow path filter excludes `plugin.json`

## Non-Goals

- No automatic skill content synchronization
- No release tagging
- No major/minor inference from commit messages or PR labels
