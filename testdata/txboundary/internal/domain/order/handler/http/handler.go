package http

import "github.com/kimtaeyun/testproject-txboundary/internal/domain/order/core/model"

type Handler struct{}

func (h *Handler) Serve(o model.Order) error {
	_ = o
	return nil
}
