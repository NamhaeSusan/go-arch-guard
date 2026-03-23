package saga

import (
	"github.com/kimtaeyun/testproject-dc/internal/domain/order"
	"github.com/kimtaeyun/testproject-dc/internal/domain/user"
)

type CreateOrderSaga struct {
	userSvc  *user.Service
	orderSvc *order.Service
}

func NewCreateOrderSaga(us *user.Service, os *order.Service) *CreateOrderSaga {
	return &CreateOrderSaga{userSvc: us, orderSvc: os}
}
