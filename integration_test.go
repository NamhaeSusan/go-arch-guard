package goarchguard_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/analyzer"
	"github.com/NamhaeSusan/go-arch-guard/report"
	"github.com/NamhaeSusan/go-arch-guard/rules"
	"golang.org/x/tools/go/packages"
)

func TestIntegration_Valid(t *testing.T) {
	pkgs, err := analyzer.Load("testdata/valid", "internal/...", "cmd/...")
	if err != nil {
		t.Fatal(err)
	}

	t.Run("domain isolation", func(t *testing.T) {
		report.AssertNoViolations(t, rules.CheckDomainIsolation(pkgs, "github.com/kimtaeyun/testproject-dc", "testdata/valid"))
	})
	t.Run("layer direction", func(t *testing.T) {
		report.AssertNoViolations(t, rules.CheckLayerDirection(pkgs, "github.com/kimtaeyun/testproject-dc", "testdata/valid"))
	})
	t.Run("naming", func(t *testing.T) {
		report.AssertNoViolations(t, rules.CheckNaming(pkgs))
	})
	t.Run("structure", func(t *testing.T) {
		report.AssertNoViolations(t, rules.CheckStructure("testdata/valid"))
	})
	t.Run("blast radius", func(t *testing.T) {
		report.AssertNoViolations(t, rules.AnalyzeBlastRadius(pkgs, "github.com/kimtaeyun/testproject-dc", "testdata/valid"))
	})
}

func TestIntegration_BlastRadius(t *testing.T) {
	pkgs, err := analyzer.Load("testdata/blast", "internal/...")
	if err != nil {
		t.Fatal(err)
	}

	violations := rules.AnalyzeBlastRadius(pkgs, "github.com/kimtaeyun/testproject-blast", "testdata/blast")
	if len(violations) == 0 {
		t.Error("expected blast radius violations for hub package")
	}
	assertHasRule(t, violations, "blast-radius.high-coupling")
}

func TestIntegration_Invalid(t *testing.T) {
	pkgs, err := analyzer.Load("testdata/invalid", "internal/...", "cmd/...")
	if err != nil {
		t.Fatal(err)
	}

	t.Run("domain isolation violations found", func(t *testing.T) {
		violations := rules.CheckDomainIsolation(pkgs, "github.com/kimtaeyun/testproject-dc-invalid", "testdata/invalid")
		assertHasRule(t, violations, "isolation.cross-domain")
	})
	t.Run("layer direction violations found", func(t *testing.T) {
		violations := rules.CheckLayerDirection(pkgs, "github.com/kimtaeyun/testproject-dc-invalid", "testdata/invalid")
		assertHasRule(t, violations, "layer.direction")
	})
	t.Run("naming violations found", func(t *testing.T) {
		violations := rules.CheckNaming(pkgs)
		assertHasRule(t, violations, "naming.no-layer-suffix")
	})
	t.Run("structure violations found", func(t *testing.T) {
		violations := rules.CheckStructure("testdata/invalid")
		assertHasRule(t, violations, "structure.banned-package")
	})

	t.Run("new rule ids are surfaced", func(t *testing.T) {
		isolationViolations := rules.CheckDomainIsolation(pkgs, "github.com/kimtaeyun/testproject-dc-invalid", "testdata/invalid")
		layerViolations := rules.CheckLayerDirection(pkgs, "github.com/kimtaeyun/testproject-dc-invalid", "testdata/invalid")
		structureViolations := rules.CheckStructure("testdata/invalid")

		assertHasRule(t, isolationViolations, "isolation.domain-imports-orchestration")
		assertHasRule(t, isolationViolations, "isolation.internal-imports-orchestration")
		assertHasRule(t, isolationViolations, "isolation.pkg-imports-domain")
		assertHasRule(t, layerViolations, "layer.unknown-sublayer")
		assertHasRule(t, layerViolations, "layer.inner-imports-pkg")
		assertHasRule(t, structureViolations, "structure.internal-top-level")
		assertHasRule(t, structureViolations, "structure.domain-root-alias-required")
		assertHasRule(t, structureViolations, "structure.domain-model-required")
		assertHasRule(t, structureViolations, "structure.dto-placement")
		assertHasRule(t, structureViolations, "structure.misplaced-layer")
	})
}

func TestIntegration_WarningMode(t *testing.T) {
	pkgs, err := analyzer.Load("testdata/invalid", "internal/...", "cmd/...")
	if err != nil {
		t.Fatal(err)
	}

	violations := rules.CheckDomainIsolation(pkgs, "github.com/kimtaeyun/testproject-dc-invalid", "testdata/invalid",
		rules.WithSeverity(rules.Warning))
	if len(violations) == 0 {
		t.Error("expected violations even in warning mode")
	}
	for _, v := range violations {
		if v.Severity != rules.Warning {
			t.Errorf("expected Warning severity, got %v", v.Severity)
		}
	}
	report.AssertNoViolations(t, violations)
}

func TestIntegration_RejectsUnexpectedInternalTopLevelPackages(t *testing.T) {
	root := t.TempDir()
	module := "example.com/supportzones"

	writeIntegrationFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.25.0\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "order", "alias.go"), "package order\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "order", "core", "model", "order.go"), "package model\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "config", "config.go"), "package config\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "platform", "platform.go"), "package platform\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "system", "system.go"), "package system\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "foundation", "foundation.go"), "package foundation\n")

	pkgs, err := analyzer.Load(root, "internal/...")
	if err != nil {
		t.Fatal(err)
	}

	assertHasPackage(t, pkgs, module+"/internal/config")
	assertHasPackage(t, pkgs, module+"/internal/platform")
	assertHasPackage(t, pkgs, module+"/internal/system")
	assertHasPackage(t, pkgs, module+"/internal/foundation")

	t.Run("domain isolation", func(t *testing.T) {
		report.AssertNoViolations(t, rules.CheckDomainIsolation(pkgs, module, root))
	})
	t.Run("layer direction", func(t *testing.T) {
		report.AssertNoViolations(t, rules.CheckLayerDirection(pkgs, module, root))
	})
	t.Run("naming", func(t *testing.T) {
		report.AssertNoViolations(t, rules.CheckNaming(pkgs))
	})
	t.Run("structure", func(t *testing.T) {
		violations := rules.CheckStructure(root)
		assertHasRule(t, violations, "structure.internal-top-level")
	})
}

