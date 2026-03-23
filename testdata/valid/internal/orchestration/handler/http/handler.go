package http

import "github.com/kimtaeyun/testproject-dc/internal/orchestration"

type Handler struct {
	createOrder *orchestration.CreateOrderOrchestration
}

func NewHandler(co *orchestration.CreateOrderOrchestration) *Handler {
	return &Handler{createOrder: co}
}
