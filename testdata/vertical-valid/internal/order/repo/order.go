package repo

import "github.com/kimtaeyun/testproject-vertical/internal/order/model"

type Order interface {
	Save(order *model.Order) error
}