func assertHasRule(t *testing.T, violations []rules.Violation, rule string) {
	t.Helper()
	for _, v := range violations {
		if v.Rule == rule {
			return
		}
	}
	seen := make(map[string]bool)
	for _, v := range violations {
		seen[v.Rule] = true
	}
	var got []string
	for r := range seen {
		got = append(got, r)
	}
	t.Fatalf("expected rule %q, got rules: %v", rule, got)
}

func assertHasPackage(t *testing.T, pkgs []*packages.Package, pkgPath string) {
	t.Helper()
	for _, pkg := range pkgs {
		if pkg.PkgPath == pkgPath {
			return
		}
	}
	t.Fatalf("expected package %q to be loaded", pkgPath)
}

func TestIntegration_CleanArchModel(t *testing.T) {
	root := t.TempDir()
	module := "example.com/cleanapp"
	m := rules.CleanArch()

	writeIntegrationFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.25.0\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "order", "handler", "handler.go"),
		"package handler\n\nimport _ \""+module+"/internal/domain/order/usecase\"\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "order", "usecase", "usecase.go"),
		"package usecase\n\nimport _ \""+module+"/internal/domain/order/entity\"\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "order", "entity", "order.go"),
		"package entity\n\ntype Order struct{ ID string }\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "order", "gateway", "repo.go"),
		"package gateway\n\nimport _ \""+module+"/internal/domain/order/entity\"\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "order", "infra", "persistence.go"),
		"package infra\n\nimport (\n\t_ \""+module+"/internal/domain/order/gateway\"\n\t_ \""+module+"/internal/domain/order/entity\"\n)\n")

	pkgs, err := analyzer.Load(root, "internal/...")
	if err != nil {
		t.Fatal(err)
	}

	opts := []rules.Option{rules.WithModel(m)}

	t.Run("layer direction", func(t *testing.T) {
		report.AssertNoViolations(t, rules.CheckLayerDirection(pkgs, module, root, opts...))
	})
	t.Run("domain isolation", func(t *testing.T) {
		report.AssertNoViolations(t, rules.CheckDomainIsolation(pkgs, module, root, opts...))
	})
	t.Run("structure", func(t *testing.T) {
		report.AssertNoViolations(t, rules.CheckStructure(root, opts...))
	})
	t.Run("naming", func(t *testing.T) {
		report.AssertNoViolations(t, rules.CheckNaming(pkgs, opts...))
	})
}

func TestIntegration_CustomModel(t *testing.T) {
	root := t.TempDir()
	module := "example.com/custom"
	m := rules.NewModel(
		rules.WithDomainDir("module"),
		rules.WithSharedDir("lib"),
		rules.WithSublayers([]string{"api", "logic", "data"}),
		rules.WithDirection(map[string][]string{
			"api":   {"logic"},
			"logic": {"data"},
			"data":  {},
		}),
		rules.WithPkgRestricted(map[string]bool{"data": true}),
		rules.WithRequireAlias(false),
		rules.WithRequireModel(false),
	)

	writeIntegrationFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.25.0\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "module", "order", "api", "handler.go"),
		"package api\n\nimport _ \""+module+"/internal/module/order/logic\"\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "module", "order", "logic", "service.go"),
		"package logic\n\nimport _ \""+module+"/internal/module/order/data\"\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "module", "order", "data", "repo.go"),
		"package data\n\ntype OrderRepo struct{}\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "lib", "logger", "logger.go"),
		"package logger\n")

	pkgs, err := analyzer.Load(root, "internal/...")
	if err != nil {
		t.Fatal(err)
	}

	opts := []rules.Option{rules.WithModel(m)}

	t.Run("layer direction", func(t *testing.T) {
		report.AssertNoViolations(t, rules.CheckLayerDirection(pkgs, module, root, opts...))
	})
	t.Run("domain isolation", func(t *testing.T) {
		report.AssertNoViolations(t, rules.CheckDomainIsolation(pkgs, module, root, opts...))
	})
	t.Run("structure", func(t *testing.T) {
		report.AssertNoViolations(t, rules.CheckStructure(root, opts...))
	})
	t.Run("naming", func(t *testing.T) {
		report.AssertNoViolations(t, rules.CheckNaming(pkgs, opts...))
	})
}

func TestIntegration_CleanArchModel_DirectionViolation(t *testing.T) {
	root := t.TempDir()
	module := "example.com/cleanviolation"
	m := rules.CleanArch()

	// usecase imports handler — should violate direction (usecase can only import entity, gateway)
	writeIntegrationFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.25.0\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "order", "usecase", "usecase.go"),
		"package usecase\n\nimport _ \""+module+"/internal/domain/order/handler\"\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "order", "handler", "handler.go"),
		"package handler\n")

	pkgs, err := analyzer.Load(root, "internal/...")
	if err != nil {
		t.Fatal(err)
	}

	violations := rules.CheckLayerDirection(pkgs, module, root, rules.WithModel(m))
	assertHasRule(t, violations, "layer.direction")
}

func TestIntegration_CustomModel_DirectionViolation(t *testing.T) {
	root := t.TempDir()
	module := "example.com/custom2"
	m := rules.NewModel(
		rules.WithDomainDir("module"),
		rules.WithSharedDir("lib"),
		rules.WithSublayers([]string{"api", "logic", "data"}),
		rules.WithDirection(map[string][]string{
			"api":   {"logic"},
			"logic": {"data"},
			"data":  {},
		}),
		rules.WithRequireAlias(false),
		rules.WithRequireModel(false),
	)

	// api imports data directly — should violate direction (api can only import logic)
	writeIntegrationFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.25.0\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "module", "order", "api", "handler.go"),
		"package api\n\nimport _ \""+module+"/internal/module/order/data\"\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "module", "order", "data", "repo.go"),
		"package data\n")

	pkgs, err := analyzer.Load(root, "internal/...")
	if err != nil {
		t.Fatal(err)
	}

	violations := rules.CheckLayerDirection(pkgs, module, root, rules.WithModel(m))
	assertHasRule(t, violations, "layer.direction")
}

