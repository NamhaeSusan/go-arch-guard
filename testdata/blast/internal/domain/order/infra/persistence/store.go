package persistence

import (
	"github.com/kimtaeyun/testproject-blast/internal/domain/order/core/model"
	"github.com/kimtaeyun/testproject-blast/internal/domain/order/core/repo"
	"github.com/kimtaeyun/testproject-blast/internal/pkg"
)

type Store struct{}

func (s Store) Find(id string) (model.Order, error) {
	_ = pkg.SharedHelper()
	return model.Order{ID: id}, nil
}

var _ repo.Repository = Store{}
