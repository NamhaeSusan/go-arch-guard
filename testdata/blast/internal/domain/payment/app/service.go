package app

import (
	"github.com/kimtaeyun/testproject-blast/internal/domain/payment/core/model"
	"github.com/kimtaeyun/testproject-blast/internal/pkg"
)

type Service struct{}

func (s Service) Pay() {
	_ = model.Payment{}
	_ = pkg.SharedHelper()
}
