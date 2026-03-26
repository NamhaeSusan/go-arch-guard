# go-arch-guard

## Concept

Go 프로젝트의 아키텍처 규칙(의존성, 네이밍, 구조)을 정적 분석으로 검증하는 라이브러리.

---

### Work Log (CRITICAL)

After every task, leave a record in the `claude_history/` folder.

- **Filename**: `yyyy-mm-dd-{work}.md` (e.g., `2026-02-10-security-audit.md`)
- **Content**: Brief summary of the task, files changed, and verification results
- For multiple tasks on the same day, distinguish by the work part

---

### Post-Implementation Checklist (CRITICAL)

After any code change, update related documentation (README.md, docs/).
CLAUDE.md and README.md must be kept in sync.
SKILL.md must also be updated when rules or API surface changes.
When `SKILL.md` changes, keep `plugins/go-arch-guard/skills/go-arch-guard/SKILL.md` in sync.
Skill changes on `main` also auto-bump `plugins/go-arch-guard/.claude-plugin/plugin.json` via GitHub Actions. Keep the versioning rule and workflow in sync with the script tests.

---

### Workflow (CRITICAL)

**Simple tasks** (single-file edits, typo fixes, one-line bug fixes) can be handled directly.

**All other tasks** must use **TeamCreate to form an agent team** for parallel execution:
- Implementation -> Implementation agents (e.g., `general-purpose`, project-specific agents)
- Testing/Validation -> Validation agents (e.g., `tdd-guide`, project-specific validators)
- Documentation -> Documentation agents (e.g., `doc-updater`, project-specific doc agents)
- Code review -> `code-reviewer` agent

Example team composition:
Phase 1 (parallel): Implementation agents -> Write code
Phase 2 (parallel): Validation agent + code-reviewer -> Verify
Phase 3: Documentation agent -> Update docs

---

### Core Design Principles

- **Rules are independent** — each rule must not depend on other rules
- **Static analysis only** — no runtime dependencies, pure code analysis

---

### Rule Scope (CRITICAL)

This library exists to add **coarse guardrails for vibe coding**, not to encode architecture purity or invent a new language on top of Go.

- The goal is to block **big-picture exceptions** that are easy to introduce during vibe coding:
  - unwanted package placement
  - broad import-direction violations
  - naming that breaks the intended project shape
- The goal is **not** to enforce theoretical purity or micro-level architectural doctrine.
- Prefer rules that control the **overall flow and boundaries** of the codebase.
- Avoid rules that try to police every corner case, implementation detail, or semantic nuance.

#### Explicit Non-Goals

Do **not** add or push for rules based on arguments like:

- "domain core must be pure"
- "alias.go must strictly control exposure"
- "cmd reverse dependency must be blocked"
- "this import is philosophically wrong even though Go already handles it"

If Go already rejects something by itself, such as import cycles, do not treat that as a special architecture rule target unless the user explicitly asks for it.

When proposing or implementing new rules, optimize for:

- obvious, large-scale guardrails
- low surprise
- low false-positive risk
- practical usefulness during vibe coding

Do **not** optimize for:

- ideology
- purity
- corner-case control
- over-modeling package internals
