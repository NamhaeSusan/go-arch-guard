package event

import (
	"github.com/kimtaeyun/testproject-dc-invalid/internal/domain/payment/core/model"
	_ "github.com/kimtaeyun/testproject-dc-invalid/internal/pkg"
)

type Created struct {
	Payment model.Payment
}
