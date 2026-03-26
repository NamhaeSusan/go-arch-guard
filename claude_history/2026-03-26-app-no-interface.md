# app-no-interface rule

## Summary

Added `naming.app-no-interface` rule to ban interface definitions in app/ packages.

## Motivation

After `handler-no-interface` was added, cross-domain interfaces were moved from handler/ to app/ to dodge the rule. This closes that loophole — interfaces belong in core/repo/ or core/svc/, not app/.

## Files Changed

- `rules/naming.go` — added `checkAppNoInterface`, `isAppPackage`
- `rules/naming_test.go` — added test case
- `testdata/invalid/.../app/service.go` — added interface fixture
- `README.md` — documented rule

## Verification

- `go test ./...` — all PASS
- `make lint` — 0 issues
