# 2026-03-23 Add repo-file-interface and no-layer-suffix naming rules

## Summary
Added two new naming rules to `rules/naming.go`:
- `naming.repo-file-interface`: Files in `repo/` directories must contain an exported interface matching the filename (snake_case -> PascalCase conversion).
- `naming.no-layer-suffix`: Files in recognized layer directories must not have redundant layer suffixes (e.g., `_svc`, `_repo`, `_handler`).

## Files Changed
- `rules/naming.go` — Added `checkRepoFileInterface()`, `checkNoLayerSuffix()`, and helpers
- `rules/naming_test.go` — Added 2 new test cases
- `testdata/valid/internal/domain/user/repo/user.go` — New: valid repo file with matching interface
- `testdata/valid/internal/app/user_service.go` → `user.go` — Renamed to avoid layer suffix violation
- `testdata/invalid/internal/domain/user/repo/user.go` — New: repo file missing interface
- `testdata/invalid/internal/app/order_svc.go` — New: file with layer suffix in layer dir
- `testdata/vertical-valid/internal/order/repo/order_repo.go` → `order.go` — Renamed, interface changed from Repository to Order
- `testdata/vertical-valid/internal/order/infra/persistence/order_store.go` → `order.go` — Renamed
- `README.md` — Added 2 new rules to Naming section

## Verification
- `go test ./... -count=1 -v` — all tests pass
- `go vet ./...` — no issues
- `make lint` — 0 issues
