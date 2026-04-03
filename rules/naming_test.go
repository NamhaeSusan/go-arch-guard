package rules_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/analyzer"
	"github.com/NamhaeSusan/go-arch-guard/rules"
)

func TestCheckNaming(t *testing.T) {
	t.Run("valid project has no violations", func(t *testing.T) {
		pkgs := loadValid(t)
		violations := rules.CheckNaming(pkgs)
		if len(violations) > 0 {
			for _, v := range violations {
				t.Log(v.String())
			}
			t.Errorf("expected no violations, got %d", len(violations))
		}
	})

	t.Run("detects interface outside core/repo in domain", func(t *testing.T) {
		pkgs := loadInvalid(t)
		violations := rules.CheckNaming(pkgs)
		// handler: Service + auditLogger (direct), app: AdminOps (direct) + OrderRepo (alias)
		wantIfaces := map[string]bool{"Service": false, "auditLogger": false, "AdminOps": false, "OrderRepo": false}
		for _, v := range violations {
			if v.Rule != "structure.interface-placement" {
				continue
			}
			for name := range wantIfaces {
				if strings.Contains(v.Message, `"`+name+`"`) {
					wantIfaces[name] = true
				}
			}
		}
		for name, found := range wantIfaces {
			if !found {
				t.Errorf("expected structure.interface-placement violation for %s", name)
			}
		}
	})

	t.Run("reports relative file paths", func(t *testing.T) {
		pkgs := loadInvalid(t)
		violations := rules.CheckNaming(pkgs)
		if len(violations) == 0 {
			t.Fatal("expected naming violations")
		}
		for _, v := range violations {
			if filepath.IsAbs(v.File) {
				t.Fatalf("expected relative path, got %q", v.File)
			}
		}
	})

	t.Run("detects non-snake-case filename and suggests fix", func(t *testing.T) {
		root := t.TempDir()
		writeTestFile(t, filepath.Join(root, "internal", "domain", "order", "app", "createOrder.go"),
			"package app\n\ntype Request struct{}\n")
		writeTestFile(t, filepath.Join(root, "go.mod"), "module example.com/snaketest\n\ngo 1.25.0\n")

		pkgs, err := analyzer.Load(root, "internal/...")
		if err != nil {
			t.Fatal(err)
		}
		violations := rules.CheckNaming(pkgs)
		found := false
		for _, v := range violations {
			if v.Rule == "naming.snake-case-file" {
				found = true
				if !strings.Contains(v.Fix, "create_order.go") {
					t.Errorf("expected fix to suggest snake_case name, got %q", v.Fix)
				}
				break
			}
		}
		if !found {
			t.Error("expected snake-case-file violation for createOrder.go")
		}
	})

	t.Run("project-relative exclude skips matching files", func(t *testing.T) {
		pkgs := loadInvalid(t)
		violations := rules.CheckNaming(pkgs, rules.WithExclude("internal/domain/order/handler/..."))
		for _, v := range violations {
			if v.File == "internal/domain/order/handler/http/bad_handler.go" {
				t.Fatalf("expected order handler naming violations to be excluded, got %s", v.String())
			}
		}
	})

	t.Run("detects hand-rolled mock in test file", func(t *testing.T) {
		pkgs := loadInvalid(t)
		violations := rules.CheckNaming(pkgs)
		wantMocks := map[string]bool{"mockOrderRepo": false, "fakeNotifier": false}
		for _, v := range violations {
			if v.Rule != "testing.no-handmock" {
				continue
			}
			for name := range wantMocks {
				if strings.Contains(v.Message, name) {
					wantMocks[name] = true
				}
			}
		}
		for name, found := range wantMocks {
			if !found {
				t.Errorf("expected testing.no-handmock violation for %s", name)
			}
		}
	})

	t.Run("valid project has no handmock violations", func(t *testing.T) {
		pkgs := loadValid(t)
		violations := rules.CheckNaming(pkgs)
		for _, v := range violations {
			if v.Rule == "testing.no-handmock" {
				t.Errorf("unexpected testing.no-handmock violation: %s", v.String())
			}
		}
	})

	t.Run("exclude skips handmock check", func(t *testing.T) {
		pkgs := loadInvalid(t)
		violations := rules.CheckNaming(pkgs, rules.WithExclude("internal/domain/order/app/..."))
		for _, v := range violations {
			if v.Rule == "testing.no-handmock" {
				t.Errorf("expected handmock violation to be excluded, got %s", v.String())
			}
		}
	})
}

