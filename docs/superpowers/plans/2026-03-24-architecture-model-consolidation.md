# Architecture Model Consolidation

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix structure rule false positives on empty directories, consolidate duplicated architecture constants into a single source of truth, and remove the `policy` drift between naming and layer engines.

**Architecture:** Extract a shared `rules/arch.go` file that declares all architecture model constants (known sublayers, allowed top-level packages, banned names, layer-dir names for naming checks). Each rule file imports from `arch.go` instead of maintaining its own copy. Structure checks gain a `hasNonTestGoFiles` guard so empty directories are not treated as Go packages.

**Tech Stack:** Go, `golang.org/x/tools/go/packages`

---

### Task 1: Fix structure false positive on empty directories

The most critical bug. `checkPackageNames` and `checkMiddlewarePlacement` walk all directories and flag banned/legacy/middleware names even when the directory contains zero Go files. This produces false positives that violate the project's "low false-positive" principle.

**Files:**
- Modify: `rules/structure.go:192-246` (`checkPackageNames`), `rules/structure.go:249-275` (`checkMiddlewarePlacement`)
- Test: `rules/structure_test.go`

- [ ] **Step 1: Write the failing test — empty directory should not trigger banned-package**

Add to `rules/structure_test.go`:

```go
t.Run("ignores empty directory with banned name", func(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "internal", "domain", "order", "alias.go"), "package order\n")
	writeFile(t, filepath.Join(root, "internal", "domain", "order", "core", "model", "order.go"), "package model\n")
	if err := os.MkdirAll(filepath.Join(root, "internal", "domain", "order", "shared"), 0o755); err != nil {
		t.Fatal(err)
	}

	violations := rules.CheckStructure(root)
	for _, v := range violations {
		if v.Rule == "structure.banned-package" && strings.Contains(v.File, "shared") {
			t.Errorf("empty directory should not trigger banned-package: %s", v.String())
		}
	}
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./rules/ -run TestCheckStructure/ignores_empty_directory_with_banned_name -v`
Expected: FAIL — `empty directory should not trigger banned-package`

- [ ] **Step 3: Write a second failing test — empty middleware directory**

Add to `rules/structure_test.go`:

```go
t.Run("ignores empty middleware directory", func(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "internal", "domain", "order", "alias.go"), "package order\n")
	writeFile(t, filepath.Join(root, "internal", "domain", "order", "core", "model", "order.go"), "package model\n")
	if err := os.MkdirAll(filepath.Join(root, "internal", "domain", "order", "middleware"), 0o755); err != nil {
		t.Fatal(err)
	}

	violations := rules.CheckStructure(root)
	for _, v := range violations {
		if v.Rule == "structure.middleware-placement" && strings.Contains(v.File, "middleware") {
			t.Errorf("empty middleware directory should not trigger violation: %s", v.String())
		}
	}
})
```

- [ ] **Step 4: Run test to verify it fails**

Run: `go test ./rules/ -run TestCheckStructure/ignores_empty_middleware_directory -v`
Expected: FAIL

- [ ] **Step 5: Implement the fix — add `hasNonTestGoFiles` guard to `checkPackageNames`**

In `rules/structure.go`, add a `dirHasGoFiles` helper that checks whether a directory (non-recursively) contains at least one non-test `.go` file. Then guard the banned/legacy/misplaced checks:

```go
// dirHasGoFiles reports whether dir contains at least one non-test Go file
// (non-recursively). Directories without Go files are not Go packages.
func dirHasGoFiles(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go") {
			return true
		}
	}
	return false
}
```

In `checkPackageNames`, after the exclude check, add:

```go
if !dirHasGoFiles(path) {
	return nil
}
```

In `checkMiddlewarePlacement`, after the `d.Name() != "middleware"` check, add:

```go
if !dirHasGoFiles(path) {
	return nil
}
```

Note: `hasNonTestGoFiles` already exists but only checks for direct non-test Go files — it does the same thing. Rename it to be more general or reuse it. Looking at the existing code, `hasNonTestGoFiles(dir)` at line 370 does exactly what we need. Just reuse it in both walkers.

