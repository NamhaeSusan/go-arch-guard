package repo

import "github.com/kimtaeyun/testproject-blast/internal/domain/order/core/model"

type Repository interface {
	Find(id string) (model.Order, error)
}
