package svc

import (
	"database/sql"
)

type Svc struct{}

// Begin returns *sql.Tx — VIOLATION (tx.type-in-signature on result).
func (s *Svc) Begin() *sql.Tx {
	return nil
}
