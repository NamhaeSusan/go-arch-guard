# README Structure Alignment Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Align the structure rule behavior and README wording so the documented constraints match actual enforcement.

**Architecture:** Tighten `CheckStructure` so `middleware/` is only allowed directly under `internal/pkg/`, then keep the `core/model/` requirement intentionally shallow and document that `core/model/` must contain a direct non-test Go file. Add focused regression tests for the new structure behavior and for the documented shallow-model behavior.

**Tech Stack:** Go, `testing`, existing structure-rule test suite

---

### Task 1: Restrict middleware placement to internal/pkg only

**Files:**
- Modify: `rules/structure_test.go`
- Modify: `rules/structure.go`

- [ ] **Step 1: Write the failing test**
- [ ] **Step 2: Run the structure test to verify it fails**
- [ ] **Step 3: Tighten middleware placement logic**
- [ ] **Step 4: Run the structure test to verify it passes**

### Task 2: Document shallow core/model requirement

**Files:**
- Modify: `rules/structure_test.go`
- Modify: `README.md`

- [ ] **Step 1: Add/confirm regression coverage for nested-only core/model**
- [ ] **Step 2: Run the structure test to verify current behavior**
- [ ] **Step 3: Update README wording to describe direct-file requirement**
- [ ] **Step 4: Run targeted and full verification**

### Task 3: Finalize

**Files:**
- Modify: `claude_history/2026-03-23-readme-structure-alignment.md`

- [ ] **Step 1: Run `go test ./...`**
- [ ] **Step 2: Run `make lint`**
- [ ] **Step 3: Write history record**
- [ ] **Step 4: Commit with required History line**