// TestIntegration_RealWorldDDD simulates a realistic multi-domain DDD project
// with orchestration, pkg, cmd, cross-domain isolation, and all layer directions.
func TestIntegration_RealWorldDDD(t *testing.T) {
	root := t.TempDir()
	module := "example.com/shop"

	writeIntegrationFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.25.0\n")

	// --- domain: order ---
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "order", "alias.go"),
		`package order

import "example.com/shop/internal/domain/order/app"

type Service = app.Service
var NewService = app.NewService
`)
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "order", "core", "model", "order.go"),
		"package model\n\ntype Order struct{ ID string }\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "order", "core", "repo", "order.go"),
		"package repo\n\nimport \"example.com/shop/internal/domain/order/core/model\"\n\ntype Order interface{ Find(id string) (model.Order, error) }\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "order", "app", "service.go"),
		`package app

import (
	"example.com/shop/internal/domain/order/core/model"
	"example.com/shop/internal/domain/order/core/repo"
)

type Service struct {
	repo repo.Order
}
func NewService(r repo.Order) *Service { return &Service{repo: r} }
func (s *Service) Get(id string) (model.Order, error) { return s.repo.Find(id) }
`)
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "order", "handler", "http", "handler.go"),
		"package http\n\nimport \"example.com/shop/internal/domain/order/app\"\n\ntype Handler struct{ svc *app.Service }\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "order", "event", "order_created.go"),
		"package event\n\nimport \"example.com/shop/internal/domain/order/core/model\"\n\ntype OrderCreated struct{ Order model.Order }\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "order", "infra", "persistence", "pg.go"),
		`package persistence

import (
	"example.com/shop/internal/domain/order/core/model"
	"example.com/shop/internal/domain/order/core/repo"
)

type Pg struct{}
func (Pg) Find(_ string) (model.Order, error) { return model.Order{}, nil }
var _ repo.Order = Pg{}
`)

	// --- domain: user ---
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "user", "alias.go"),
		`package user

import "example.com/shop/internal/domain/user/app"

type Service = app.Service
var NewService = app.NewService
`)
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "user", "core", "model", "user.go"),
		"package model\n\ntype User struct{ ID string; Name string }\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "user", "core", "repo", "user.go"),
		"package repo\n\nimport \"example.com/shop/internal/domain/user/core/model\"\n\ntype User interface{ GetByID(id string) (model.User, error) }\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "user", "app", "service.go"),
		`package app

import (
	"example.com/shop/internal/domain/user/core/model"
	"example.com/shop/internal/domain/user/core/repo"
)

type Service struct{ repo repo.User }
func NewService(r repo.User) *Service { return &Service{repo: r} }
func (s *Service) Get(id string) (model.User, error) { return s.repo.GetByID(id) }
`)
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "user", "infra", "persistence", "pg.go"),
		`package persistence

import (
	"example.com/shop/internal/domain/user/core/model"
	"example.com/shop/internal/domain/user/core/repo"
)

type Pg struct{}
func (Pg) GetByID(_ string) (model.User, error) { return model.User{}, nil }
var _ repo.User = Pg{}
`)

	// --- orchestration ---
	writeIntegrationFile(t, filepath.Join(root, "internal", "orchestration", "create_order.go"),
		`package orchestration

import (
	"example.com/shop/internal/domain/order"
	"example.com/shop/internal/domain/user"
)

type CreateOrder struct {
	orderSvc *order.Service
	userSvc  *user.Service
}
func New(o *order.Service, u *user.Service) *CreateOrder { return &CreateOrder{orderSvc: o, userSvc: u} }
`)

	// --- pkg ---
	writeIntegrationFile(t, filepath.Join(root, "internal", "pkg", "middleware", "auth.go"),
		"package middleware\n\nfunc Auth() {}\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "pkg", "logger", "logger.go"),
		"package logger\n\nfunc Info(_ string) {}\n")

	// --- cmd ---
	writeIntegrationFile(t, filepath.Join(root, "cmd", "api", "main.go"),
		`package main

import (
	_ "example.com/shop/internal/domain/order"
	_ "example.com/shop/internal/domain/user"
)

func main() {}
`)

	pkgs, err := analyzer.Load(root, "internal/...", "cmd/...")
	if err != nil {
		t.Log("partial load:", err)
	}
	if len(pkgs) == 0 {
		t.Fatal("no packages loaded")
	}

	t.Run("domain isolation", func(t *testing.T) {
		violations := rules.CheckDomainIsolation(pkgs, module, root)
		for _, v := range violations {
			t.Log(v.String())
		}
		report.AssertNoViolations(t, violations)
	})
	t.Run("layer direction", func(t *testing.T) {
		violations := rules.CheckLayerDirection(pkgs, module, root)
		for _, v := range violations {
			t.Log(v.String())
		}
		report.AssertNoViolations(t, violations)
	})
	t.Run("naming", func(t *testing.T) {
		violations := rules.CheckNaming(pkgs)
		for _, v := range violations {
			t.Log(v.String())
		}
		report.AssertNoViolations(t, violations)
	})
	t.Run("structure", func(t *testing.T) {
		violations := rules.CheckStructure(root)
		for _, v := range violations {
			t.Log(v.String())
		}
		report.AssertNoViolations(t, violations)
	})
}

