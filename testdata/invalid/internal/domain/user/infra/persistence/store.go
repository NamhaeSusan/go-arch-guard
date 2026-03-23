package persistence

import (
	"github.com/kimtaeyun/testproject-dc-invalid/internal/domain/user/core/model"
	"github.com/kimtaeyun/testproject-dc-invalid/internal/domain/user/core/repo"
)

var _ repo.Repository = (*Store)(nil)

type Store struct{}

func (s *Store) GetByID(id string) (*model.User, error) {
	return &model.User{ID: id, Name: "test", Email: "test@example.com"}, nil
}
