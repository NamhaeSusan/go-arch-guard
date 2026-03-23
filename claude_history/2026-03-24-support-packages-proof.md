## Summary

- Added an integration test proving neutral support packages outside `internal/domain`, `internal/orchestration`, and `internal/pkg` are currently allowed.
- The test creates `internal/config`, `internal/platform`, `internal/system`, and `internal/foundation` in a temp module and verifies all current rule checks pass.

## Files Changed

- `integration_test.go`
- `claude_history/2026-03-24-support-packages-proof.md`

## Verification

- `go test ./... -run TestIntegration_AllowsNeutralSupportPackagesOutsideCoreZones -v`
- `make lint`
- `go test ./...`
