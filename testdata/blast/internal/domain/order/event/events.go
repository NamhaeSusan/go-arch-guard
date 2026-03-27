package event

import "github.com/kimtaeyun/testproject-blast/internal/domain/order/core/model"

type OrderCreated struct{ Order model.Order }
