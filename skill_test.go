package goarchguard_test

// skill_test.go verifies that the patterns described in the go-arch-guard
// SKILL.md actually work. An AI agent following the skill should be able to
// set up architecture guards on any Go server project using these patterns.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/analyzer"
	"github.com/NamhaeSusan/go-arch-guard/report"
	"github.com/NamhaeSusan/go-arch-guard/rules"
)

// TestSkill_NewProjectSetup simulates the "new project from scratch" scenario
// from the skill. A correctly structured project must pass all four rules.
func TestSkill_NewProjectSetup(t *testing.T) {
	root := t.TempDir()
	mod := "example.com/myserver"

	files := map[string]string{
		"go.mod": "module " + mod + "\n\ngo 1.26.1\n",

		// cmd
		"cmd/api/main.go": `package main

import _ "` + mod + `/internal/domain/order"

func main() {}
`,
		// domain: order
		"internal/domain/order/alias.go": `package order

import "` + mod + `/internal/domain/order/app"

type Service = app.Service
`,
		"internal/domain/order/core/model/order.go": "package model\n\ntype Order struct{ ID int64 }\n",
		"internal/domain/order/core/repo/repository.go": `package repo

import "` + mod + `/internal/domain/order/core/model"

type Repository interface {
	FindByID(id int64) (*model.Order, error)
}
`,
		"internal/domain/order/core/svc/order.go": `package svc

import "` + mod + `/internal/domain/order/core/model"

func Validate(o *model.Order) bool { return o.ID > 0 }
`,
		"internal/domain/order/event/events.go": `package event

import "` + mod + `/internal/domain/order/core/model"

type Created struct{ Order model.Order }
`,
		"internal/domain/order/app/service.go": `package app

import (
	"` + mod + `/internal/domain/order/core/model"
	"` + mod + `/internal/domain/order/core/repo"
)

type Service struct{ repo repo.Repository }

func (s *Service) Get(id int64) (*model.Order, error) {
	return s.repo.FindByID(id)
}
`,
		"internal/domain/order/handler/http/handler.go": `package http

import "` + mod + `/internal/domain/order/app"

type Handler struct{ svc *app.Service }
`,
		"internal/domain/order/infra/persistence/store.go": `package persistence

import (
	"` + mod + `/internal/domain/order/core/model"
	"` + mod + `/internal/domain/order/core/repo"
)

var _ repo.Repository = (*Store)(nil)

type Store struct{}

func (s *Store) FindByID(id int64) (*model.Order, error) { return nil, nil }
`,
		// orchestration
		"internal/orchestration/create_order.go": `package orchestration

import _ "` + mod + `/internal/domain/order"
`,
		// pkg
		"internal/pkg/middleware.go": "package pkg\n",
	}

	writeProjectFiles(t, root, files)

	pkgs, err := analyzer.Load(root, "internal/...", "cmd/...")
	if err != nil {
		t.Log(err)
	}
	if len(pkgs) == 0 {
		t.Fatalf("no packages loaded: %v", err)
	}

	t.Run("domain isolation", func(t *testing.T) {
		report.AssertNoViolations(t, rules.CheckDomainIsolation(pkgs, "", ""))
	})
	t.Run("layer direction", func(t *testing.T) {
		report.AssertNoViolations(t, rules.CheckLayerDirection(pkgs, "", ""))
	})
	t.Run("naming", func(t *testing.T) {
		report.AssertNoViolations(t, rules.CheckNaming(pkgs))
	})
	t.Run("structure", func(t *testing.T) {
		report.AssertNoViolations(t, rules.CheckStructure(root))
	})
}

