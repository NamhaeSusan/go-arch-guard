# 2026-03-23 Architecture Hardening

## Summary

- Added an explicit whitelist around `internal/orchestration` so only `cmd/...` and orchestration itself may depend on it.
- Blocked `core`, `core/model`, `core/repo`, `core/svc`, and `event` from importing `internal/pkg/...`.
- Kept `event` as a first-class sublayer and documented its allowed dependency direction.
- Normalized exclude handling around project-relative paths while preserving module-qualified compatibility.
- Hardened structure checks for recursive banned/legacy directories, alias package validation, and non-empty `core/model/`.
- Synced README, design docs, plan docs, tests, and fixtures with the tightened rules.

## Files Changed

- `README.md`
- `integration_test.go`
- `rules/helpers.go`
- `rules/isolation.go`
- `rules/isolation_test.go`
- `rules/layer.go`
- `rules/layer_test.go`
- `rules/rule.go`
- `rules/structure.go`
- `rules/structure_test.go`
- `docs/superpowers/specs/2026-03-23-architecture-hardening-design.md`
- `docs/superpowers/plans/2026-03-23-architecture-hardening.md`
- `testdata/invalid/internal/config/orchestration.go`
- `testdata/invalid/internal/domain/payment/core/model/pkg_leak.go`
- `testdata/invalid/internal/domain/payment/event/events.go`
- `testdata/invalid/internal/domain/payment/handler/http/event_handler.go`
- `testdata/invalid/internal/platform/bootstrap/bootstrap.go`
- `testdata/invalid/internal/platform/common/common.go`
- `testdata/invalid/internal/platform/handler/handler.go`
- `testdata/valid/internal/domain/order/app/service.go`
- `testdata/valid/internal/domain/order/event/events.go`

## Verification

- `go test ./rules -run 'TestCheckDomainIsolation|TestCheckLayerDirection|TestCheckStructure' -v`
- `go test ./...`
- `make lint`
