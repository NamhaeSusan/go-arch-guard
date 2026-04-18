package main

import (
	"context"
	"database/sql"
)

// main is the composition root. BeginTx here is a VIOLATION
// (tx.start-outside-allowed-layer) because cmd/ is not in AllowedLayers.
func main() {
	var db *sql.DB
	_, _ = db.BeginTx(context.Background(), nil)
}
