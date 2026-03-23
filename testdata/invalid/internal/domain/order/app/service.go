package app

import "github.com/kimtaeyun/testproject-dc-invalid/internal/domain/order/core/model"

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) CreateOrder(userID string, amount int) *model.Order {
	return &model.Order{ID: "1", UserID: userID, Amount: amount}
}
