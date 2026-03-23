package app

import (
	"github.com/kimtaeyun/testproject-dc/internal/domain/user/core/model"
	"github.com/kimtaeyun/testproject-dc/internal/domain/user/core/repo"
	"github.com/kimtaeyun/testproject-dc/internal/domain/user/core/svc"
)

type Service struct {
	repo repo.Repository
}

func NewService(r repo.Repository) *Service {
	return &Service{repo: r}
}

func (s *Service) GetUser(id string) (*model.User, error) {
	u, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if err := svc.Validate(u); err != nil {
		return nil, err
	}
	return u, nil
}
