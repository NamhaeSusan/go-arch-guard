package repo

import "github.com/kimtaeyun/testproject-vertical/internal/order/model"

type Repository interface {
	Save(order *model.Order) error
}
