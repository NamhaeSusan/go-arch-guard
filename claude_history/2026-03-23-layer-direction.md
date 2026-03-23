# CheckLayerDirection implementation

## Task
Implement `CheckLayerDirection` — intra-domain layer direction rule for domain-centric architecture.

## Files changed
- `rules/layer.go` — new file: `CheckLayerDirection`, `identifySublayer`, `allowedLayerImports` table
- `rules/layer_test.go` — new file: tests against valid/invalid testdata
- `README.md` — added documentation for `CheckDomainIsolation` and `CheckLayerDirection`

## Verification
- `go test ./rules/ -run TestCheckLayerDirection -v -count=1` — PASS
- `go test ./... -count=1` — all packages PASS
- `make lint` — 0 issues