// TestIntegration_RealWorldCleanArch simulates a realistic Clean Architecture project
// with multiple domains, cross-domain isolation via orchestration, and all layer directions.
func TestIntegration_RealWorldCleanArch(t *testing.T) {
	root := t.TempDir()
	module := "example.com/cleanshop"
	m := rules.CleanArch()

	writeIntegrationFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.25.0\n")

	// --- domain: product ---
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "product", "entity", "product.go"),
		"package entity\n\ntype Product struct{ ID string; Price int }\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "product", "gateway", "product.go"),
		"package gateway\n\nimport \"example.com/cleanshop/internal/domain/product/entity\"\n\ntype Product interface{ FindByID(id string) (entity.Product, error) }\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "product", "usecase", "get_product.go"),
		`package usecase

import (
	"example.com/cleanshop/internal/domain/product/entity"
	"example.com/cleanshop/internal/domain/product/gateway"
)

type GetProduct struct{ gw gateway.Product }
func NewGetProduct(gw gateway.Product) *GetProduct { return &GetProduct{gw: gw} }
func (uc *GetProduct) Execute(id string) (entity.Product, error) { return uc.gw.FindByID(id) }
`)
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "product", "handler", "http.go"),
		"package handler\n\nimport \"example.com/cleanshop/internal/domain/product/usecase\"\n\ntype HTTP struct{ uc *usecase.GetProduct }\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "product", "infra", "pg_product.go"),
		`package infra

import (
	"example.com/cleanshop/internal/domain/product/entity"
	"example.com/cleanshop/internal/domain/product/gateway"
)

type PgProduct struct{}
func (PgProduct) FindByID(_ string) (entity.Product, error) { return entity.Product{}, nil }
var _ gateway.Product = PgProduct{}
`)

	// --- domain: cart ---
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "cart", "entity", "cart.go"),
		"package entity\n\ntype Cart struct{ ID string; Items []string }\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "cart", "gateway", "cart.go"),
		"package gateway\n\nimport \"example.com/cleanshop/internal/domain/cart/entity\"\n\ntype Cart interface{ Save(c entity.Cart) error }\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "cart", "usecase", "add_to_cart.go"),
		`package usecase

import (
	"example.com/cleanshop/internal/domain/cart/entity"
	"example.com/cleanshop/internal/domain/cart/gateway"
)

type AddToCart struct{ gw gateway.Cart }
func NewAddToCart(gw gateway.Cart) *AddToCart { return &AddToCart{gw: gw} }
func (uc *AddToCart) Execute(c entity.Cart) error { return uc.gw.Save(c) }
`)
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "cart", "infra", "redis_cart.go"),
		`package infra

import (
	"example.com/cleanshop/internal/domain/cart/entity"
	"example.com/cleanshop/internal/domain/cart/gateway"
)

type RedisCart struct{}
func (RedisCart) Save(_ entity.Cart) error { return nil }
var _ gateway.Cart = RedisCart{}
`)

	// --- pkg ---
	writeIntegrationFile(t, filepath.Join(root, "internal", "pkg", "logger", "logger.go"),
		"package logger\n\nfunc Log(_ string) {}\n")

	pkgs, err := analyzer.Load(root, "internal/...")
	if err != nil {
		t.Log("partial load:", err)
	}
	if len(pkgs) == 0 {
		t.Fatal("no packages loaded")
	}

	opts := []rules.Option{rules.WithModel(m)}

	t.Run("layer direction", func(t *testing.T) {
		violations := rules.CheckLayerDirection(pkgs, module, root, opts...)
		for _, v := range violations {
			t.Log(v.String())
		}
		report.AssertNoViolations(t, violations)
	})
	t.Run("domain isolation", func(t *testing.T) {
		violations := rules.CheckDomainIsolation(pkgs, module, root, opts...)
		for _, v := range violations {
			t.Log(v.String())
		}
		report.AssertNoViolations(t, violations)
	})
	t.Run("structure", func(t *testing.T) {
		violations := rules.CheckStructure(root, opts...)
		for _, v := range violations {
			t.Log(v.String())
		}
		report.AssertNoViolations(t, violations)
	})
	t.Run("naming", func(t *testing.T) {
		violations := rules.CheckNaming(pkgs, opts...)
		for _, v := range violations {
			t.Log(v.String())
		}
		report.AssertNoViolations(t, violations)
	})
}

// TestIntegration_RealWorldCleanArch_Violations tests that CleanArch model catches
// various violations: cross-domain imports, wrong direction, unknown sublayers.
func TestIntegration_RealWorldCleanArch_Violations(t *testing.T) {
	root := t.TempDir()
	module := "example.com/cleaninvalid"
	m := rules.CleanArch()

	writeIntegrationFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.25.0\n")

	// domain: order — usecase imports handler (direction violation)
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "order", "entity", "order.go"),
		"package entity\n\ntype Order struct{ ID string }\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "order", "handler", "handler.go"),
		"package handler\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "order", "usecase", "order_uc.go"),
		"package usecase\n\nimport _ \"example.com/cleaninvalid/internal/domain/order/handler\"\n")

	// domain: user — entity imports from pkg (pkg-restricted violation)
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "user", "entity", "user.go"),
		"package entity\n\nimport _ \"example.com/cleaninvalid/internal/pkg/logger\"\n\ntype User struct{ ID string }\n")

	// cross-domain: order imports user directly
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "order", "infra", "cross.go"),
		"package infra\n\nimport _ \"example.com/cleaninvalid/internal/domain/user/entity\"\n")

	// unknown sublayer
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "order", "service", "bad.go"),
		"package service\n")

	writeIntegrationFile(t, filepath.Join(root, "internal", "pkg", "logger", "logger.go"),
		"package logger\n")

	pkgs, err := analyzer.Load(root, "internal/...")
	if err != nil {
		t.Log("partial load:", err)
	}

	opts := []rules.Option{rules.WithModel(m)}

	layerViolations := rules.CheckLayerDirection(pkgs, module, root, opts...)
	t.Run("detects direction violation", func(t *testing.T) {
		assertHasRule(t, layerViolations, "layer.direction")
	})
	t.Run("detects entity importing pkg", func(t *testing.T) {
		assertHasRule(t, layerViolations, "layer.inner-imports-pkg")
	})
	t.Run("detects unknown sublayer", func(t *testing.T) {
		assertHasRule(t, layerViolations, "layer.unknown-sublayer")
	})

	isolationViolations := rules.CheckDomainIsolation(pkgs, module, root, opts...)
	t.Run("detects cross-domain import", func(t *testing.T) {
		assertHasRule(t, isolationViolations, "isolation.cross-domain")
	})
}

