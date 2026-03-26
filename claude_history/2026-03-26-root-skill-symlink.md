# Root skill symlink

## Summary

Switched the repository to keep `plugins/go-arch-guard/skills/go-arch-guard/SKILL.md` as the real file and expose the root `SKILL.md` as a symlink to it.
Also clarified the Claude Code plugin installation steps in `README.md`.

## Files Changed

- `SKILL.md` — replaced with symlink to `plugins/go-arch-guard/skills/go-arch-guard/SKILL.md`
- `plugins/go-arch-guard/skills/go-arch-guard/SKILL.md` — restored as the real file with current content
- `CLAUDE.md` — updated maintenance guidance for the symlink layout
- `README.md` — made Claude Code plugin installation steps explicit
- `claude_history/2026-03-26-root-skill-symlink.md`

## Verification

- `claude plugin validate .` — PASS
- `go test ./...` — PASS
- `make lint` — PASS
