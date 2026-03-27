package app

import (
	"github.com/kimtaeyun/testproject-blast/internal/domain/product/core/model"
	"github.com/kimtaeyun/testproject-blast/internal/pkg"
)

type Service struct{}

func (s Service) Get() {
	_ = model.Product{}
	_ = pkg.SharedHelper()
}
