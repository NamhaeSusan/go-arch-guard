# Vertical Slice Rules

## Date
2026-03-22

## Summary
Implemented vertical slice architecture enforcement rules for go-arch-guard.

## Changes

### Testdata (`testdata/vertical-valid/`, `testdata/vertical-invalid/`)
- Created valid vertical slice project structure with proper domain isolation
- Created invalid project with cross-domain and internal-layer violations

### Rules (`rules/vertical.go`)
- `CheckVerticalSlice`: enforces cross-domain isolation under `internal/`
  - Same-domain and `shared/` imports allowed
  - `shared/` importing a domain is a violation
  - `app/usecase/` cross-domain imports to root or `port/` are allowed
  - All other cross-domain imports are violations
- `CheckVerticalSliceInternal`: enforces intra-domain layer direction
  - Defines allowed import targets per sublayer (handler, app, domain, policy, infra, model, repo, event, port)
  - Skips cross-domain and shared imports (handled by CheckVerticalSlice)

### Integration tests (`integration_test.go`)
- `TestIntegration_VerticalValid`: verifies no violations for valid vertical slice project
- `TestIntegration_VerticalInvalid`: verifies violations detected for invalid project

### Documentation (`README.md`)
- Added CheckVerticalSlice and CheckVerticalSliceInternal sections with usage examples and layer tables
