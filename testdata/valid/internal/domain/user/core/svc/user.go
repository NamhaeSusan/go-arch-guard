package svc

import (
	"errors"

	"github.com/kimtaeyun/testproject-dc/internal/domain/user/core/model"
)

func Validate(u *model.User) error {
	if u.Name == "" {
		return errors.New("name is required")
	}
	return nil
}
