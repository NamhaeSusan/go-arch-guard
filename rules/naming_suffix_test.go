package rules_test

import (
	"path/filepath"
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/analyzer"
	"github.com/NamhaeSusan/go-arch-guard/rules"
)

func TestCheckNaming_EventPipeline_CommandSuffix(t *testing.T) {
	root := t.TempDir()
	module := "example.com/evpipe-naming"
	m := rules.EventPipeline()

	writeTestFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.21\n")
	// order_command.go in command/ must trigger no-layer-suffix
	writeTestFile(t, filepath.Join(root, "internal", "command", "order_command.go"),
		"package command\n\ntype PlaceOrder struct{}\n")

	pkgs, err := analyzer.Load(root, "internal/...")
	if err != nil {
		t.Fatal(err)
	}
	violations := rules.CheckNaming(pkgs, rules.WithModel(m))

	found := false
	for _, v := range violations {
		if v.Rule == "naming.no-layer-suffix" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected naming.no-layer-suffix violation for order_command.go in EventPipeline preset")
		for _, v := range violations {
			t.Log(v.String())
		}
	}
}

func TestCheckNaming_ModularMonolith_ApplicationSuffix(t *testing.T) {
	root := t.TempDir()
	module := "example.com/modmonolith-naming"
	m := rules.ModularMonolith()

	writeTestFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.21\n")
	// order_application.go in application/ must trigger no-layer-suffix
	writeTestFile(t, filepath.Join(root, "internal", "domain", "order", "application", "order_application.go"),
		"package application\n\ntype OrderUsecase struct{}\n")

	pkgs, err := analyzer.Load(root, "internal/...")
	if err != nil {
		t.Fatal(err)
	}
	violations := rules.CheckNaming(pkgs, rules.WithModel(m))

	found := false
	for _, v := range violations {
		if v.Rule == "naming.no-layer-suffix" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected naming.no-layer-suffix violation for order_application.go in ModularMonolith preset")
		for _, v := range violations {
			t.Log(v.String())
		}
	}
}

func TestCheckNaming_CustomModel_UsecaseSuffix(t *testing.T) {
	root := t.TempDir()
	module := "example.com/custom-naming"

	// Custom model with sublayer "usecase" that isn't in DDD's LayerDirNames.
	m := rules.NewModel(
		rules.WithDomainDir(""),
		rules.WithOrchestrationDir(""),
		rules.WithSublayers([]string{"usecase", "gateway", "model"}),
		rules.WithDirection(map[string][]string{
			"usecase": {"gateway", "model"},
			"gateway": {"model"},
			"model":   {},
		}),
		rules.WithPkgRestricted(map[string]bool{"model": true}),
	)

	writeTestFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.21\n")
	// foo_usecase.go in usecase/ must trigger no-layer-suffix even though
	// DDD's LayerDirNames does not include "usecase".
	writeTestFile(t, filepath.Join(root, "internal", "usecase", "foo_usecase.go"),
		"package usecase\n\ntype Foo struct{}\n")

	pkgs, err := analyzer.Load(root, "internal/...")
	if err != nil {
		t.Fatal(err)
	}
	violations := rules.CheckNaming(pkgs, rules.WithModel(m))

	found := false
	for _, v := range violations {
		if v.Rule == "naming.no-layer-suffix" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected naming.no-layer-suffix violation for foo_usecase.go in custom model with WithSublayers")
		for _, v := range violations {
			t.Log(v.String())
		}
	}
}

func TestCheckNaming_DDD_HandlerSuffix_StillWorks(t *testing.T) {
	root := t.TempDir()
	module := "example.com/ddd-naming"
	m := rules.DDD()

	writeTestFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.21\n")
	// order_handler.go in handler/ must trigger no-layer-suffix
	writeTestFile(t, filepath.Join(root, "internal", "domain", "order", "handler", "order_handler.go"),
		"package handler\n\ntype HTTP struct{}\n")

	pkgs, err := analyzer.Load(root, "internal/...")
	if err != nil {
		t.Fatal(err)
	}
	violations := rules.CheckNaming(pkgs, rules.WithModel(m))

	found := false
	for _, v := range violations {
		if v.Rule == "naming.no-layer-suffix" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected naming.no-layer-suffix violation for order_handler.go in DDD preset")
		for _, v := range violations {
			t.Log(v.String())
		}
	}
}
