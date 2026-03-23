# Rule Hardening Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Close the remaining enforcement gaps around domain boundaries, sublayer definition, and domain root public API requirements.

**Architecture:** Keep the current rule split by concern. Extend isolation for outer-layer reverse dependencies, extend layer checks with a closed-set sublayer contract, and extend structure checks with an explicit `alias.go` presence requirement. Add regression tests before implementation and update README to document the new rule IDs and guarantees.

**Tech Stack:** Go, `go/packages`, filesystem structure checks, Go tests

---

### Task 1: Lock the new behavior with failing tests

**Files:**
- Modify: `rules/isolation_test.go`
- Modify: `rules/layer_test.go`
- Modify: `rules/structure_test.go`
- Modify: `integration_test.go`
- Modify: `testdata/invalid/internal/domain/order/app/service.go`
- Create: `testdata/invalid/internal/domain/payment/core/model/payment.go`
- Create: `testdata/invalid/internal/domain/noalias/core/model/model.go`
- Create: `testdata/invalid/internal/domain/user/app/user_dto.go`
- Create: `testdata/invalid/internal/pkg/domain_alias.go`

- [ ] **Step 1: Add isolation tests for domain importing orchestration and pkg importing domain**
- [ ] **Step 2: Add layer tests for unknown sublayer rejection**
- [ ] **Step 3: Add structure tests for missing `alias.go`, missing model, and DTO placement**
- [ ] **Step 4: Run targeted tests to verify they fail**

Run: `go test ./rules -run 'TestCheckDomainIsolation|TestCheckLayerDirection|TestCheckStructure' -v`
Expected: FAIL with the new assertions.

### Task 2: Implement the minimal rule changes

**Files:**
- Modify: `rules/isolation.go`
- Modify: `rules/layer.go`
- Modify: `rules/structure.go`

- [ ] **Step 1: Add `isolation.domain-imports-orchestration`**
- [ ] **Step 2: Add `layer.unknown-sublayer` using a closed known-sublayer set**
- [ ] **Step 3: Add `structure.domain-root-alias-required`**
- [ ] **Step 4: Run targeted tests to verify they pass**

Run: `go test ./rules -run 'TestCheckDomainIsolation|TestCheckLayerDirection|TestCheckStructure' -v`
Expected: PASS

### Task 3: Sync docs and integration coverage

**Files:**
- Modify: `README.md`
- Modify: `integration_test.go`

- [ ] **Step 1: Document the new rule IDs and strengthened guarantees**
- [ ] **Step 2: Ensure integration coverage reflects the added invalid fixtures**
- [ ] **Step 3: Run focused integration tests**

Run: `go test ./... -run 'TestIntegration_Invalid|TestIntegration_Valid' -v`
Expected: PASS

### Task 4: Verify, review, and finish

**Files:**
- Create: `claude_history/2026-03-23-rule-hardening.md`

- [ ] **Step 1: Run full verification**

Run:
- `go test ./...`
- `make lint`

- [ ] **Step 2: Run code review**
- [ ] **Step 3: Record the work in `claude_history/2026-03-23-rule-hardening.md`**
- [ ] **Step 4: Commit with required `History:` footer**
