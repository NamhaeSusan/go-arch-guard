package order

import "github.com/kimtaeyun/testproject-dc-invalid/internal/domain/order/app"

type Service = app.Service

var NewService = app.NewService

// VIOLATION: alias.go defines interface — suspected cross-domain dependency
type CrossDomainOps interface {
	GetExternalData() error
}
