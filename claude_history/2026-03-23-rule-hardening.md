# 2026-03-23 Rule Hardening

## Summary

- Added explicit protection against `domain -> orchestration` reverse dependencies.
- Rejected unsupported domain sublayers with a dedicated rule.
- Required `alias.go` to exist in every domain root.
- Backfilled missing regression coverage for `pkg -> domain`, missing domain model, and DTO placement.
- Synced README plus design/plan docs with the tightened rules.

## Files Changed

- `README.md`
- `integration_test.go`
- `rules/isolation.go`
- `rules/isolation_test.go`
- `rules/layer.go`
- `rules/layer_test.go`
- `rules/structure.go`
- `rules/structure_test.go`
- `docs/superpowers/specs/2026-03-23-rule-hardening-design.md`
- `docs/superpowers/plans/2026-03-23-rule-hardening.md`
- `testdata/invalid/internal/domain/ghost/alias.go`
- `testdata/invalid/internal/domain/noalias/core/model/item.go`
- `testdata/invalid/internal/domain/payment/alias.go`
- `testdata/invalid/internal/domain/payment/app/service.go`
- `testdata/invalid/internal/domain/payment/core/model/payment.go`
- `testdata/invalid/internal/domain/payment/policy/rule.go`
- `testdata/invalid/internal/domain/user/core/model/user_dto.go`
- `testdata/invalid/internal/pkg/domain_alias.go`

## Verification

- `go test ./rules -run 'TestCheckDomainIsolation|TestCheckLayerDirection|TestCheckStructure' -v`
- `go test ./... -run 'TestIntegration_Invalid|TestIntegration_Valid' -v`
- `go test ./...`
- `make lint`