func TestIntegration_RealWorldLayered(t *testing.T) {
	root := t.TempDir()
	module := "example.com/layeredshop"
	m := rules.Layered()

	writeIntegrationFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.25.0\n")

	// --- domain: order ---
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "order", "handler", "http.go"),
		"package handler\n\nimport _ \""+module+"/internal/domain/order/service\"\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "order", "service", "order.go"),
		"package service\n\nimport (\n\t_ \""+module+"/internal/domain/order/repository\"\n\t_ \""+module+"/internal/domain/order/model\"\n)\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "order", "repository", "order.go"),
		"package repository\n\nimport _ \""+module+"/internal/domain/order/model\"\n\ntype Order struct{}\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "order", "model", "order.go"),
		"package model\n\ntype Order struct{ ID string }\n")

	// --- pkg ---
	writeIntegrationFile(t, filepath.Join(root, "internal", "pkg", "logger", "logger.go"),
		"package logger\n\nfunc Log(_ string) {}\n")

	pkgs, err := analyzer.Load(root, "internal/...")
	if err != nil {
		t.Log("partial load:", err)
	}
	if len(pkgs) == 0 {
		t.Fatal("no packages loaded")
	}

	opts := []rules.Option{rules.WithModel(m)}

	t.Run("layer direction", func(t *testing.T) {
		violations := rules.CheckLayerDirection(pkgs, module, root, opts...)
		for _, v := range violations {
			t.Log(v.String())
		}
		report.AssertNoViolations(t, violations)
	})
	t.Run("domain isolation", func(t *testing.T) {
		violations := rules.CheckDomainIsolation(pkgs, module, root, opts...)
		for _, v := range violations {
			t.Log(v.String())
		}
		report.AssertNoViolations(t, violations)
	})
	t.Run("structure", func(t *testing.T) {
		violations := rules.CheckStructure(root, opts...)
		for _, v := range violations {
			t.Log(v.String())
		}
		report.AssertNoViolations(t, violations)
	})
	t.Run("naming", func(t *testing.T) {
		violations := rules.CheckNaming(pkgs, opts...)
		for _, v := range violations {
			t.Log(v.String())
		}
		report.AssertNoViolations(t, violations)
	})
}

func TestIntegration_Layered_Violations(t *testing.T) {
	root := t.TempDir()
	module := "example.com/layeredinvalid"
	m := rules.Layered()

	writeIntegrationFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.25.0\n")

	// repository imports handler — direction violation
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "order", "handler", "handler.go"),
		"package handler\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "order", "repository", "order.go"),
		"package repository\n\nimport _ \""+module+"/internal/domain/order/handler\"\n")
	// model imports service — direction violation
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "order", "model", "order.go"),
		"package model\n\nimport _ \""+module+"/internal/domain/order/service\"\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "order", "service", "service.go"),
		"package service\n")
	// model imports pkg — pkg-restricted violation
	writeIntegrationFile(t, filepath.Join(root, "internal", "pkg", "logger", "logger.go"),
		"package logger\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "order", "model", "with_logger.go"),
		"package model\n\nimport _ \""+module+"/internal/pkg/logger\"\n")

	pkgs, err := analyzer.Load(root, "internal/...")
	if err != nil {
		t.Log("partial load:", err)
	}

	opts := []rules.Option{rules.WithModel(m)}

	layerViolations := rules.CheckLayerDirection(pkgs, module, root, opts...)
	t.Run("detects direction violation", func(t *testing.T) {
		assertHasRule(t, layerViolations, "layer.direction")
	})
	t.Run("detects model importing pkg", func(t *testing.T) {
		assertHasRule(t, layerViolations, "layer.inner-imports-pkg")
	})
}

func TestIntegration_RealWorldHexagonal(t *testing.T) {
	root := t.TempDir()
	module := "example.com/hexshop"
	m := rules.Hexagonal()

	writeIntegrationFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.25.0\n")

	// --- domain: order ---
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "order", "handler", "http.go"),
		"package handler\n\nimport _ \""+module+"/internal/domain/order/usecase\"\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "order", "usecase", "create_order.go"),
		"package usecase\n\nimport (\n\t_ \""+module+"/internal/domain/order/port\"\n\t_ \""+module+"/internal/domain/order/domain\"\n)\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "order", "port", "repository.go"),
		"package port\n\nimport _ \""+module+"/internal/domain/order/domain\"\n\ntype OrderRepo interface{ Save(id string) error }\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "order", "domain", "order.go"),
		"package domain\n\ntype Order struct{ ID string }\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "order", "adapter", "pg_order.go"),
		"package adapter\n\nimport (\n\t_ \""+module+"/internal/domain/order/port\"\n\t_ \""+module+"/internal/domain/order/domain\"\n)\n\ntype PgOrder struct{}\n")

	// --- pkg ---
	writeIntegrationFile(t, filepath.Join(root, "internal", "pkg", "logger", "logger.go"),
		"package logger\n")

	pkgs, err := analyzer.Load(root, "internal/...")
	if err != nil {
		t.Log("partial load:", err)
	}
	if len(pkgs) == 0 {
		t.Fatal("no packages loaded")
	}

	opts := []rules.Option{rules.WithModel(m)}

	t.Run("layer direction", func(t *testing.T) {
		violations := rules.CheckLayerDirection(pkgs, module, root, opts...)
		for _, v := range violations {
			t.Log(v.String())
		}
		report.AssertNoViolations(t, violations)
	})
	t.Run("domain isolation", func(t *testing.T) {
		violations := rules.CheckDomainIsolation(pkgs, module, root, opts...)
		for _, v := range violations {
			t.Log(v.String())
		}
		report.AssertNoViolations(t, violations)
	})
	t.Run("structure", func(t *testing.T) {
		violations := rules.CheckStructure(root, opts...)
		for _, v := range violations {
			t.Log(v.String())
		}
		report.AssertNoViolations(t, violations)
	})
	t.Run("naming", func(t *testing.T) {
		violations := rules.CheckNaming(pkgs, opts...)
		for _, v := range violations {
			t.Log(v.String())
		}
		report.AssertNoViolations(t, violations)
	})
}

