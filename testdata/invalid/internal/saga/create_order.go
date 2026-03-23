package saga

import (
	"github.com/kimtaeyun/testproject-dc-invalid/internal/domain/order"
	"github.com/kimtaeyun/testproject-dc-invalid/internal/domain/user"
)

type CreateOrderSaga struct {
	userSvc  *user.Service
	orderSvc *order.Service
}

func NewCreateOrderSaga(us *user.Service, os *order.Service) *CreateOrderSaga {
	return &CreateOrderSaga{userSvc: us, orderSvc: os}
}
