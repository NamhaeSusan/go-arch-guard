# Domain Root Public API Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Align the rule engine and docs with the approved model where each domain exposes only its root package and `alias.go` defines the public API surface.

**Architecture:** Remove the old "`alias.go` imports `app` only" assumption. Enforce external root-only imports and a single non-test root file per domain. Update README and tests first, then implement the minimal rule changes.

**Tech Stack:** Go, go/packages, filesystem checks, Go tests

---

### Task 1: Lock the public API model in docs and tests

**Files:**
- Modify: `README.md`
- Modify: `rules/isolation_test.go`
- Modify: `rules/structure_test.go`
- Modify: `integration_test.go`

- [ ] **Step 1: Write the failing tests**

Add assertions for:
- external deep imports into a domain are rejected
- extra non-test files in a domain root are rejected
- integration test runs structure checks

- [ ] **Step 2: Run targeted tests to verify they fail**

Run: `go test ./rules ./... -run 'TestCheckDomainIsolation|TestCheckStructure|TestIntegration_Invalid|TestIntegration_Valid' -v`

Expected: FAIL because the new rule is not implemented yet.

- [ ] **Step 3: Update README to the approved design**

Change docs so they describe:
- domain root package as the external API
- `alias.go` as the public surface declaration file
- no handler-only domain in v1
- canonical `github.com/NamhaeSusan/go-arch-guard` module path

### Task 2: Implement root-only import enforcement

**Files:**
- Modify: `rules/isolation.go`

- [ ] **Step 1: Implement the minimal isolation change**

Reject domain deep imports from any package outside the same domain, while preserving:
- same-domain internal imports
- root-package imports for router/orchestration/cmd/bootstrap
- `pkg/` importing domains remains forbidden

- [ ] **Step 2: Run targeted isolation tests**

Run: `go test ./rules -run TestCheckDomainIsolation -v`

Expected: PASS

### Task 3: Implement the single-file root surface rule

**Files:**
- Modify: `rules/structure.go`

- [ ] **Step 1: Implement the minimal structure change**

Each `internal/domain/<name>/` directory may contain only `alias.go` as a non-test Go file.

- [ ] **Step 2: Run targeted structure tests**

Run: `go test ./rules -run TestCheckStructure -v`

Expected: PASS

### Task 4: Full verification and completion

**Files:**
- Modify: `claude_history/2026-03-23-domain-root-public-api.md`

- [ ] **Step 1: Run full verification**

Run:
- `go test ./...`
- `make lint`

- [ ] **Step 2: Record the task**

Write a concise history note with changed files and verification results.

- [ ] **Step 3: Commit**

Use a commit that follows project rules and includes a `History:` line.
