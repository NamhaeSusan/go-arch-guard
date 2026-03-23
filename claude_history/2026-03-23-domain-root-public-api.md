# 2026-03-23 domain-root-public-api

## Summary

- Reframed the domain root package as the only external API surface.
- Added isolation coverage for non-domain internal packages that deep-import domains.
- Added a structure rule that each domain root may contain only `alias.go` as a non-test Go file.
- Updated README to match the approved public API model and canonical module path.

## Files Changed

- `README.md`
- `integration_test.go`
- `rules/isolation.go`
- `rules/isolation_test.go`
- `rules/structure.go`
- `rules/structure_test.go`
- `docs/superpowers/specs/2026-03-23-domain-root-public-api-design.md`
- `docs/superpowers/plans/2026-03-23-domain-root-public-api.md`
- `testdata/invalid/internal/config/config.go`

## Verification

- `go test ./...`
- `make lint`
