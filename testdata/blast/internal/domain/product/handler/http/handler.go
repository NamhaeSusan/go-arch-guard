package http

import (
	"github.com/kimtaeyun/testproject-blast/internal/domain/product/app"
	"github.com/kimtaeyun/testproject-blast/internal/pkg"
)

type Handler struct{ Svc app.Service }

func (h Handler) Handle() { _ = pkg.SharedHelper() }
