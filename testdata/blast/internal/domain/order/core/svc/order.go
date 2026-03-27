package svc

import "github.com/kimtaeyun/testproject-blast/internal/domain/order/core/model"

type Validator interface {
	Validate(o model.Order) error
}
