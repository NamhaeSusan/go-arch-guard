package app

import (
	_ "github.com/kimtaeyun/testproject-dc-invalid/internal/orchestration"

	"github.com/kimtaeyun/testproject-dc-invalid/internal/domain/payment/core/model"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Create() *model.Payment {
	return &model.Payment{ID: "p1"}
}
