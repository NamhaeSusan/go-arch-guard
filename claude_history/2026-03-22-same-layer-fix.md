# Fix same-layer sub-package import false positives

## Summary
Fixed dependency rule to allow same-layer sub-package imports (handler→handler, infra→infra).
Domain cross-domain isolation remains enforced.

## Files changed
- `rules/dependency.go` — added same-layer skip before `isAllowed` check
- `testdata/valid/internal/handler/dto/user_dto.go` — new: handler sub-package for testing
- `testdata/valid/internal/handler/http/user_handler.go` — added handler→handler/dto import
- `testdata/valid/internal/infra/db/db.go` — new: infra sub-package for testing
- `testdata/valid/internal/infra/postgres/user_repo.go` — added infra→infra/db import

## Verification
- `go test ./...` — all pass
- `go vet ./...` — clean
- Valid project: handler→handler/dto and infra→infra/db imports are NOT flagged
- Invalid project: handler→infra, domain→app, cross-domain imports still flagged correctly
