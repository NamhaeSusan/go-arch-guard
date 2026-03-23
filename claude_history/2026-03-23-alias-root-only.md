# 2026-03-23 alias-root-only

## Summary

- Tightened `structure.domain-model-required` so a domain model must live under `core/model/`.
- Stopped treating root-level `model.go` as a valid substitute for `core/model/`.
- Added regression coverage for the case where a domain root contains `alias.go` and `model.go`.
- Updated README wording to match the alias-only domain root rule.

## Files Changed

- `README.md`
- `rules/structure.go`
- `rules/structure_test.go`

## Verification

- `go test ./rules -run 'TestCheckStructure/root model file does not satisfy alias-only domain model requirement' -v`
- `go test ./...`
- `make lint`
