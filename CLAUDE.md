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
