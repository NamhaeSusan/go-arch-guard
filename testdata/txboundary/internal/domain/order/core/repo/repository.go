package repo

import (
    "context"
    "database/sql"

    "github.com/kimtaeyun/testproject-txboundary/internal/domain/order/core/model"
)

type Repo struct {
    db *sql.DB
}

// BeginInRepo starts a tx in repo — VIOLATION (tx.start-outside-allowed-layer).
func (r *Repo) BeginInRepo(ctx context.Context) error {
    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    return tx.Commit()
}

// Save accepts *sql.Tx — VIOLATION (tx.type-in-signature).
func (r *Repo) Save(tx *sql.Tx, o model.Order) error {
    _ = tx
    _ = o
    return nil
}
