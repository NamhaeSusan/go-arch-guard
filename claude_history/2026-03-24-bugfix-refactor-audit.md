# 2026-03-24 Architecture Audit: Bug Fix & Refactoring

## Task
Deep analysis of architecture, rules, and code for critical gaps and refactoring opportunities.

## Changes

### P0 Bug Fix: `naming.repo-file-interface` index panic
- **File**: `rules/naming.go` (`checkRepoFileInterface`)
- **Problem**: Used `pkg.Syntax[i]` indexed by `pkg.GoFiles` loop counter. `GoFiles` and `Syntax` (which maps to `CompiledGoFiles`) are not guaranteed to correspond 1:1 (cgo, generated files).
- **Fix**: Iterate `pkg.Syntax` directly, deriving filenames from AST positions. Eliminates index mismatch entirely.

### P1 API Clarity: `analyzer.Load` partial error handling
- **Files**: `analyzer/loader.go`, `README.md`
- **Problem**: `Load` returns `(validPkgs, error)` on partial failures, but Quick Start example used `t.Fatal(err)` — discarding valid packages on any partial error.
- **Fix**: Added godoc explaining partial-error semantics; updated Quick Start to `t.Log(err)` + `len(pkgs) == 0` fatal guard.

### P2 Refactor: `filepath.Walk` → `filepath.WalkDir`
- **File**: `rules/structure.go`
- **Change**: Replaced all 3 `filepath.Walk` calls with `filepath.WalkDir` (`fs.DirEntry` avoids unnecessary `os.Stat` per entry).

### Lint fixes
- `rules/helpers.go`: `HasPrefix+TrimPrefix` → `CutPrefix`
- `rules/naming.go`: `HasSuffix+TrimSuffix` → `CutSuffix`

## Verification
- `go test ./...` — all pass
- `make lint` — 0 issues
