package persistence

import "github.com/kimtaeyun/testproject-dc/internal/domain/order/core/model"

type Store struct{}

func (s *Store) Save(o *model.Order) error {
	return nil
}
