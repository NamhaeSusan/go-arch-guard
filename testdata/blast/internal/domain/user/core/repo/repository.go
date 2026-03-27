package repo

import "github.com/kimtaeyun/testproject-blast/internal/domain/user/core/model"

type Repository interface {
	Find(id string) (model.User, error)
}
