# Domain Import Whitelist Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Tighten domain import permissions to `orchestration` and `cmd` only, remove router/bootstrap from the architecture contract, and clean up rule API quality.

**Architecture:** Enforce a narrow whitelist for domain-aware packages, keep `internal/pkg` domain-unaware, normalize violation paths, and stop requiring full type information in the loader. Tests go red first, then docs and implementation follow.

**Tech Stack:** Go, go/packages, filesystem checks, Go tests

---

### Task 1: Lock the whitelist policy and API quality with tests

**Files:**
- Modify: `rules/isolation_test.go`
- Modify: `rules/structure_test.go`
- Modify: `rules/naming_test.go`
- Modify: `integration_test.go`
- Modify: `analyzer/loader_test.go`
- Modify: `testdata/invalid/internal/router/router.go`
- Create: `testdata/invalid/internal/config/domain_alias.go`
- Create: `testdata/invalid/internal/bootstrap/bootstrap.go`

- [ ] **Step 1: Write failing tests**
- [ ] **Step 2: Run targeted tests and confirm they fail**

### Task 2: Update docs for the new architecture contract

**Files:**
- Modify: `README.md`

- [ ] **Step 1: Remove router/bootstrap as privileged internal layers**
- [ ] **Step 2: Document `cmd` as the composition root and `internal/pkg` as shared support**

### Task 3: Implement whitelist enforcement and API cleanup

**Files:**
- Modify: `rules/isolation.go`
- Modify: `rules/structure.go`
- Modify: `rules/naming.go`
- Modify: `analyzer/loader.go`

- [ ] **Step 1: Restrict domain imports to same-domain, orchestration, and cmd**
- [ ] **Step 2: Mark top-level router/bootstrap as legacy**
- [ ] **Step 3: Normalize naming violation paths to project-relative**
- [ ] **Step 4: Remove unneeded type loading from analyzer**

### Task 4: Verify and complete

**Files:**
- Modify: `claude_history/2026-03-23-domain-import-whitelist.md`

- [ ] **Step 1: Run `go test ./...`**
- [ ] **Step 2: Run `make lint`**
- [ ] **Step 3: Record the task and commit with History line**
