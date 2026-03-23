package repo

import "github.com/kimtaeyun/testproject-dc-invalid/internal/domain/order/core/model"

type Repository interface {
	Save(o *model.Order) error
}