func TestIntegration_Hexagonal_Violations(t *testing.T) {
	root := t.TempDir()
	module := "example.com/hexinvalid"
	m := rules.Hexagonal()

	writeIntegrationFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.25.0\n")

	// domain imports usecase — direction violation
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "order", "domain", "order.go"),
		"package domain\n\nimport _ \""+module+"/internal/domain/order/usecase\"\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "order", "usecase", "uc.go"),
		"package usecase\n")
	// adapter imports handler — direction violation
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "order", "adapter", "bad.go"),
		"package adapter\n\nimport _ \""+module+"/internal/domain/order/handler\"\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "order", "handler", "handler.go"),
		"package handler\n")
	// domain imports pkg — pkg-restricted violation
	writeIntegrationFile(t, filepath.Join(root, "internal", "pkg", "logger", "logger.go"),
		"package logger\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "order", "domain", "with_logger.go"),
		"package domain\n\nimport _ \""+module+"/internal/pkg/logger\"\n")

	pkgs, err := analyzer.Load(root, "internal/...")
	if err != nil {
		t.Log("partial load:", err)
	}

	opts := []rules.Option{rules.WithModel(m)}

	layerViolations := rules.CheckLayerDirection(pkgs, module, root, opts...)
	t.Run("detects direction violation", func(t *testing.T) {
		assertHasRule(t, layerViolations, "layer.direction")
	})
	t.Run("detects domain importing pkg", func(t *testing.T) {
		assertHasRule(t, layerViolations, "layer.inner-imports-pkg")
	})
}

func TestIntegration_RealWorldModularMonolith(t *testing.T) {
	root := t.TempDir()
	module := "example.com/modshop"
	m := rules.ModularMonolith()

	writeIntegrationFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.25.0\n")

	// --- domain: order ---
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "order", "api", "handler.go"),
		"package api\n\nimport _ \""+module+"/internal/domain/order/application\"\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "order", "application", "create_order.go"),
		"package application\n\nimport _ \""+module+"/internal/domain/order/core\"\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "order", "core", "order.go"),
		"package core\n\ntype Order struct{ ID string }\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "order", "infrastructure", "pg_order.go"),
		"package infrastructure\n\nimport _ \""+module+"/internal/domain/order/core\"\n\ntype PgOrder struct{}\n")

	// --- pkg ---
	writeIntegrationFile(t, filepath.Join(root, "internal", "pkg", "logger", "logger.go"),
		"package logger\n")

	pkgs, err := analyzer.Load(root, "internal/...")
	if err != nil {
		t.Log("partial load:", err)
	}
	if len(pkgs) == 0 {
		t.Fatal("no packages loaded")
	}

	opts := []rules.Option{rules.WithModel(m)}

	t.Run("layer direction", func(t *testing.T) {
		violations := rules.CheckLayerDirection(pkgs, module, root, opts...)
		for _, v := range violations {
			t.Log(v.String())
		}
		report.AssertNoViolations(t, violations)
	})
	t.Run("domain isolation", func(t *testing.T) {
		violations := rules.CheckDomainIsolation(pkgs, module, root, opts...)
		for _, v := range violations {
			t.Log(v.String())
		}
		report.AssertNoViolations(t, violations)
	})
	t.Run("structure", func(t *testing.T) {
		violations := rules.CheckStructure(root, opts...)
		for _, v := range violations {
			t.Log(v.String())
		}
		report.AssertNoViolations(t, violations)
	})
	t.Run("naming", func(t *testing.T) {
		violations := rules.CheckNaming(pkgs, opts...)
		for _, v := range violations {
			t.Log(v.String())
		}
		report.AssertNoViolations(t, violations)
	})
}

func TestIntegration_ModularMonolith_Violations(t *testing.T) {
	root := t.TempDir()
	module := "example.com/modinvalid"
	m := rules.ModularMonolith()

	writeIntegrationFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.25.0\n")

	// core imports api — direction violation
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "order", "core", "order.go"),
		"package core\n\nimport _ \""+module+"/internal/domain/order/api\"\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "order", "api", "handler.go"),
		"package api\n")
	// infrastructure imports application — direction violation
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "order", "infrastructure", "bad.go"),
		"package infrastructure\n\nimport _ \""+module+"/internal/domain/order/application\"\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "order", "application", "service.go"),
		"package application\n")
	// cross-domain: order imports user
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "order", "application", "cross.go"),
		"package application\n\nimport _ \""+module+"/internal/domain/user/core\"\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "user", "core", "user.go"),
		"package core\n")

	pkgs, err := analyzer.Load(root, "internal/...")
	if err != nil {
		t.Log("partial load:", err)
	}

	opts := []rules.Option{rules.WithModel(m)}

	layerViolations := rules.CheckLayerDirection(pkgs, module, root, opts...)
	t.Run("detects direction violation", func(t *testing.T) {
		assertHasRule(t, layerViolations, "layer.direction")
	})

	isolationViolations := rules.CheckDomainIsolation(pkgs, module, root, opts...)
	t.Run("detects cross-domain import", func(t *testing.T) {
		assertHasRule(t, isolationViolations, "isolation.cross-domain")
	})
}

func TestIntegration_ConsumerWorker_Valid(t *testing.T) {
	root := t.TempDir()
	module := "example.com/consumerworker"
	m := rules.ConsumerWorker()

	writeIntegrationFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.21\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "worker", "worker_order.go"),
		"package worker\n\nimport (\n\t\"context\"\n\t_ \""+module+"/internal/service\"\n)\n\ntype OrderWorker struct{}\nfunc (w *OrderWorker) Process(ctx context.Context) error { return nil }\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "worker", "worker_payment.go"),
		"package worker\n\nimport (\n\t\"context\"\n\t_ \""+module+"/internal/model\"\n)\n\ntype PaymentWorker struct{}\nfunc (w *PaymentWorker) Process(ctx context.Context) error { return nil }\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "service", "order.go"),
		"package service\n\nimport _ \""+module+"/internal/store\"\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "store", "order.go"),
		"package store\n\nimport _ \""+module+"/internal/model\"\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "model", "order.go"),
		"package model\n\ntype Order struct{ ID string }\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "pkg", "consumer", "consumer.go"),
		"package consumer\n")

	pkgs, err := analyzer.Load(root, "internal/...")
	if err != nil {
		t.Log("partial load:", err)
	}
	if len(pkgs) == 0 {
		t.Fatal("no packages loaded")
	}

	opts := []rules.Option{rules.WithModel(m)}

	t.Run("layer direction", func(t *testing.T) {
		report.AssertNoViolations(t, rules.CheckLayerDirection(pkgs, module, root, opts...))
	})
	t.Run("domain isolation skipped", func(t *testing.T) {
		report.AssertNoViolations(t, rules.CheckDomainIsolation(pkgs, module, root, opts...))
	})
	t.Run("structure", func(t *testing.T) {
		report.AssertNoViolations(t, rules.CheckStructure(root, opts...))
	})
	t.Run("naming", func(t *testing.T) {
		report.AssertNoViolations(t, rules.CheckNaming(pkgs, opts...))
	})
	t.Run("type patterns", func(t *testing.T) {
		report.AssertNoViolations(t, rules.CheckTypePatterns(pkgs, opts...))
	})
	t.Run("run all", func(t *testing.T) {
		report.AssertNoViolations(t, rules.RunAll(pkgs, module, root, opts...))
	})
}