func TestCheckNaming_RepoFileInterface(t *testing.T) {
	root := t.TempDir()
	module := "example.com/repo-iface"

	// Use DDD model (has repo sublayer)
	writeTestFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.21\n")
	// order.go in repo/ must define Order interface — but doesn't
	writeTestFile(t, filepath.Join(root, "internal", "domain", "order", "core", "repo", "order.go"),
		"package repo\n\nfunc FindOrder() {}\n")
	// Need alias.go for DDD
	writeTestFile(t, filepath.Join(root, "internal", "domain", "order", "alias.go"),
		"package order\n")
	writeTestFile(t, filepath.Join(root, "internal", "domain", "order", "core", "model", "order.go"),
		"package model\n\ntype Order struct{}\n")

	pkgs := loadTestPackages(t, root)
	violations := rules.CheckNaming(pkgs) // default model = DDD
	found := false
	for _, v := range violations {
		if v.Rule == "structure.repo-file-interface" {
			found = true
		}
	}
	if !found {
		t.Error("expected structure.repo-file-interface violation for repo/order.go without Order interface")
	}
}

func TestCheckNaming_FlatLayout_WorkerFileAllowed(t *testing.T) {
	root := t.TempDir()
	module := "example.com/workernaming"
	m := rules.ConsumerWorker()

	writeTestFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.21\n")
	// worker_service.go would normally trigger no-layer-suffix ("_service" is banned),
	// but in ConsumerWorker preset the worker dir has TypePattern exemption.
	writeTestFile(t, filepath.Join(root, "internal", "worker", "worker_service.go"),
		"package worker\n\ntype ServiceWorker struct{}\n")

	pkgs := loadTestPackages(t, root)
	violations := rules.CheckNaming(pkgs, rules.WithModel(m))
	for _, v := range violations {
		if v.Rule == "naming.no-layer-suffix" {
			t.Errorf("worker_service.go should be exempt from no-layer-suffix, got: %s", v.Message)
		}
	}
}

func TestCheckNaming_RepoFileExtraInterface(t *testing.T) {
	root := t.TempDir()
	module := "example.com/repomulti"

	writeTestFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.21\n")
	writeTestFile(t, filepath.Join(root, "internal", "domain", "order", "alias.go"),
		"package order\n")
	writeTestFile(t, filepath.Join(root, "internal", "domain", "order", "core", "model", "order.go"),
		"package model\n\ntype Order struct{}\n")
	// review.go defines Review (correct) plus an extra Helper interface
	writeTestFile(t, filepath.Join(root, "internal", "domain", "order", "core", "repo", "review.go"),
		"package repo\n\ntype Review interface { Find() }\ntype Helper interface { Assist() }\n")

	pkgs := loadTestPackages(t, root)
	violations := rules.CheckNaming(pkgs)
	found := false
	for _, v := range violations {
		if v.Rule == "structure.repo-file-extra-interface" {
			found = true
		}
	}
	if !found {
		t.Error("expected structure.repo-file-extra-interface violation for extra interface in repo/review.go")
	}
}

func TestCheckNaming_RepoInterfaceTooLarge(t *testing.T) {
	root := t.TempDir()
	module := "example.com/repobig"

	writeTestFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.21\n")
	writeTestFile(t, filepath.Join(root, "internal", "domain", "order", "alias.go"),
		"package order\n")
	writeTestFile(t, filepath.Join(root, "internal", "domain", "order", "core", "model", "order.go"),
		"package model\n\ntype Order struct{}\n")
	// review.go has an interface with 11 methods
	writeTestFile(t, filepath.Join(root, "internal", "domain", "order", "core", "repo", "review.go"),
		"package repo\n\ntype Review interface {\n"+
			"\tA()\n\tB()\n\tC()\n\tD()\n\tE()\n"+
			"\tF()\n\tG()\n\tH()\n\tI()\n\tJ()\n\tK()\n}\n")

	pkgs := loadTestPackages(t, root)
	violations := rules.CheckNaming(pkgs, rules.WithMaxRepoInterfaceMethods(10))

	found := false
	for _, v := range violations {
		if v.Rule == "interface.too-many-methods" {
			found = true
		}
	}
	if !found {
		t.Error("expected interface.too-many-methods violation for 11-method interface")
	}

	// Without option, no violation
	violations2 := rules.CheckNaming(pkgs)
	for _, v := range violations2 {
		if v.Rule == "interface.too-many-methods" {
			t.Error("should not flag repo-interface-too-large without WithMaxRepoInterfaceMethods option")
		}
	}
}
