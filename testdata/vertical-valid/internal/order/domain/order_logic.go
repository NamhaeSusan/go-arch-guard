package domain

import (
	"errors"

	"github.com/kimtaeyun/testproject-vertical/internal/order/model"
)

func Validate(o *model.Order) error {
	if o.Amount <= 0 {
		return errors.New("invalid amount")
	}
	return nil
}
