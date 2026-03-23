# 2026-03-23 rule-scope

## Summary

- Added a `Rule Scope (CRITICAL)` section to `CLAUDE.md`.
- Clarified that go-arch-guard is for coarse vibe-coding guardrails, not architecture purity.
- Added explicit non-goals for overreaching rule ideas such as domain-core purity, alias exposure policing, and `cmd` reverse-dependency arguments that Go already rejects via import cycles.

## Files Changed

- `CLAUDE.md`

## Verification

- Reviewed `git diff -- CLAUDE.md`
