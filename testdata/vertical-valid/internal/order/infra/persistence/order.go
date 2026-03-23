package persistence

import "github.com/kimtaeyun/testproject-vertical/internal/order/model"

type Store struct{}

func (s *Store) Save(o *model.Order) error { return nil }
