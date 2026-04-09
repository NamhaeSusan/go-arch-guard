package app

import (
	"github.com/kimtaeyun/testproject-dc-invalid/internal/domain/order/core/model"
	"github.com/kimtaeyun/testproject-dc-invalid/internal/domain/order/core/repo"
)

// OK: consumer-defined interface is allowed in app.
type AdminOps interface {
	GetUserByID(id string) (string, error)
}

// VIOLATION: type alias re-exports interface from core/repo
type OrderRepo = repo.Repository

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) CreateOrder(userID string, amount int) *model.Order {
	return &model.Order{ID: "1", UserID: userID, Amount: amount}
}
