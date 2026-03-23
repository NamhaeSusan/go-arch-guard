package http

import "github.com/kimtaeyun/testproject-dc/internal/domain/user/app"

type Handler struct {
	svc *app.Service
}

func NewHandler(s *app.Service) *Handler {
	return &Handler{svc: s}
}