func TestIntegration_ConsumerWorker_Violations(t *testing.T) {
	root := t.TempDir()
	module := "example.com/cw-violations"
	m := rules.ConsumerWorker()

	writeIntegrationFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.21\n")
	// store imports worker — layer direction violation
	writeIntegrationFile(t, filepath.Join(root, "internal", "store", "bad.go"),
		"package store\n\nimport _ \""+module+"/internal/worker\"\n")
	// model imports pkg — inner-imports-pkg violation
	writeIntegrationFile(t, filepath.Join(root, "internal", "model", "bad.go"),
		"package model\n\nimport _ \""+module+"/internal/pkg/consumer\"\n")
	// worker file defines BadName instead of OrderWorker — type mismatch
	writeIntegrationFile(t, filepath.Join(root, "internal", "worker", "worker_order.go"),
		"package worker\n\nimport \"context\"\n\ntype BadName struct{}\nfunc (w *BadName) Process(ctx context.Context) error { return nil }\n")
	// unexpected top-level package
	writeIntegrationFile(t, filepath.Join(root, "internal", "random", "stuff.go"),
		"package random\n")
	// needed for imports to resolve
	writeIntegrationFile(t, filepath.Join(root, "internal", "worker", "worker_payment.go"),
		"package worker\n\nimport \"context\"\n\ntype PaymentWorker struct{}\nfunc (w *PaymentWorker) Process(ctx context.Context) error { return nil }\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "pkg", "consumer", "consumer.go"),
		"package consumer\n")

	pkgs, err := analyzer.Load(root, "internal/...")
	if err != nil {
		t.Log("partial load:", err)
	}

	violations := rules.RunAll(pkgs, module, root, rules.WithModel(m))
	if len(violations) == 0 {
		t.Fatal("expected violations")
	}

	assertHasRule(t, violations, "layer.direction")
	assertHasRule(t, violations, "layer.inner-imports-pkg")
	assertHasRule(t, violations, "naming.type-pattern-mismatch")
	assertHasRule(t, violations, "structure.internal-top-level")
}

func TestIntegration_Batch_Valid(t *testing.T) {
	root := t.TempDir()
	module := "example.com/batchapp"
	m := rules.Batch()

	writeIntegrationFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.21\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "job", "job_expire.go"),
		"package job\n\nimport (\n\t\"context\"\n\t_ \""+module+"/internal/service\"\n)\n\ntype ExpireJob struct{}\nfunc (j *ExpireJob) Run(ctx context.Context) error { return nil }\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "job", "job_cleanup.go"),
		"package job\n\nimport (\n\t\"context\"\n\t_ \""+module+"/internal/model\"\n)\n\ntype CleanupJob struct{}\nfunc (j *CleanupJob) Run(ctx context.Context) error { return nil }\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "service", "file.go"),
		"package service\n\nimport _ \""+module+"/internal/store\"\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "store", "file.go"),
		"package store\n\nimport _ \""+module+"/internal/model\"\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "model", "file.go"),
		"package model\n\ntype File struct{ ID string }\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "pkg", "batchutil", "util.go"),
		"package batchutil\n")

	pkgs, err := analyzer.Load(root, "internal/...")
	if err != nil {
		t.Log("partial load:", err)
	}
	if len(pkgs) == 0 {
		t.Fatal("no packages loaded")
	}

	opts := []rules.Option{rules.WithModel(m)}

	t.Run("layer direction", func(t *testing.T) {
		report.AssertNoViolations(t, rules.CheckLayerDirection(pkgs, module, root, opts...))
	})
	t.Run("domain isolation skipped", func(t *testing.T) {
		report.AssertNoViolations(t, rules.CheckDomainIsolation(pkgs, module, root, opts...))
	})
	t.Run("structure", func(t *testing.T) {
		report.AssertNoViolations(t, rules.CheckStructure(root, opts...))
	})
	t.Run("naming", func(t *testing.T) {
		report.AssertNoViolations(t, rules.CheckNaming(pkgs, opts...))
	})
	t.Run("type patterns", func(t *testing.T) {
		report.AssertNoViolations(t, rules.CheckTypePatterns(pkgs, opts...))
	})
	t.Run("run all", func(t *testing.T) {
		report.AssertNoViolations(t, rules.RunAll(pkgs, module, root, opts...))
	})
}

