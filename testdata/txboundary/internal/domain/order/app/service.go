package app

import (
	"context"
	"database/sql"

	"github.com/kimtaeyun/testproject-txboundary/internal/domain/order/core/model"
)

type Service struct {
	db *sql.DB
}

// Place starts a transaction — legal because app is an allowed layer.
func (s *Service) Place(ctx context.Context, o model.Order) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// Helper takes *sql.Tx but still inside app — legal.
func (s *Service) apply(tx *sql.Tx, o model.Order) error {
	_ = tx
	_ = o
	return nil
}
