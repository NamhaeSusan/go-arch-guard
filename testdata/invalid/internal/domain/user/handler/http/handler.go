package http

import "github.com/kimtaeyun/testproject-dc-invalid/internal/domain/user/app"

type Handler struct {
	svc *app.Service
}

func NewHandler(s *app.Service) *Handler {
	return &Handler{svc: s}
}
