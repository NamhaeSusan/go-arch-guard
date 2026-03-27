package app

import (
	"github.com/kimtaeyun/testproject-blast/internal/domain/user/core/model"
	"github.com/kimtaeyun/testproject-blast/internal/domain/user/core/repo"
	"github.com/kimtaeyun/testproject-blast/internal/pkg"
)

type Service struct{}

func (s Service) Get() {
	_ = model.User{}
	_ = (*repo.Repository)(nil)
	_ = pkg.SharedHelper()
}
