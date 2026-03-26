# Skill Version Bump Action Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Automatically bump the Claude Code plugin version when `SKILL.md` changes.

**Architecture:** A Python script in `scripts/` owns version parsing and rollover. A dedicated GitHub Actions workflow triggers on skill-file changes to `main`, runs the script against `plugins/go-arch-guard/.claude-plugin/plugin.json`, and commits the changed version back to the branch. CI runs the Python unit tests so the bump logic stays verified.

**Tech Stack:** Python 3 stdlib (`json`, `pathlib`, `unittest`), GitHub Actions

**Spec:** `docs/superpowers/specs/2026-03-26-skill-version-bump-action-design.md`

---

## File Structure

| Action | File | Responsibility |
|--------|------|----------------|
| Create | `scripts/bump_plugin_version.py` | Parse and bump plugin version with rollover rules |
| Create | `scripts/test_bump_plugin_version.py` | Unit tests for bump logic and file rewrite |
| Create | `.github/workflows/plugin-version-bump.yml` | Run bump on skill changes and push updated manifest |
| Modify | `.github/workflows/ci.yml` | Run Python unit tests |
| Modify | `plugins/go-arch-guard/.claude-plugin/plugin.json` | Set initial version to `0.0.1` |
| Modify | `README.md` | Document auto version bump behavior |
| Modify | `CLAUDE.md` | Document workflow expectation for skill changes |
| Create | `claude_history/2026-03-26-skill-version-bump-action.md` | Work log |

---

### Task 1: Write failing tests for version rules

**Files:**
- Create: `scripts/test_bump_plugin_version.py`

- [ ] **Step 1: Add a unit test for a normal patch bump**

Test `0.0.1 -> 0.0.2`.

- [ ] **Step 2: Add a unit test for patch rollover**

Test `0.0.99 -> 0.1.0`.

- [ ] **Step 3: Add a unit test for overflow rejection**

Test `0.99.99` raises an error.

- [ ] **Step 4: Add a unit test for `plugin.json` rewrite**

Write a temp `plugin.json`, run the file bump helper, and assert that `version` changes correctly.

- [ ] **Step 5: Run tests to verify RED**

Run: `python3 -m unittest scripts/test_bump_plugin_version.py -v`
Expected: FAIL because `scripts.bump_plugin_version` does not exist yet

---

### Task 2: Implement the bump script

**Files:**
- Create: `scripts/bump_plugin_version.py`
- Modify: `plugins/go-arch-guard/.claude-plugin/plugin.json`

- [ ] **Step 1: Implement version parsing and validation**

Support exactly three numeric components, each in `0..99`.

- [ ] **Step 2: Implement patch bump with rollover**

Apply patch bump, then rollover to minor when patch exceeds `99`.

- [ ] **Step 3: Implement JSON file update**

Read `plugin.json`, replace `version`, write formatted JSON with trailing newline.

- [ ] **Step 4: Set the repository starting version**

Change plugin version to `0.0.1`.

- [ ] **Step 5: Run tests to verify GREEN**

Run: `python3 -m unittest scripts/test_bump_plugin_version.py -v`
Expected: PASS

---

### Task 3: Add the GitHub Actions workflow

**Files:**
- Create: `.github/workflows/plugin-version-bump.yml`
- Modify: `.github/workflows/ci.yml`

- [ ] **Step 1: Add the bump workflow**

Trigger on `push` to `main` with skill-file path filters. Check out the repo, set up Python, run the bump script, and commit/push only when `plugin.json` changed.

- [ ] **Step 2: Add the Python unit test to CI**

Run the unit test file in `.github/workflows/ci.yml`.

- [ ] **Step 3: Validate workflow syntax by inspection and local test coverage**

Run the Python unit tests again after workflow changes.

---

### Task 4: Update docs and verify

**Files:**
- Modify: `README.md`
- Modify: `CLAUDE.md`
- Create: `claude_history/2026-03-26-skill-version-bump-action.md`

- [ ] **Step 1: Document the auto bump behavior**

Explain that skill changes on `main` automatically bump the plugin patch version with the repository rollover rule.

- [ ] **Step 2: Write the work log**

Record changed files and verification results.

- [ ] **Step 3: Run verification**

Run: `python3 -m unittest scripts/test_bump_plugin_version.py -v`
Expected: PASS

Run: `go test ./...`
Expected: PASS

Run: `make lint`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add scripts .github/workflows plugins/go-arch-guard/.claude-plugin/plugin.json README.md CLAUDE.md claude_history docs/superpowers
git commit -m "feat: automate plugin version bumps for skill updates"
```