// TestSkill_AutoExtractModuleRoot verifies the skill's simplified usage:
// passing "" for module and root auto-extracts them from loaded packages.
func TestSkill_AutoExtractModuleRoot(t *testing.T) {
	root := t.TempDir()
	mod := "example.com/autoextract"

	files := map[string]string{
		"go.mod":                                         "module " + mod + "\n\ngo 1.26.1\n",
		"internal/domain/user/alias.go":                  "package user\n",
		"internal/domain/user/core/model/user.go":        "package model\n\ntype User struct{ ID int64 }\n",
		"internal/domain/user/app/service.go":            "package app\n",
		"internal/domain/user/handler/http/handler.go":   "package http\n",
		"internal/domain/user/infra/persistence/store.go": "package persistence\n",
	}

	writeProjectFiles(t, root, files)

	pkgs, err := analyzer.Load(root, "internal/...")
	if err != nil {
		t.Log(err)
	}
	if len(pkgs) == 0 {
		t.Fatalf("no packages loaded: %v", err)
	}

	// "" triggers auto-extraction — same as skill template
	violations := rules.CheckDomainIsolation(pkgs, "", "")
	report.AssertNoViolations(t, violations)

	violations = rules.CheckLayerDirection(pkgs, "", "")
	report.AssertNoViolations(t, violations)
}

// TestSkill_ExcludeOption verifies the migration scenario from the skill:
// WithExclude suppresses violations for paths being migrated.
func TestSkill_ExcludeOption(t *testing.T) {
	root := t.TempDir()
	mod := "example.com/migration"

	files := map[string]string{
		"go.mod": "module " + mod + "\n\ngo 1.26.1\n",

		"internal/domain/order/alias.go":                  "package order\n",
		"internal/domain/order/core/model/order.go":       "package model\n\ntype Order struct{}\n",
		"internal/domain/order/app/service.go":            "package app\n",
		"internal/domain/order/handler/http/handler.go":   "package http\n",
		"internal/domain/order/infra/persistence/store.go": "package persistence\n",

		// legacy — violates structure.internal-top-level
		"internal/legacy/old.go": "package legacy\n",
	}

	writeProjectFiles(t, root, files)

	// Without exclude: should have violation
	violations := rules.CheckStructure(root)
	hasTopLevel := false
	for _, v := range violations {
		if v.Rule == "structure.internal-top-level" {
			hasTopLevel = true
			break
		}
	}
	if !hasTopLevel {
		t.Fatal("expected structure.internal-top-level violation without exclude")
	}

	// With exclude: structure check on legacy path excluded
	structExcluded := rules.CheckStructure(root, rules.WithExclude("internal/legacy/..."))
	for _, v := range structExcluded {
		if v.Rule == "structure.internal-top-level" && strings.Contains(v.File, "legacy") {
			t.Error("expected legacy to be excluded from structure check")
		}
	}

	pkgs, err := analyzer.Load(root, "internal/...")
	if err != nil {
		t.Log(err)
	}

	isolationViolations := rules.CheckDomainIsolation(pkgs, "", "",
		rules.WithExclude("internal/legacy/..."),
	)
	report.AssertNoViolations(t, isolationViolations)
}

// TestSkill_WarningMode verifies that WithSeverity(Warning) downgrades
// violations so AssertNoViolations passes even with violations present.
func TestSkill_WarningMode(t *testing.T) {
	root := t.TempDir()
	mod := "example.com/warnmode"

	files := map[string]string{
		"go.mod": "module " + mod + "\n\ngo 1.26.1\n",

		"internal/domain/order/alias.go":            "package order\n",
		"internal/domain/order/core/model/order.go": "package model\n\ntype Order struct{}\n",
		"internal/domain/order/app/service.go": `package app

import _ "` + mod + `/internal/domain/user/app"
`,
		"internal/domain/user/alias.go":            "package user\n",
		"internal/domain/user/core/model/user.go":  "package model\n\ntype User struct{}\n",
		"internal/domain/user/app/service.go":      "package app\n",
	}

	writeProjectFiles(t, root, files)

	pkgs, err := analyzer.Load(root, "internal/...")
	if err != nil {
		t.Log(err)
	}
	if len(pkgs) == 0 {
		t.Fatalf("no packages loaded: %v", err)
	}

	// Error mode: should have violations
	errViolations := rules.CheckDomainIsolation(pkgs, "", "")
	if len(errViolations) == 0 {
		t.Fatal("expected isolation violations in error mode")
	}

	// Warning mode: same violations but AssertNoViolations passes
	warnViolations := rules.CheckDomainIsolation(pkgs, "", "",
		rules.WithSeverity(rules.Warning),
	)
	if len(warnViolations) == 0 {
		t.Fatal("expected violations in warning mode too")
	}
	for _, v := range warnViolations {
		if v.Severity != rules.Warning {
			t.Errorf("expected Warning severity, got %v", v.Severity)
		}
	}
	report.AssertNoViolations(t, warnViolations) // must pass
}

