package http

import "github.com/kimtaeyun/testproject-dc-invalid/internal/saga"

type Handler struct {
	createOrder *saga.CreateOrderSaga
}

func NewHandler(co *saga.CreateOrderSaga) *Handler {
	return &Handler{createOrder: co}
}
