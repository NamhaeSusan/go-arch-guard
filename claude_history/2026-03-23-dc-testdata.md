# Domain-Centric Testdata Creation

## Task
Create testdata for domain-centric architecture validation (valid + invalid).

## Files Created

### testdata/valid/ (module: testproject-dc)
- Full domain-centric structure with user/order domains, saga, and pkg
- Proper alias pattern (root package re-exports app types)
- Correct layer dependencies (handler->app, infra->core, svc->model only)
- Cross-domain communication via saga using root aliases

### testdata/invalid/ (module: testproject-dc-invalid)
- Same base structure as valid, plus 4 violation files:
- HARD: order/handler imports user/core/model (cross-domain leak)
- HARD: saga imports user/core/model directly (bypasses alias)
- HARD: order/alias_bad.go imports core/model (alias should only import app)
- SOFT: order/core/svc imports core/repo (svc should only depend on model)

## Verification
- `go build ./...` passes for both valid and invalid projects
