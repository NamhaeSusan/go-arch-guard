# Internal Top-Level Whitelist Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Restrict `internal/` top-level packages to `domain`, `orchestration`, and `pkg`, and align README/tests with the enforced rule.

**Architecture:** Keep the policy hard-coded for this repository's company convention. Enforce it in `CheckStructure` so filesystem checks reject unexpected `internal/*` top-level packages before they become hidden architecture centers. Update docs and tests so the documented architecture matches actual behavior.

**Tech Stack:** Go, `go test`, existing rule/unit/integration tests

---

### Task 1: Add RED tests for the top-level whitelist

**Files:**
- Modify: `rules/structure_test.go`
- Modify: `integration_test.go`

- [ ] **Step 1: Write the failing tests**

Add:
- a structure test proving `internal/config/config.go` triggers a structure violation
- a structure test proving `internal/platform/platform.go` triggers the same violation
- update the integration test that currently proves neutral support packages are allowed so it now expects violations for `config/platform/system/foundation`

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./rules ./... -run 'TestCheckStructure|TestIntegration_InternalTopLevelWhitelist' -v`
Expected: FAIL because `CheckStructure` does not yet reject those top-level packages.

### Task 2: Implement the whitelist rule

**Files:**
- Modify: `rules/structure.go`

- [ ] **Step 1: Write minimal implementation**

Add a new structure rule that:
- inspects the immediate children under `internal/`
- allows only `domain`, `orchestration`, and `pkg`
- reports a dedicated structure violation for any other top-level directory or `.go` file

- [ ] **Step 2: Run focused tests to verify they pass**

Run: `go test ./rules ./... -run 'TestCheckStructure|TestIntegration_InternalTopLevelWhitelist' -v`
Expected: PASS

### Task 3: Align docs and complete verification

**Files:**
- Modify: `README.md`
- Create: `claude_history/2026-03-24-internal-top-level-whitelist.md`

- [ ] **Step 1: Update README**

Document that `internal/` top-level packages are restricted to `domain`, `orchestration`, and `pkg`, and add the new structure rule to the rules table.

- [ ] **Step 2: Run full verification**

Run:
- `make lint`
- `go test ./...`

Expected: both commands pass.

- [ ] **Step 3: Commit**

```bash
git add rules/structure.go rules/structure_test.go integration_test.go README.md claude_history/2026-03-24-internal-top-level-whitelist.md docs/superpowers/plans/2026-03-24-internal-top-level-whitelist.md
git commit -m "fix: restrict internal top-level packages" -m "History: claude_history/2026-03-24-internal-top-level-whitelist.md"
```