- [ ] **Step 6: Run all structure tests**

Run: `go test ./rules/ -run TestCheckStructure -v`
Expected: ALL PASS (including the 2 new tests and all existing tests)

- [ ] **Step 7: Commit**

```bash
git add rules/structure.go rules/structure_test.go
git commit -m "fix: skip empty directories in structure checks

Directories without non-test Go files are not Go packages.
checkPackageNames and checkMiddlewarePlacement now skip them,
preventing false positives like flagging internal/domain/order/shared/
when it has no Go files."
```

---

### Task 2: Extract shared architecture constants to `rules/arch.go`

Address the drift where sublayer/layer knowledge is duplicated across `layer.go`, `naming.go`, and `structure.go`. The `policy` entry in naming's `layerDirs` but absent from layer's `knownSublayers` is the concrete symptom.

**Files:**
- Create: `rules/arch.go`
- Modify: `rules/layer.go:11-41` (remove local maps, import from arch.go)
- Modify: `rules/naming.go:244-250` (remove `layerDirs`, import from arch.go)
- Modify: `rules/structure.go:12-19` (remove local vars, import from arch.go)
- Test: `rules/layer_test.go`, `rules/naming_test.go`, `rules/structure_test.go`

- [ ] **Step 1: Write test — `layerDirs` and `knownSublayers` must be consistent**

Create a consistency test in `rules/layer_test.go` (or a new `rules/arch_test.go`):