// TestSkill_BannedPatterns verifies the banned patterns table from the skill.
func TestSkill_BannedPatterns(t *testing.T) {
	root := t.TempDir()
	mod := "example.com/banned"

	files := map[string]string{
		"go.mod": "module " + mod + "\n\ngo 1.26.1\n",

		"internal/domain/order/alias.go":            "package order\n",
		"internal/domain/order/core/model/order.go": "package model\n\ntype Order struct{}\n",

		// banned package names
		"internal/domain/order/app/util/helper.go": "package util\n",
	}

	writeProjectFiles(t, root, files)

	violations := rules.CheckStructure(root)
	hasBanned := false
	for _, v := range violations {
		if v.Rule == "structure.banned-package" {
			hasBanned = true
			break
		}
	}
	if !hasBanned {
		t.Error("expected structure.banned-package violation for 'util' package")
	}
}

// TestSkill_CrossDomainViolation verifies that the isolation rule catches
// the most common vibe-coding mistake: importing another domain directly.
func TestSkill_CrossDomainViolation(t *testing.T) {
	root := t.TempDir()
	mod := "example.com/crossdomain"

	files := map[string]string{
		"go.mod": "module " + mod + "\n\ngo 1.26.1\n",

		"internal/domain/order/alias.go":            "package order\n",
		"internal/domain/order/core/model/order.go": "package model\n\ntype Order struct{}\n",
		"internal/domain/order/app/service.go": `package app

import _ "` + mod + `/internal/domain/user"
`,
		"internal/domain/user/alias.go":           "package user\n",
		"internal/domain/user/core/model/user.go": "package model\n\ntype User struct{}\n",
	}

	writeProjectFiles(t, root, files)

	pkgs, err := analyzer.Load(root, "internal/...")
	if err != nil {
		t.Log(err)
	}
	if len(pkgs) == 0 {
		t.Fatalf("no packages loaded: %v", err)
	}

	violations := rules.CheckDomainIsolation(pkgs, "", "")
	hasCross := false
	for _, v := range violations {
		if v.Rule == "isolation.cross-domain" {
			hasCross = true
			break
		}
	}
	if !hasCross {
		t.Error("expected isolation.cross-domain violation")
	}
}

// TestSkill_LayerDirectionViolation verifies that wrong-direction imports
// within a domain are caught.
func TestSkill_LayerDirectionViolation(t *testing.T) {
	root := t.TempDir()
	mod := "example.com/layerdir"

	files := map[string]string{
		"go.mod": "module " + mod + "\n\ngo 1.26.1\n",

		"internal/domain/order/alias.go":            "package order\n",
		"internal/domain/order/core/model/order.go": "package model\n\ntype Order struct{}\n",
		// core/model importing app — wrong direction
		"internal/domain/order/core/svc/bad.go": `package svc

import _ "` + mod + `/internal/domain/order/app"
`,
		"internal/domain/order/app/service.go": "package app\n",
	}

	writeProjectFiles(t, root, files)

	pkgs, err := analyzer.Load(root, "internal/...")
	if err != nil {
		t.Log(err)
	}

	violations := rules.CheckLayerDirection(pkgs, "", "")
	hasDirection := false
	for _, v := range violations {
		if v.Rule == "layer.direction" {
			hasDirection = true
			break
		}
	}
	if !hasDirection {
		t.Error("expected layer.direction violation for core/svc importing app")
	}
}

func writeProjectFiles(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for rel, content := range files {
		abs := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}
