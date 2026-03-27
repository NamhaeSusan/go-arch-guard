# Warning output to stderr

## Summary
Changed `report.AssertNoViolations` to print Warning-level violations to
stderr instead of using `t.Log`. This makes warnings always visible in
`go test ./...` output without requiring the `-v` flag.

## Files Changed
- `report/report.go` — Warnings go to stderr, errors stay in t.Log/t.Errorf

## Verification
- `make lint` — 0 issues
- `go test ./...` — all pass
