package policy

import "github.com/kimtaeyun/testproject-dc-invalid/internal/domain/payment/core/model"

func Allow(payment model.Payment) bool {
	return payment.ID != ""
}
