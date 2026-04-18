package generic

import (
	"context"
	"database/sql"
)

// DB is a package-level db used by call sites.
var DB *sql.DB

// BeginGeneric is a generic function that calls BeginTx.
// When called with explicit type param (BeginGeneric[string](...)), the
// AST Fun node is an *ast.IndexExpr — currently skipped by resolveCalleeID.
// VIOLATION: generic package is not in AllowedLayers.
func BeginGeneric[T any](ctx context.Context, db *sql.DB) (*sql.Tx, error) {
	return db.BeginTx(ctx, nil)
}

// Runner is a generic type whose method also calls BeginTx.
type Runner[T any] struct {
	db *sql.DB
}

// Start calls BeginTx — VIOLATION (generic package not allowed).
func (r *Runner[T]) Start(ctx context.Context) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, nil)
}

// UseExplicitTypeParam exercises explicit generic instantiation syntax so
// the AST contains *ast.IndexExpr as the CallExpr.Fun.
// e.g. BeginGeneric[string](ctx, db) — Fun is IndexExpr{X: Ident("BeginGeneric"), Index: Ident("string")}
func UseExplicitTypeParam(ctx context.Context) {
	_, _ = BeginGeneric[string](ctx, DB)
}
