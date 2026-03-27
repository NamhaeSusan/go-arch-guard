package app

import (
	"github.com/kimtaeyun/testproject-blast/internal/domain/shipping/core/model"
	"github.com/kimtaeyun/testproject-blast/internal/pkg"
)

type Service struct{}

func (s Service) Ship() {
	_ = model.Shipping{}
	_ = pkg.SharedHelper()
}
