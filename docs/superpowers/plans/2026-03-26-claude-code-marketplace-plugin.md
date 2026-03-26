# Claude Code Marketplace Plugin Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a Claude Code marketplace/plugin wrapper around the existing skill without changing library behavior.

**Architecture:** The repository root will expose a `.claude-plugin/marketplace.json` marketplace manifest. A single plugin will live under `plugins/go-arch-guard/`, and the plugin will ship the existing `go-arch-guard` skill under `skills/go-arch-guard/SKILL.md`. README updates will document install and validation commands.

**Tech Stack:** JSON manifests, Markdown docs, Claude Code plugin validation, existing Go test/lint commands

**Spec:** `docs/superpowers/specs/2026-03-26-claude-code-marketplace-plugin-design.md`

---

## File Structure

| Action | File | Responsibility |
|--------|------|----------------|
| Create | `.claude-plugin/marketplace.json` | Marketplace metadata for this repository |
| Create | `plugins/go-arch-guard/.claude-plugin/plugin.json` | Plugin metadata for installation |
| Create | `plugins/go-arch-guard/skills/go-arch-guard/SKILL.md` | Plugin-packaged copy/link of the existing skill |
| Modify | `README.md` | Document Claude Code marketplace install and validation |
| Modify | `SKILL.md` | Mention Claude Code plugin packaging if needed |
| Create | `claude_history/2026-03-26-claude-code-marketplace-plugin.md` | Task log with verification evidence |

---

### Task 1: Add marketplace manifests

**Files:**
- Create: `.claude-plugin/marketplace.json`
- Create: `plugins/go-arch-guard/.claude-plugin/plugin.json`

- [ ] **Step 1: Write the marketplace manifest**

Add a repository-level marketplace manifest with one plugin entry pointing at `./plugins/go-arch-guard`.

- [ ] **Step 2: Write the plugin manifest**

Add plugin metadata with the plugin name `go-arch-guard`, a short description, and an initial version.

- [ ] **Step 3: Validate JSON syntax**

Run: `python -m json.tool .claude-plugin/marketplace.json`
Expected: formatted JSON output

Run: `python -m json.tool plugins/go-arch-guard/.claude-plugin/plugin.json`
Expected: formatted JSON output

---

### Task 2: Package the existing skill for Claude Code

**Files:**
- Create: `plugins/go-arch-guard/skills/go-arch-guard/SKILL.md`

- [ ] **Step 1: Reuse the existing skill content**

Package the current root `SKILL.md` as the plugin skill without changing the instruction body.

- [ ] **Step 2: Verify the packaged skill exists at the expected path**

Run: `test -f plugins/go-arch-guard/skills/go-arch-guard/SKILL.md`
Expected: exit code 0

---

### Task 3: Document installation and validation

**Files:**
- Modify: `README.md`
- Modify: `SKILL.md`

- [ ] **Step 1: Document marketplace installation**

Add commands for:

```text
/plugin marketplace add NamhaeSusan/go-arch-guard
/plugin install go-arch-guard@go-arch-guard-marketplace
```

- [ ] **Step 2: Document local validation**

Add `claude plugin validate .` as the validation command for the new packaging layer.

---

### Task 4: Verify and record

**Files:**
- Create: `claude_history/2026-03-26-claude-code-marketplace-plugin.md`

- [ ] **Step 1: Validate the plugin**

Run: `claude plugin validate .`
Expected: validation succeeds

- [ ] **Step 2: Run regression checks**

Run: `go test ./...`
Expected: PASS

Run: `make lint`
Expected: no issues

- [ ] **Step 3: Write the work log**

Record changed files and verification results in the task history file.

- [ ] **Step 4: Commit**

```bash
git add .claude-plugin plugins README.md SKILL.md claude_history docs/superpowers
git commit -m "feat: add Claude Code marketplace plugin wrapper"
```
