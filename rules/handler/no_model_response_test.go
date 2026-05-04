package handler_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/analyzer"
	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/presets"
	"github.com/NamhaeSusan/go-arch-guard/rules/handler"
)

const noModelResponseRule = "handler.no-model-response"

func TestNoModelResponseSpec(t *testing.T) {
	rule := handler.NewNoModelResponse()
	spec := rule.Spec()

	if spec.ID != noModelResponseRule {
		t.Fatalf("Spec().ID = %q, want %q", spec.ID, noModelResponseRule)
	}
	if spec.DefaultSeverity != core.Warning {
		t.Fatalf("Spec().DefaultSeverity = %v, want Warning", spec.DefaultSeverity)
	}
	got := spec.ViolationIDs()
	if len(got) != 1 || got[0] != noModelResponseRule {
		t.Fatalf("ViolationIDs() = %v, want [%s]", got, noModelResponseRule)
	}

	errorRule := handler.NewNoModelResponse(handler.WithSeverity(core.Error))
	if errorRule.Spec().DefaultSeverity != core.Error {
		t.Fatalf("WithSeverity(Error) default = %v, want Error", errorRule.Spec().DefaultSeverity)
	}
}

func TestNoModelResponseDetectsDomainHandlerModelResponses(t *testing.T) {
	module := "example.com/shop"
	root := writeFixture(t, module, map[string]string{
		"internal/domain/order/core/model/order.go": `package model

type Order struct {
	ID string
}
`,
		"internal/domain/order/handler/http/alias.go": `package http

import "example.com/shop/internal/domain/order/core/model"

type OrderAliasResponse = model.Order
type OrderSliceResponse = []*model.Order
`,
		"internal/domain/order/handler/http/response.go": `package http

import "example.com/shop/internal/domain/order/core/model"

type OrderEnvelopeResponse struct {
	Order model.Order ` + "`json:\"order\"`" + `
}
`,
		"internal/domain/order/handler/http/handler.go": `package http

import "example.com/shop/internal/domain/order/core/model"

type Handler struct{}

func (h Handler) GetOrder() (model.Order, error) {
	return model.Order{}, nil
}

func (h Handler) ListOrders() ([]*model.Order, error) {
	return nil, nil
}
`,
	})

	violations := handler.NewNoModelResponse().Check(loadContext(t, root, module, presets.DDD()))

	assertViolationCount(t, violations, noModelResponseRule, 5)
	assertViolationAt(t, violations, "internal/domain/order/handler/http/alias.go", "OrderAliasResponse")
	assertViolationAt(t, violations, "internal/domain/order/handler/http/alias.go", "OrderSliceResponse")
	assertViolationAt(t, violations, "internal/domain/order/handler/http/response.go", "OrderEnvelopeResponse")
	assertViolationAt(t, violations, "internal/domain/order/handler/http/handler.go", "GetOrder")
	assertViolationAt(t, violations, "internal/domain/order/handler/http/handler.go", "ListOrders")
}

func TestNoModelResponseIgnoresRequestDTOsAndParams(t *testing.T) {
	module := "example.com/shop"
	root := writeFixture(t, module, map[string]string{
		"internal/domain/order/core/model/order.go": `package model

type Order struct {
	ID string
}
`,
		"internal/domain/order/handler/http/request.go": `package http

import "example.com/shop/internal/domain/order/core/model"

type OrderRequest struct {
	Order model.Order ` + "`json:\"order\"`" + `
}

type Handler struct{}

func (h Handler) CreateOrder(input OrderRequest, order model.Order) error {
	return nil
}
`,
	})

	violations := handler.NewNoModelResponse().Check(loadContext(t, root, module, presets.DDD()))

	assertNoRule(t, violations, noModelResponseRule)
}

func TestNoModelResponseAllowsLocalAndAppDTOResponses(t *testing.T) {
	module := "example.com/shop"
	root := writeFixture(t, module, map[string]string{
		"internal/domain/order/core/model/order.go": `package model

type Order struct {
	ID string
}
`,
		"internal/domain/order/app/output.go": `package app

type OrderOutput struct {
	ID string
}
`,
		"internal/domain/order/handler/http/response.go": `package http

import "example.com/shop/internal/domain/order/app"

type OrderResponse struct {
	ID    string          ` + "`json:\"id\"`" + `
	Order app.OrderOutput ` + "`json:\"order\"`" + `
}

type Handler struct{}

func (h Handler) GetOrder() (OrderResponse, error) {
	return OrderResponse{}, nil
}
`,
	})

	violations := handler.NewNoModelResponse().Check(loadContext(t, root, module, presets.DDD()))

	assertNoRule(t, violations, noModelResponseRule)
}

func TestNoModelResponseDetectsTopLevelTransportModelResponses(t *testing.T) {
	module := "example.com/shop"
	root := writeFixture(t, module, map[string]string{
		"internal/domain/order/core/model/order.go": `package model

type Order struct {
	ID string
}
`,
		"internal/server/http/handler.go": `package http

import "example.com/shop/internal/domain/order/core/model"

func GetOrder() (model.Order, error) {
	return model.Order{}, nil
}
`,
	})

	violations := handler.NewNoModelResponse().Check(loadContext(t, root, module, presets.DDD()))

	assertViolationCount(t, violations, noModelResponseRule, 1)
	assertViolationAt(t, violations, "internal/server/http/handler.go", "GetOrder")
}

