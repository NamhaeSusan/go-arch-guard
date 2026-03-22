package app

import "github.com/kimtaeyun/testproject-vertical/internal/user/model"

type Service struct{}

func (s *Service) GetUser(id string) *model.User {
	return &model.User{ID: id}
}
