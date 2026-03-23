# Domain Root Public API Design

**Date:** 2026-03-23

## Goal

Make the domain root package the only external API surface for a domain while keeping internal sub-packages hidden behind re-exports in `alias.go`.

## Decisions

1. `handler-only domain` is removed from v1.
2. Canonical module path is `github.com/NamhaeSusan/go-arch-guard`.
3. External packages may import only `internal/domain/<name>` for a domain.
4. External deep imports like `internal/domain/<name>/handler/http` or `internal/domain/<name>/core/svc` are violations.
5. `alias.go` is not a layer rule target. It is the public surface declaration file for the domain root package.
6. `alias.go` may selectively import and re-export internal symbols such as `app.Public`, `handler/http.Handler`, and constructors.
7. The domain root package should contain only `alias.go` as a non-test Go file so the public surface stays in one place.

## Implications

- Router and other external packages must assemble domain behavior through the root package only.
- Deep imports stay forbidden even when the root package re-exports selected symbols.
- The previous rule "`alias.go` can only import `app`" is removed.
- Architecture enforcement moves from "what alias may import" to "what external packages may import".

## Example

```go
// internal/domain/order/alias.go
package order

import (
    orderapp "mymodule/internal/domain/order/app"
    orderhttp "mymodule/internal/domain/order/handler/http"
)

type Public = orderapp.Public

var NewPublic = orderapp.NewPublic

type Handler = orderhttp.Handler

var NewHandler = orderhttp.NewHandler
```

External packages use only:

```go
import "mymodule/internal/domain/order"
```

## Enforcement Changes

- Update README to describe the root package as the public API.
- Add a structure rule that the domain root package may contain only `alias.go` as a non-test Go file.
- Tighten isolation so any external deep import into a domain is rejected unless the importer is inside the same domain.