func TestNoModelResponseDetectsModelWrittenToResponseHelpers(t *testing.T) {
	module := "example.com/shop"
	root := writeFixture(t, module, map[string]string{
		"internal/domain/order/core/model/order.go": `package model

type Order struct {
	ID string
}
`,
		"internal/domain/order/app/types.go": `package app

import "example.com/shop/internal/domain/order/core/model"

type Order = model.Order
`,
		"internal/domain/order/handler/http/handler.go": `package http

import "example.com/shop/internal/domain/order/app"

type Context struct{}

func (Context) JSON(code int, data any) {}
func OK(c any, data any) {}

type Handler struct{}

func (h Handler) WriteOrder(c Context, order app.Order) {
	OK(c, order)
	c.JSON(200, order)
}
`,
	})

	violations := handler.NewNoModelResponse().Check(loadContext(t, root, module, presets.DDD()))

	assertViolationCount(t, violations, noModelResponseRule, 2)
	assertViolationAt(t, violations, "internal/domain/order/handler/http/handler.go", "response body")
}

func TestNoModelResponseDetectsOrchestrationHandlerModelResponses(t *testing.T) {
	module := "example.com/shop"
	root := writeFixture(t, module, map[string]string{
		"internal/domain/order/core/model/order.go": `package model

type Order struct {
	ID string
}
`,
		"internal/orchestration/handler/http/response.go": `package http

import "example.com/shop/internal/domain/order/core/model"

type CheckoutResponse = model.Order
`,
	})

	violations := handler.NewNoModelResponse().Check(loadContext(t, root, module, presets.DDD()))

	assertViolationCount(t, violations, noModelResponseRule, 1)
	assertViolationAt(t, violations, "internal/orchestration/handler/http/response.go", "CheckoutResponse")
}

func TestNoModelResponseDetectsCleanArchEntityResponses(t *testing.T) {
	module := "example.com/clean"
	root := writeFixture(t, module, map[string]string{
		"internal/domain/catalog/entity/product.go": `package entity

type Product struct {
	ID string
}
`,
		"internal/domain/catalog/handler/http/response.go": `package http

import "example.com/clean/internal/domain/catalog/entity"

type ProductResponse = entity.Product
`,
	})

	violations := handler.NewNoModelResponse().Check(loadContext(t, root, module, presets.CleanArch()))

	assertViolationCount(t, violations, noModelResponseRule, 1)
	assertViolationAt(t, violations, "internal/domain/catalog/handler/http/response.go", "ProductResponse")
}

func TestNoModelResponseAllowsConfiguredModelTypes(t *testing.T) {
	module := "example.com/shop"
	root := writeFixture(t, module, map[string]string{
		"internal/domain/order/core/model/order.go": `package model

type Order struct {
	ID string
}
`,
		"internal/domain/order/handler/http/response.go": `package http

import "example.com/shop/internal/domain/order/core/model"

type OrderResponse = model.Order
`,
	})
	allowed := module + "/internal/domain/order/core/model.Order"

	violations := handler.NewNoModelResponse(handler.WithAllowedModelTypes("", allowed)).
		Check(loadContext(t, root, module, presets.DDD()))

	assertNoRule(t, violations, noModelResponseRule)
}

func TestNoModelResponseIsOptInForRecommendedDDD(t *testing.T) {
	for _, rule := range presets.RecommendedDDD().Rules() {
		if rule.Spec().ID == noModelResponseRule {
			t.Fatalf("%s must stay opt-in, found in RecommendedDDD()", noModelResponseRule)
		}
	}
}

func writeFixture(t *testing.T, module string, files map[string]string) string {
	t.Helper()

	root := t.TempDir()
	writeFixtureFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.26.1\n")
	for name, content := range files {
		writeFixtureFile(t, filepath.Join(root, name), content)
	}
	return root
}

func writeFixtureFile(t *testing.T, path, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func loadContext(t *testing.T, root, module string, arch core.Architecture) *core.Context {
	t.Helper()

	pkgs, err := analyzer.Load(root, "internal/...")
	if err != nil {
		t.Fatalf("load packages: %v", err)
	}
	return core.NewContext(pkgs, module, root, arch, nil)
}

func assertViolationCount(t *testing.T, violations []core.Violation, rule string, want int) {
	t.Helper()

	var got int
	for _, v := range violations {
		if v.Rule == rule {
			got++
		}
	}
	if got != want {
		t.Fatalf("%s violation count = %d, want %d; violations: %+v", rule, got, want, violations)
	}
}

func assertViolationAt(t *testing.T, violations []core.Violation, file, messagePart string) {
	t.Helper()

	for _, v := range violations {
		if v.Rule == noModelResponseRule && v.File == file && strings.Contains(v.Message, messagePart) {
			return
		}
	}
	t.Fatalf("missing %s violation at %s containing %q; got %+v", noModelResponseRule, file, messagePart, violations)
}

func assertNoRule(t *testing.T, violations []core.Violation, rule string) {
	t.Helper()

	for _, v := range violations {
		if v.Rule == rule {
			t.Fatalf("unexpected %s violation: %+v", rule, v)
		}
	}
}
