package event

import "github.com/kimtaeyun/testproject-dc/internal/domain/order/core/model"

type Created struct {
	OrderID string
}

func NewCreated(order model.Order) Created {
	return Created{OrderID: order.ID}
}
