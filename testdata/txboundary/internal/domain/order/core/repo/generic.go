package repo

import (
	"context"
	"database/sql"
)

// BeginGeneric is a generic function used to exercise the generic-unwrap
// path in resolveCalleeID. Callers invoke it as BeginGeneric[T](...), which
// arrives in the AST as *ast.IndexExpr wrapping the identifier.
// The body does not call BeginTx so it doesn't contribute to BeginTx-
// forbidding tests; only the call site BeginGeneric[T](...) is of interest.
func BeginGeneric[T any](ctx context.Context, db *sql.DB) (*sql.Tx, error) {
	_, _ = ctx, db
	return nil, nil
}

// CallGenericStart invokes BeginGeneric with an explicit type parameter,
// producing an *ast.IndexExpr at the call site.
func CallGenericStart(ctx context.Context, db *sql.DB) (*sql.Tx, error) {
	return BeginGeneric[string](ctx, db)
}
