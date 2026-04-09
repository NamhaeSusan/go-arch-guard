# go-arch-guard

## Concept

Go 프로젝트의 아키텍처 규칙(의존성, 네이밍, 구조)을 정적 분석으로 검증하는 라이브러리이며, AI 코딩 에이전트가 쉽게 스캐폴딩하고 유지할 수 있는 표면을 제공한다.

AI 에이전트 친화적인 기본 surface:
- `scaffold.ArchitectureTest(...)` — 프리셋별 `architecture_test.go` 생성
- `report.BuildJSONReport(...)` / `report.MarshalJSONReport(...)` — machine-readable violation 출력
- `rules.RunAll(...)` — 권장 기본 rule 묶음 실행

---

### Post-Implementation Checklist (CRITICAL)

After any code change, update related documentation (README.md, docs/).
CLAUDE.md and README.md must be kept in sync.
SKILL.md (`plugins/go-arch-guard/skills/go-arch-guard/SKILL.md`) must also be updated when rules or API surface changes.
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

#### Warning Category vs Hard Rules

The Non-Goals above apply to **hard-enforcing rules** (Error severity) that block builds. They do **not** apply to **Warning-severity smell detectors** that inform without blocking, as long as the smell:

- is **mechanically detectable** with low false-positive risk,
- correlates with a real vibe-coding regression (not theoretical purity),
- has a clearly stated motivation in `docs/` or the rule's godoc,
- defaults to `Warning` severity (callers can opt into `Error` via `WithSeverity`),
- and is independent of any specific project layout.

A Warning surfaces a smell to the developer without forcing a fix. This is different from "policing implementation details" because the developer remains in control. Examples that fit this category:

- `interface.container-only` — interface declared but used only as a struct field, never as a function parameter or return type. This catches the wiring-layer workaround pattern where a developer declares a local interface just to give a struct field a type.

When adding a new Warning-severity rule, document the motivation (the *why*, not just the *what*) and prefer to point at the *cause* of the smell rather than the *symptom*.

#### When a Hard Rule Is Justified

A hard rule (Error severity) is justified when:

- the violation creates a parallel or uncontrolled architectural surface that bypasses an existing controlled surface (e.g. `alias.go`),
- the project has explicitly committed to that controlled surface as a convention,
- the violation is mechanically detectable with low false-positive risk,
- and the fix is clear (use the controlled surface instead).

Example:

- `interface.cross-domain-anonymous` — anonymous interface in any package whose method signatures touch a foreign domain's types. This bypasses `alias.go`'s role as the sole controlled cross-domain public surface, creating a second uncontrolled surface inline. Severity Error because the project commits to alias-only cross-domain access, and the fix is to expose a small named interface in the target domain's `alias.go` and use that named type.
