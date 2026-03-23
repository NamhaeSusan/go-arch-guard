package repo

import "github.com/kimtaeyun/testproject-dc-invalid/internal/domain/user/core/model"

type Repository interface {
	GetByID(id string) (*model.User, error)
}
