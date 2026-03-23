# 2026-03-22 Dependency Redesign

## Summary
Rewrote integration tests and docs to reflect domain-centric architecture rules.

## Changes
- `integration_test.go`: replaced empty file with domain-centric integration tests (Valid, Invalid, WarningMode)
- `example_test.go`: replaced empty file with Example using CheckDomainIsolation + CheckLayerDirection + CheckNaming
- `README.md`: removed old CheckDependency/CheckVerticalSlice/CheckVerticalSliceInternal references, updated Usage to domain-centric example

## Decisions
- Tests use testdata/valid and testdata/invalid fixtures with domain-centric module paths
- WarningMode test verifies severity propagation and that AssertNoViolations passes for warnings
