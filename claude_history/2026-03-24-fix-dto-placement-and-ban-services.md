# Fix DTO placement rule and ban services package

## Task
1. Fix `structure.dto-placement` rule contradiction — was blocking DTOs everywhere under `domain/` including `handler/` and `app/` where they legitimately belong
2. Add `services` to banned package names

## Changes
- `rules/structure.go`: Added `isDTOAllowedSublayer()` to skip `handler/` and `app/` sublayers in DTO placement check; added `"services"` to `bannedPackageNames`
- `rules/structure_test.go`: Added 4 tests (dto allowed in handler, dto allowed in app, dto still rejected in core/model, services detected as banned)
- `README.md`: Updated `structure.dto-placement` and `structure.banned-package` descriptions

## Verification
- All tests pass (`go test ./...`)
- Lint clean (`make lint`)
