# Architecture Hardening Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Enforce the approved architecture hardening rules around orchestration ownership, inner-layer purity, recursive structure validation, and clearer exclude semantics.

**Architecture:** Keep the existing split between isolation, layer, naming, and structure checks. Add the new hardening as focused rules instead of introducing a config DSL or mode flag. Normalize shared helper behavior so rule output and exclusions stay consistent across packages and filesystem scans.

**Tech Stack:** Go, `go/packages`, filesystem checks, Go tests

---

### Task 1: Lock the new behavior with failing tests

**Files:**
- Modify: `rules/isolation_test.go`
- Modify: `rules/layer_test.go`
- Modify: `rules/structure_test.go`
- Modify: `integration_test.go`
- Modify: `testdata/invalid/internal/config/config.go`
- Modify: `testdata/invalid/internal/domain/order/app/service.go`
- Modify: `testdata/invalid/internal/domain/order/alias.go`
- Create: `testdata/invalid/internal/common/common.go`
- Create: `testdata/invalid/internal/domain/ghost/core/model/README.md`
- Create: `testdata/invalid/internal/domain/order/core/model/helper.go`
- Create: `testdata/invalid/internal/platform/bootstrap/bootstrap.go`

- [ ] **Step 1: Add isolation tests for unauthorized imports of `internal/orchestration`**
- [ ] **Step 2: Add layer tests for inner layers importing `internal/pkg` and the explicit `event` policy**
- [ ] **Step 3: Add structure tests for recursive banned/legacy directories and strengthened alias/model checks**
- [ ] **Step 4: Run targeted tests to verify they fail**

Run: `go test ./rules -run 'TestCheckDomainIsolation|TestCheckLayerDirection|TestCheckStructure' -v`
Expected: FAIL with the new assertions.

### Task 2: Implement the minimal rule changes

**Files:**
- Modify: `rules/isolation.go`
- Modify: `rules/layer.go`
- Modify: `rules/structure.go`
- Modify: `rules/rule.go`
- Modify: `rules/helpers.go`

- [ ] **Step 1: Add the orchestration import whitelist and violation**
- [ ] **Step 2: Add `layer.inner-imports-pkg` and explicit `event` imports**
- [ ] **Step 3: Normalize exclude-path handling helpers**
- [ ] **Step 4: Harden structure checks for recursive names and alias/model validation**
- [ ] **Step 5: Run targeted tests to verify they pass**

Run: `go test ./rules -run 'TestCheckDomainIsolation|TestCheckLayerDirection|TestCheckStructure' -v`
Expected: PASS

### Task 3: Sync docs and integration coverage

**Files:**
- Modify: `README.md`
- Modify: `integration_test.go`

- [ ] **Step 1: Document the new whitelist, inner-layer purity rule, and `event` policy**
- [ ] **Step 2: Update exclude examples to project-relative paths**
- [ ] **Step 3: Ensure integration tests surface the new rule IDs**
- [ ] **Step 4: Run focused integration coverage**

Run: `go test ./... -run 'TestIntegration_Invalid|TestIntegration_Valid|TestIntegration_WarningMode' -v`
Expected: PASS

### Task 4: Verify and finish

**Files:**
- Create: `claude_history/2026-03-23-architecture-hardening.md`

- [ ] **Step 1: Run full verification**

Run:
- `go test ./...`
- `make lint`

- [ ] **Step 2: Review final diff for documentation and rule consistency**
- [ ] **Step 3: Record the work in `claude_history/2026-03-23-architecture-hardening.md`**
- [ ] **Step 4: Commit with required `History:` footer**
