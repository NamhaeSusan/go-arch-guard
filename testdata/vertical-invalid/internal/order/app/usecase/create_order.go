package usecase

import (
	"github.com/kimtaeyun/testproject-vertical-invalid/internal/order/model"
	"github.com/kimtaeyun/testproject-vertical-invalid/internal/user"
	userPort "github.com/kimtaeyun/testproject-vertical-invalid/internal/user/port"
)

type CreateOrder struct {
	userSvc *user.Service
}

func (uc *CreateOrder) Execute(userID string, amount int) (*model.Order, error) {
	_ = userPort.UserResponse{}
	return &model.Order{ID: "1", UserID: userID, Amount: amount}, nil
}
