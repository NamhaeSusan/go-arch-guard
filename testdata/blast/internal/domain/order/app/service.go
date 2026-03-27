package app

import (
	"github.com/kimtaeyun/testproject-blast/internal/domain/order/core/model"
	"github.com/kimtaeyun/testproject-blast/internal/domain/order/core/repo"
	"github.com/kimtaeyun/testproject-blast/internal/domain/order/core/svc"
	"github.com/kimtaeyun/testproject-blast/internal/domain/order/event"
	"github.com/kimtaeyun/testproject-blast/internal/pkg"
)

type Service struct{}

func (s Service) Create() {
	_ = model.Order{}
	_ = (*repo.Repository)(nil)
	_ = (*svc.Validator)(nil)
	_ = event.OrderCreated{}
	_ = pkg.SharedHelper()
}
