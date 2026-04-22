package http

// BAD: transport must not import domain directly (should go through app).
import _ "github.com/kimtaeyun/testproject-ddd-appserver/internal/domain/order/core/model"