```go
// rules/arch_test.go
package rules

import "testing"

func TestArchModelConsistency(t *testing.T) {
	// Every domain sublayer in KnownSublayers must be a key in AllowedLayerImports
	for _, sl := range KnownDomainSublayers {
		if _, ok := AllowedLayerImports[sl]; !ok {
			t.Errorf("KnownDomainSublayers contains %q but AllowedLayerImports does not", sl)
		}
	}

	// LayerDirNames must be a superset of KnownDomainSublayers leaf names
	for _, sl := range KnownDomainSublayers {
		// Extract leaf: "core/model" → "model", "handler" → "handler"
		leaf := sl
		if idx := len(sl) - 1; idx >= 0 {
			for i := len(sl) - 1; i >= 0; i-- {
				if sl[i] == '/' {
					leaf = sl[i+1:]
					break
				}
			}
		}
		if !LayerDirNames[leaf] {
			t.Errorf("KnownDomainSublayers leaf %q (from %q) missing in LayerDirNames", leaf, sl)
		}
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./rules/ -run TestArchModelConsistency -v`
Expected: FAIL (compilation error — exported names don't exist yet)

- [ ] **Step 3: Create `rules/arch.go` with consolidated constants**

```go
// Package rules – arch.go is the single source of truth for architecture
// model constants shared across rule implementations.
package rules

// KnownDomainSublayers is the exhaustive set of sublayer paths recognised
// inside a domain (e.g. "handler", "core/model"). Everything else is
// flagged as unknown by CheckLayerDirection.
var KnownDomainSublayers = []string{
	"handler",
	"app",
	"core",
	"core/model",
	"core/repo",
	"core/svc",
	"event",
	"infra",
}

// AllowedLayerImports defines the intra-domain import graph.
// Key = source sublayer, value = allowed target sublayers.
var AllowedLayerImports = map[string][]string{
	"handler":    {"app"},
	"app":        {"core/model", "core/repo", "core/svc", "event"},
	"core":       {"core/model"},
	"core/model": {},
	"core/repo":  {"core/model"},
	"core/svc":   {"core/model"},
	"event":      {"core/model"},
	"infra":      {"core/repo", "core/model", "event"},
}

// PkgRestrictedSublayers are sublayers that must not import internal/pkg.
var PkgRestrictedSublayers = map[string]bool{
	"core":       true,
	"core/model": true,
	"core/repo":  true,
	"core/svc":   true,
	"event":      true,
}

// AllowedInternalTopLevel lists the only packages permitted directly under
// internal/.
var AllowedInternalTopLevel = map[string]bool{
	"domain":        true,
	"orchestration": true,
	"pkg":           true,
}

// BannedPackageNames are package names rejected anywhere under internal/.
var BannedPackageNames = []string{
	"util", "common", "misc", "helper", "shared", "services",
}

// LegacyPackageNames are package names that trigger a migration warning.
var LegacyPackageNames = []string{"router", "bootstrap"}

// LayerDirNames identifies directory names that are "layer-like" for the
// naming.no-layer-suffix check. Derived from KnownDomainSublayers plus
// common synonyms (controller, entity, persistence, store, service, etc.).
var LayerDirNames = map[string]bool{
	// from KnownDomainSublayers
	"handler": true, "app": true, "core": true,
	"model": true, "repo": true, "svc": true,
	"event": true, "infra": true,
	// common synonyms
	"service": true, "controller": true,
	"entity": true, "store": true, "persistence": true,
	"domain": true,
}
```

Note: `policy` is **removed** from `LayerDirNames` because it is not a known sublayer. If it were a legitimate sublayer, it would need to be in `KnownDomainSublayers` first.

- [ ] **Step 4: Update `layer.go` — replace local maps with arch.go references**

Remove the local `allowedLayerImports`, `knownSublayers`, and `pkgRestrictedSublayers` maps. Replace references:

- `allowedLayerImports` → `AllowedLayerImports`
- `knownSublayers[x]` → `isKnownSublayer(x)` (helper using `slices.Contains(KnownDomainSublayers, x)`)
- `pkgRestrictedSublayers` → `PkgRestrictedSublayers`
- `knownSublayerList()` → `KnownDomainSublayers` directly

- [ ] **Step 5: Update `naming.go` — replace `layerDirs` with `LayerDirNames`**

Remove the local `layerDirs` map. Replace `layerDirs[dir]` with `LayerDirNames[dir]`.

- [ ] **Step 6: Update `structure.go` — replace local vars with arch.go references**

- Remove `bannedPackageNames`, `legacyPackageNames`, `allowedInternalTopLevelPackages`
- Replace with `BannedPackageNames`, `LegacyPackageNames`, `AllowedInternalTopLevel`

- [ ] **Step 7: Run the consistency test**

Run: `go test ./rules/ -run TestArchModelConsistency -v`
Expected: PASS

- [ ] **Step 8: Run all tests**

Run: `go test ./... -v`
Expected: ALL PASS

- [ ] **Step 9: Run lint**

Run: `make lint`
Expected: 0 issues

- [ ] **Step 10: Commit**

```bash
git add rules/arch.go rules/arch_test.go rules/layer.go rules/naming.go rules/structure.go
git commit -m "refactor: consolidate architecture constants into rules/arch.go

Single source of truth for sublayers, allowed imports, banned names,
and layer-dir names. Removes policy from LayerDirNames (it was never
a known sublayer). Adds TestArchModelConsistency to prevent future drift."
```

---

### Task 3: Update README sublayer table and documentation

After the code changes, ensure README stays in sync.

**Files:**
- Modify: `README.md`
- Create: `claude_history/2026-03-24-arch-model-consolidation.md`

- [ ] **Step 1: Verify README sublayer list matches `KnownDomainSublayers`**

Check that the "Domain Layers" section (lines 153-163) lists exactly the sublayers from `arch.go`. Currently it lists: handler, app, core, core/model, core/repo, core/svc, event, infra — this matches, so no change needed.

- [ ] **Step 2: Verify import matrix matches `AllowedLayerImports`**

Check the "Layer Direction" allowed-imports table (lines 230-239). Verify each row matches `AllowedLayerImports`. Currently matches — no change needed.

- [ ] **Step 3: Write claude_history entry**

Create `claude_history/2026-03-24-arch-model-consolidation.md` with a summary of all changes made.

- [ ] **Step 4: Run final full test suite**

Run: `go test ./... && make lint`
Expected: ALL PASS, 0 lint issues

- [ ] **Step 5: Commit docs**

```bash
git add README.md claude_history/2026-03-24-arch-model-consolidation.md
git commit -m "docs: add work log for architecture model consolidation"
```