func TestIntegration_Batch_Violations(t *testing.T) {
	root := t.TempDir()
	module := "example.com/batch-violations"
	m := rules.Batch()

	writeIntegrationFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.21\n")
	// store imports job — layer direction violation
	writeIntegrationFile(t, filepath.Join(root, "internal", "store", "bad.go"),
		"package store\n\nimport _ \""+module+"/internal/job\"\n")
	// model imports pkg — inner-imports-pkg violation
	writeIntegrationFile(t, filepath.Join(root, "internal", "model", "bad.go"),
		"package model\n\nimport _ \""+module+"/internal/pkg/batchutil\"\n")
	// job file defines BadName instead of SyncJob — type mismatch
	writeIntegrationFile(t, filepath.Join(root, "internal", "job", "job_sync.go"),
		"package job\n\nimport \"context\"\n\ntype BadName struct{}\nfunc (j *BadName) Run(ctx context.Context) error { return nil }\n")
	// unexpected top-level package
	writeIntegrationFile(t, filepath.Join(root, "internal", "random", "stuff.go"),
		"package random\n")
	// needed for imports to resolve
	writeIntegrationFile(t, filepath.Join(root, "internal", "job", "job_cleanup.go"),
		"package job\n\nimport \"context\"\n\ntype CleanupJob struct{}\nfunc (j *CleanupJob) Run(ctx context.Context) error { return nil }\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "pkg", "batchutil", "util.go"),
		"package batchutil\n")

	pkgs, err := analyzer.Load(root, "internal/...")
	if err != nil {
		t.Log("partial load:", err)
	}

	violations := rules.RunAll(pkgs, module, root, rules.WithModel(m))
	if len(violations) == 0 {
		t.Fatal("expected violations")
	}

	assertHasRule(t, violations, "layer.direction")
	assertHasRule(t, violations, "layer.inner-imports-pkg")
	assertHasRule(t, violations, "naming.type-pattern-mismatch")
	assertHasRule(t, violations, "structure.internal-top-level")
}

func TestIntegration_EventPipeline_Valid(t *testing.T) {
	root := t.TempDir()
	module := "example.com/eventapp"
	m := rules.EventPipeline()

	writeIntegrationFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.21\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "command", "command_create_order.go"),
		"package command\n\nimport (\n\t\"context\"\n\t_ \""+module+"/internal/aggregate\"\n\t_ \""+module+"/internal/eventstore\"\n)\n\ntype CreateOrderCommand struct{}\nfunc (c *CreateOrderCommand) Execute(ctx context.Context) error { return nil }\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "aggregate", "aggregate_order.go"),
		"package aggregate\n\nimport (\n\t\"context\"\n\t_ \""+module+"/internal/event\"\n)\n\ntype OrderAggregate struct{}\nfunc (a *OrderAggregate) Apply(ctx context.Context) error { return nil }\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "event", "order_created.go"),
		"package event\n\nimport _ \""+module+"/internal/model\"\n\ntype OrderCreated struct{ ID string }\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "projection", "order_view.go"),
		"package projection\n\nimport (\n\t_ \""+module+"/internal/event\"\n\t_ \""+module+"/internal/readstore\"\n)\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "eventstore", "pg.go"),
		"package eventstore\n\nimport (\n\t_ \""+module+"/internal/event\"\n\t_ \""+module+"/internal/model\"\n)\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "readstore", "pg.go"),
		"package readstore\n\nimport _ \""+module+"/internal/model\"\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "model", "order.go"),
		"package model\n\ntype Order struct{ ID string }\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "pkg", "eventbus", "bus.go"),
		"package eventbus\n")

	pkgs, err := analyzer.Load(root, "internal/...")
	if err != nil {
		t.Log("partial load:", err)
	}
	if len(pkgs) == 0 {
		t.Fatal("no packages loaded")
	}

	opts := []rules.Option{rules.WithModel(m)}

	t.Run("layer direction", func(t *testing.T) {
		report.AssertNoViolations(t, rules.CheckLayerDirection(pkgs, module, root, opts...))
	})
	t.Run("domain isolation skipped", func(t *testing.T) {
		report.AssertNoViolations(t, rules.CheckDomainIsolation(pkgs, module, root, opts...))
	})
	t.Run("structure", func(t *testing.T) {
		report.AssertNoViolations(t, rules.CheckStructure(root, opts...))
	})
	t.Run("naming", func(t *testing.T) {
		report.AssertNoViolations(t, rules.CheckNaming(pkgs, opts...))
	})
	t.Run("type patterns", func(t *testing.T) {
		report.AssertNoViolations(t, rules.CheckTypePatterns(pkgs, opts...))
	})
	t.Run("run all", func(t *testing.T) {
		report.AssertNoViolations(t, rules.RunAll(pkgs, module, root, opts...))
	})
}

func TestIntegration_EventPipeline_Violations(t *testing.T) {
	root := t.TempDir()
	module := "example.com/ep-violations"
	m := rules.EventPipeline()

	writeIntegrationFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.21\n")
	// readstore imports command — layer direction violation
	writeIntegrationFile(t, filepath.Join(root, "internal", "readstore", "bad.go"),
		"package readstore\n\nimport _ \""+module+"/internal/command\"\n")
	// event imports pkg — inner-imports-pkg violation
	writeIntegrationFile(t, filepath.Join(root, "internal", "event", "bad.go"),
		"package event\n\nimport _ \""+module+"/internal/pkg/eventbus\"\n")
	// command file defines BadName instead of ShipCommand — type mismatch
	writeIntegrationFile(t, filepath.Join(root, "internal", "command", "command_ship.go"),
		"package command\n\nimport \"context\"\n\ntype BadName struct{}\nfunc (c *BadName) Execute(ctx context.Context) error { return nil }\n")
	// aggregate file defines OrderAggregate but no Apply method — missing process
	writeIntegrationFile(t, filepath.Join(root, "internal", "aggregate", "aggregate_order.go"),
		"package aggregate\n\ntype OrderAggregate struct{}\n")
	// unexpected top-level package
	writeIntegrationFile(t, filepath.Join(root, "internal", "random", "stuff.go"),
		"package random\n")
	// needed for imports to resolve
	writeIntegrationFile(t, filepath.Join(root, "internal", "pkg", "eventbus", "bus.go"),
		"package eventbus\n")

	pkgs, err := analyzer.Load(root, "internal/...")
	if err != nil {
		t.Log("partial load:", err)
	}

	violations := rules.RunAll(pkgs, module, root, rules.WithModel(m))
	if len(violations) == 0 {
		t.Fatal("expected violations")
	}

	assertHasRule(t, violations, "layer.direction")
	assertHasRule(t, violations, "layer.inner-imports-pkg")
	assertHasRule(t, violations, "naming.type-pattern-mismatch")
	assertHasRule(t, violations, "naming.type-pattern-missing-method")
	assertHasRule(t, violations, "structure.internal-top-level")
}

func writeIntegrationFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
