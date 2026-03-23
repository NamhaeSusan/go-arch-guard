package orchestration

import (
	"github.com/kimtaeyun/testproject-dc-invalid/internal/domain/order"
	"github.com/kimtaeyun/testproject-dc-invalid/internal/domain/user"
)

type CreateOrderOrchestration struct {
	userSvc  *user.Service
	orderSvc *order.Service
}

func NewCreateOrderOrchestration(us *user.Service, os *order.Service) *CreateOrderOrchestration {
	return &CreateOrderOrchestration{userSvc: us, orderSvc: os}
}
