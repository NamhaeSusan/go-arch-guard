package testutil

import (
	"context"
	"database/sql"
)

// DB is an unclassified internal package — not in any known layer.
// BeginTx here should be a VIOLATION (tx.start-outside-allowed-layer)
// because unclassified packages are not in AllowedLayers.
var DB *sql.DB

func SetupTx(ctx context.Context) (*sql.Tx, error) {
	return DB.BeginTx(ctx, nil)
}
