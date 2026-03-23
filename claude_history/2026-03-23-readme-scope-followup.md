# 2026-03-23 readme-scope-followup

## Summary

- Updated `README.md` to describe `internal/orchestration` as the cross-domain saga/workflow layer.
- Relaxed README wording around domain-core purity so it reflects project-internal dependency-flow guarantees rather than absolute purity claims.
- Clarified that `CheckNaming` is an opinionated convention layer and can be omitted when teams only want boundary and structure guardrails.

## Files Changed

- `README.md`

## Verification

- `go test ./...`
- `make lint`
