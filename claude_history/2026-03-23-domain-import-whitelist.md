# 2026-03-23 domain-import-whitelist

## Summary

- Tightened domain import permissions so only `internal/orchestration` and `cmd/...` may depend on domain root packages.
- Removed `router/bootstrap` from the privileged architecture contract and marked them as legacy top-level packages.
- Made `internal/pkg` explicitly domain-unaware by rejecting imports of domains and orchestration.
- Normalized naming violation file paths to project-relative paths.
- Relaxed `analyzer.Load` so syntax/import analysis does not require successful type checking.

## Files Changed

- `README.md`
- `analyzer/loader.go`
- `analyzer/loader_test.go`
- `example_test.go`
- `integration_test.go`
- `rules/helpers.go`
- `rules/isolation.go`
- `rules/isolation_test.go`
- `rules/naming.go`
- `rules/naming_test.go`
- `rules/structure.go`
- `rules/structure_test.go`
- `docs/superpowers/specs/2026-03-23-domain-import-whitelist-design.md`
- `docs/superpowers/plans/2026-03-23-domain-import-whitelist.md`
- `testdata/invalid/internal/router/router.go`
- `testdata/invalid/internal/bootstrap/bootstrap.go`
- `testdata/invalid/internal/config/domain_alias.go`
- `testdata/invalid/internal/pkg/orchestration.go`
- `testdata/load_type_error/go.mod`
- `testdata/load_type_error/internal/domain/user/app/service.go`

## Verification

- `go test ./...`
- `make lint`
