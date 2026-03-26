# handler-no-interface rule

## Summary

Replaced `naming.handler-no-exported-interface` with `naming.handler-no-interface` to ban ALL interface definitions (exported + unexported) in handler packages.

## Motivation

Handler packages were defining unexported interfaces (e.g., `contractCreator`, `auditService`, `adminOps`) to inject cross-domain dependencies directly. These should go through `app.Service` or `orchestration/` instead.

## Files Changed

- `rules/naming.go` — renamed `checkHandlerNoExportedInterface` → `checkHandlerNoInterface`, removed `IsExported()` filter
- `rules/naming_test.go` — updated test to check both exported and unexported interfaces
- `testdata/invalid/.../bad_service.go` — added unexported interface fixture
- `README.md` — updated rule table

## Verification

- `go test ./...` — all PASS
- `make lint` — 0 issues
