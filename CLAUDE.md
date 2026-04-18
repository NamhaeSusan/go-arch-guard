# go-arch-guard

## Concept

Go 프로젝트의 아키텍처 규칙(의존성, 네이밍, 구조)을 정적 분석으로 검증하는 라이브러리. 팀이 커밋한 컨벤션을 vibe coding 중에 깨지지 않게 지켜주는 **중립 인프라**이며, AI 코딩 에이전트가 쉽게 스캐폴딩하고 유지할 수 있는 표면을 제공한다.

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

This library is neutral infrastructure for enforcing **team-owned conventions**.
The question "is this convention worth enforcing?" is a team decision, not a
library decision. If a team commits to a convention — however micro — and
violations can be detected mechanically with low false positives, the library
should be able to express it.

#### What the library ships

- Rule primitives that encode common vibe-coding failure modes (cross-domain
  imports, layer direction, placement drift, container interfaces, tx boundary,
  etc.)
- Presets (DDD, CleanArch, Hexagonal, ModularMonolith, ConsumerWorker, Batch,
  EventPipeline) that bundle sensible defaults for common team shapes.
- Extension points so teams can add their own rules or tune existing ones.

#### What the team decides

- Which conventions are worth enforcing in their codebase. "Domain core must
  be pure" or "alias.go is the only exposed surface" are legitimate team
  choices; the library provides rules for those if they are mechanically
  detectable.
- Severity per rule — `Error` (blocks builds) or `Warning` (advisory). Override
  via `WithSeverity`.
- False-positive tolerance. A rule that's noisy in one project may be a clean
  win in another.

#### Practical gates (apply to every rule, regardless of scope)

- **Mechanically detectable** — AST- or types-based, no guessing at intent.
- **Low false-positive risk** — a rule that fires on legitimate code is worse
  than no rule. This is the firmest gate; "micro" scope is fine, "noisy" is
  not.
- **Static analysis only** — no runtime dependencies.
- **Independent** — no hidden ordering or state sharing between rules.

#### When to add a rule to the recommended bundle

- A pattern that breaks during vibe coding as a project scales up, and a team
  convention that catches it.
- Motivation (the *why*, not just the *what*) documented in godoc or
  `docs/`.
- Severity chosen per how firmly the pattern actually breaks things in
  practice. The library does not require "parallel uncontrolled surface" as
  the sole trigger for `Error` severity — teams may promote a rule to `Error`
  when the convention is firm.

Example: `interface.cross-domain-anonymous` is shipped as `Error` because the
fix is clear (move the adapter into `internal/orchestration/`) and the
convention — cross-domain abstractions owned by the orchestration package —
is firm in every preset that enables it. Teams that use a different
orchestration convention can downgrade or exclude it.
