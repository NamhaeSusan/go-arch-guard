# New Architecture Presets Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `Layered()`, `Hexagonal()`, and `ModularMonolith()` preset factory functions to the rules package.

**Architecture:** Each preset is a standalone `Model` factory function in `rules/model.go`, following the exact pattern of existing `DDD()` and `CleanArch()`. All rule checks already consume `Model` — no changes needed in rule logic. Integration tests validate both valid structures and direction violations.

**Tech Stack:** Go, `golang.org/x/tools/go/packages`

---

### Task 1: Add Layered() unit test and implementation

**Files:**
- Modify: `rules/model_test.go`
- Modify: `rules/model.go`

- [ ] **Step 1: Write the failing test in `rules/model_test.go`**

Add after `TestCleanArch_ReturnsValidModel`:

```go
func TestLayered_ReturnsValidModel(t *testing.T) {
	m := Layered()
	if len(m.Sublayers) != 4 {
		t.Fatalf("Layered model sublayer count = %d, want 4", len(m.Sublayers))
	}
	if m.RequireAlias {
		t.Error("Layered should not require alias")
	}
	if m.RequireModel {
		t.Error("Layered should not require domain model")
	}
	if m.ModelPath != "model" {
		t.Errorf("ModelPath = %q, want %q", m.ModelPath, "model")
	}
	if !m.PkgRestricted["model"] {
		t.Error("model sublayer must be pkg-restricted")
	}
	if m.DomainDir != "domain" {
		t.Errorf("DomainDir = %q, want %q", m.DomainDir, "domain")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/kimtaeyun/go-arch-guard && go test ./rules/ -run TestLayered_ReturnsValidModel -v`
Expected: FAIL — `Layered` undefined

- [ ] **Step 3: Implement `Layered()` in `rules/model.go`**

Add after the `CleanArch()` function:

```go
// Layered returns a Spring-style layered architecture model.
func Layered() Model {
	return Model{
		Sublayers: []string{
			"handler", "service", "repository", "model",
		},
		Direction: map[string][]string{
			"handler":    {"service"},
			"service":    {"repository", "model"},
			"repository": {"model"},
			"model":      {},
		},
		PkgRestricted: map[string]bool{
			"model": true,
		},
		InternalTopLevel: map[string]bool{
			"domain": true, "orchestration": true, "pkg": true,
		},
		DomainDir:        "domain",
		OrchestrationDir: "orchestration",
		SharedDir:        "pkg",
		RequireAlias:     false,
		AliasFileName:    "alias.go",
		RequireModel:     false,
		ModelPath:        "model",
		DTOAllowedLayers: []string{"handler", "service"},
		BannedPkgNames:   []string{"util", "common", "misc", "helper", "shared", "services"},
		LegacyPkgNames:   []string{"router", "bootstrap"},
		LayerDirNames: map[string]bool{
			"handler": true, "service": true, "repository": true, "model": true,
			"controller": true, "entity": true, "store": true,
			"persistence": true, "domain": true,
		},
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /Users/kimtaeyun/go-arch-guard && go test ./rules/ -run TestLayered_ReturnsValidModel -v`
Expected: PASS

- [ ] **Step 5: Update `TestModelConsistency` in `rules/model_test.go`**

Find the table in `TestModelConsistency` and add the new entry:

```go
{"Layered", Layered()},
```

- [ ] **Step 6: Run consistency test**

Run: `cd /Users/kimtaeyun/go-arch-guard && go test ./rules/ -run TestModelConsistency -v`
Expected: PASS — validates all sublayers appear in Direction map and vice versa

---

### Task 2: Add Hexagonal() unit test and implementation

**Files:**
- Modify: `rules/model_test.go`
- Modify: `rules/model.go`

- [ ] **Step 1: Write the failing test in `rules/model_test.go`**

Add after `TestLayered_ReturnsValidModel`:

```go
func TestHexagonal_ReturnsValidModel(t *testing.T) {
	m := Hexagonal()
	if len(m.Sublayers) != 5 {
		t.Fatalf("Hexagonal model sublayer count = %d, want 5", len(m.Sublayers))
	}
	if m.RequireAlias {
		t.Error("Hexagonal should not require alias")
	}
	if m.RequireModel {
		t.Error("Hexagonal should not require domain model")
	}
	if m.ModelPath != "domain" {
		t.Errorf("ModelPath = %q, want %q", m.ModelPath, "domain")
	}
	if !m.PkgRestricted["domain"] {
		t.Error("domain sublayer must be pkg-restricted")
	}
	// Verify adapter can reach port and domain
	allowed := m.Direction["adapter"]
	if len(allowed) != 2 {
		t.Errorf("adapter allowed imports = %v, want [port domain]", allowed)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/kimtaeyun/go-arch-guard && go test ./rules/ -run TestHexagonal_ReturnsValidModel -v`
Expected: FAIL — `Hexagonal` undefined

- [ ] **Step 3: Implement `Hexagonal()` in `rules/model.go`**

Add after `Layered()`:

```go
// Hexagonal returns a Ports & Adapters architecture model.
func Hexagonal() Model {
	return Model{
		Sublayers: []string{
			"handler", "usecase", "port", "domain", "adapter",
		},
		Direction: map[string][]string{
			"handler": {"usecase"},
			"usecase": {"port", "domain"},
			"port":    {"domain"},
			"domain":  {},
			"adapter": {"port", "domain"},
		},
		PkgRestricted: map[string]bool{
			"domain": true,
		},
		InternalTopLevel: map[string]bool{
			"domain": true, "orchestration": true, "pkg": true,
		},
		DomainDir:        "domain",
		OrchestrationDir: "orchestration",
		SharedDir:        "pkg",
		RequireAlias:     false,
		AliasFileName:    "alias.go",
		RequireModel:     false,
		ModelPath:        "domain",
		DTOAllowedLayers: []string{"handler", "usecase"},
		BannedPkgNames:   []string{"util", "common", "misc", "helper", "shared", "services"},
		LegacyPkgNames:   []string{"router", "bootstrap"},
		LayerDirNames: map[string]bool{
			"handler": true, "usecase": true, "port": true,
			"domain": true, "adapter": true,
			"controller": true, "service": true, "entity": true,
			"store": true, "persistence": true,
		},
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /Users/kimtaeyun/go-arch-guard && go test ./rules/ -run TestHexagonal_ReturnsValidModel -v`
Expected: PASS

- [ ] **Step 5: Add to `TestModelConsistency`**

Add to the table:

```go
{"Hexagonal", Hexagonal()},
```

- [ ] **Step 6: Run consistency test**

Run: `cd /Users/kimtaeyun/go-arch-guard && go test ./rules/ -run TestModelConsistency -v`
Expected: PASS

---

### Task 3: Add ModularMonolith() unit test and implementation

**Files:**
- Modify: `rules/model_test.go`
- Modify: `rules/model.go`

- [ ] **Step 1: Write the failing test in `rules/model_test.go`**

Add after `TestHexagonal_ReturnsValidModel`:

```go
func TestModularMonolith_ReturnsValidModel(t *testing.T) {
	m := ModularMonolith()
	if len(m.Sublayers) != 4 {
		t.Fatalf("ModularMonolith model sublayer count = %d, want 4", len(m.Sublayers))
	}
	if m.RequireAlias {
		t.Error("ModularMonolith should not require alias")
	}
	if m.RequireModel {
		t.Error("ModularMonolith should not require domain model")
	}
	if m.ModelPath != "domain" {
		t.Errorf("ModelPath = %q, want %q", m.ModelPath, "domain")
	}
	if !m.PkgRestricted["domain"] {
		t.Error("domain sublayer must be pkg-restricted")
	}
	if m.DomainDir != "domain" {
		t.Errorf("DomainDir = %q, want %q", m.DomainDir, "domain")
	}
	// infrastructure can only reach domain
	allowed := m.Direction["infrastructure"]
	if len(allowed) != 1 || allowed[0] != "domain" {
		t.Errorf("infrastructure allowed = %v, want [domain]", allowed)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/kimtaeyun/go-arch-guard && go test ./rules/ -run TestModularMonolith_ReturnsValidModel -v`
Expected: FAIL — `ModularMonolith` undefined

- [ ] **Step 3: Implement `ModularMonolith()` in `rules/model.go`**

Add after `Hexagonal()`:

```go
// ModularMonolith returns a module-based layered architecture model.
func ModularMonolith() Model {
	return Model{
		Sublayers: []string{
			"api", "application", "domain", "infrastructure",
		},
		Direction: map[string][]string{
			"api":            {"application"},
			"application":    {"domain"},
			"domain":         {},
			"infrastructure": {"domain"},
		},
		PkgRestricted: map[string]bool{
			"domain": true,
		},
		InternalTopLevel: map[string]bool{
			"domain": true, "orchestration": true, "pkg": true,
		},
		DomainDir:        "domain",
		OrchestrationDir: "orchestration",
		SharedDir:        "pkg",
		RequireAlias:     false,
		AliasFileName:    "alias.go",
		RequireModel:     false,
		ModelPath:        "domain",
		DTOAllowedLayers: []string{"api", "application"},
		BannedPkgNames:   []string{"util", "common", "misc", "helper", "shared", "services"},
		LegacyPkgNames:   []string{"router", "bootstrap"},
		LayerDirNames: map[string]bool{
			"api": true, "application": true, "domain": true,
			"infrastructure": true,
			"controller": true, "service": true, "entity": true,
			"store": true, "persistence": true,
		},
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /Users/kimtaeyun/go-arch-guard && go test ./rules/ -run TestModularMonolith_ReturnsValidModel -v`
Expected: PASS

- [ ] **Step 5: Add to `TestModelConsistency`**

Add to the table:

```go
{"ModularMonolith", ModularMonolith()},
```

- [ ] **Step 6: Run all model tests**

Run: `cd /Users/kimtaeyun/go-arch-guard && go test ./rules/ -run "TestModelConsistency|TestLayered|TestHexagonal|TestModularMonolith" -v`
Expected: All PASS

- [ ] **Step 7: Commit**

```bash
git add rules/model.go rules/model_test.go
git commit -m "feat: add Layered, Hexagonal, ModularMonolith architecture presets"
```

---

### Task 4: Integration test — Layered valid project

**Files:**
- Modify: `integration_test.go`

- [ ] **Step 1: Write the integration test**

Add to `integration_test.go`:

```go
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
```

- [ ] **Step 2: Run test**

Run: `cd /Users/kimtaeyun/go-arch-guard && go test . -run TestIntegration_RealWorldLayered -v -count=1`
Expected: PASS

---

### Task 5: Integration test — Layered direction violation

**Files:**
- Modify: `integration_test.go`

- [ ] **Step 1: Write the violation test**

Add to `integration_test.go`:

```go
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

	t.Run("detects direction violation", func(t *testing.T) {
		violations := rules.CheckLayerDirection(pkgs, module, root, opts...)
		assertHasRule(t, violations, "layer.direction")
	})
	t.Run("detects model importing pkg", func(t *testing.T) {
		violations := rules.CheckLayerDirection(pkgs, module, root, opts...)
		assertHasRule(t, violations, "layer.inner-imports-pkg")
	})
}
```

- [ ] **Step 2: Run test**

Run: `cd /Users/kimtaeyun/go-arch-guard && go test . -run TestIntegration_Layered_Violations -v -count=1`
Expected: PASS

---

### Task 6: Integration test — Hexagonal valid + violations

**Files:**
- Modify: `integration_test.go`

- [ ] **Step 1: Write Hexagonal valid test**

```go
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
```

- [ ] **Step 2: Write Hexagonal violation test**

```go
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

	t.Run("detects direction violation", func(t *testing.T) {
		violations := rules.CheckLayerDirection(pkgs, module, root, opts...)
		assertHasRule(t, violations, "layer.direction")
	})
	t.Run("detects domain importing pkg", func(t *testing.T) {
		violations := rules.CheckLayerDirection(pkgs, module, root, opts...)
		assertHasRule(t, violations, "layer.inner-imports-pkg")
	})
}
```

- [ ] **Step 3: Run both tests**

Run: `cd /Users/kimtaeyun/go-arch-guard && go test . -run "TestIntegration_RealWorldHexagonal|TestIntegration_Hexagonal_Violations" -v -count=1`
Expected: PASS

---

### Task 7: Integration test — ModularMonolith valid + violations

**Files:**
- Modify: `integration_test.go`

- [ ] **Step 1: Write ModularMonolith valid test**

```go
func TestIntegration_RealWorldModularMonolith(t *testing.T) {
	root := t.TempDir()
	module := "example.com/modshop"
	m := rules.ModularMonolith()

	writeIntegrationFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.25.0\n")

	// --- domain: order ---
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "order", "api", "handler.go"),
		"package api\n\nimport _ \""+module+"/internal/domain/order/application\"\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "order", "application", "create_order.go"),
		"package application\n\nimport _ \""+module+"/internal/domain/order/domain\"\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "order", "domain", "order.go"),
		"package domain\n\ntype Order struct{ ID string }\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "order", "infrastructure", "pg_order.go"),
		"package infrastructure\n\nimport _ \""+module+"/internal/domain/order/domain\"\n\ntype PgOrder struct{}\n")

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
```

- [ ] **Step 2: Write ModularMonolith violation test**

```go
func TestIntegration_ModularMonolith_Violations(t *testing.T) {
	root := t.TempDir()
	module := "example.com/modinvalid"
	m := rules.ModularMonolith()

	writeIntegrationFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.25.0\n")

	// domain imports api — direction violation
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "order", "domain", "order.go"),
		"package domain\n\nimport _ \""+module+"/internal/domain/order/api\"\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "order", "api", "handler.go"),
		"package api\n")
	// infrastructure imports application — direction violation
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "order", "infrastructure", "bad.go"),
		"package infrastructure\n\nimport _ \""+module+"/internal/domain/order/application\"\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "order", "application", "service.go"),
		"package application\n")
	// cross-domain: order imports user
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "order", "application", "cross.go"),
		"package application\n\nimport _ \""+module+"/internal/domain/user/domain\"\n")
	writeIntegrationFile(t, filepath.Join(root, "internal", "domain", "user", "domain", "user.go"),
		"package domain\n")

	pkgs, err := analyzer.Load(root, "internal/...")
	if err != nil {
		t.Log("partial load:", err)
	}

	opts := []rules.Option{rules.WithModel(m)}

	t.Run("detects direction violation", func(t *testing.T) {
		violations := rules.CheckLayerDirection(pkgs, module, root, opts...)
		assertHasRule(t, violations, "layer.direction")
	})
	t.Run("detects cross-domain import", func(t *testing.T) {
		violations := rules.CheckDomainIsolation(pkgs, module, root, opts...)
		assertHasRule(t, violations, "isolation.cross-domain")
	})
}
```

- [ ] **Step 3: Run both tests**

Run: `cd /Users/kimtaeyun/go-arch-guard && go test . -run "TestIntegration_RealWorldModularMonolith|TestIntegration_ModularMonolith_Violations" -v -count=1`
Expected: PASS

- [ ] **Step 4: Run full test suite**

Run: `cd /Users/kimtaeyun/go-arch-guard && go test ./... -count=1`
Expected: All PASS

- [ ] **Step 5: Commit integration tests**

```bash
git add integration_test.go
git commit -m "test: add integration tests for Layered, Hexagonal, ModularMonolith presets"
```

---

### Task 8: Update documentation

**Files:**
- Modify: `README.md`
- Modify: `README.ko.md`
- Modify: `plugins/go-arch-guard/skills/go-arch-guard/SKILL.md`

- [ ] **Step 1: Update README.md — Built-in Presets table**

In `README.md`, find the "Built-in Presets" table and replace it with:

```markdown
| Preset | Sublayers | Direction | Alias Required | Model Required |
|--------|-----------|-----------|:-:|:-:|
| `DDD()` | handler, app, core/model, core/repo, core/svc, event, infra | handler→app→core/\*, infra→core/repo | Yes | Yes |
| `CleanArch()` | handler, usecase, entity, gateway, infra | handler→usecase→entity+gateway, infra→gateway | No | No |
| `Layered()` | handler, service, repository, model | handler→service→repository+model, repository→model | No | No |
| `Hexagonal()` | handler, usecase, port, domain, adapter | handler→usecase→port+domain, adapter→port+domain | No | No |
| `ModularMonolith()` | api, application, domain, infrastructure | api→application→domain, infrastructure→domain | No | No |
```

- [ ] **Step 2: Add Layered layout and direction table after Clean Architecture section in README.md**

Insert after the Clean Architecture direction table and before "### Custom Model":

```markdown
### Layered (Spring-style) Layout

​```text
internal/
├── domain/
│   └── order/
│       ├── handler/              # HTTP/gRPC handlers
│       ├── service/              # business logic
│       ├── repository/           # data access
│       └── model/                # domain models
├── orchestration/
└── pkg/
​```

Layered direction:

| from | allowed to import |
|------|-------------------|
| `handler` | `service` |
| `service` | `repository`, `model` |
| `repository` | `model` |
| `model` | nothing |

### Hexagonal (Ports & Adapters) Layout

​```text
internal/
├── domain/
│   └── order/
│       ├── handler/              # driving adapters (HTTP, gRPC)
│       ├── usecase/              # application logic
│       ├── port/                 # interfaces (inbound + outbound)
│       ├── domain/               # entities, value objects
│       └── adapter/              # driven adapters (DB, messaging)
├── orchestration/
└── pkg/
​```

Hexagonal direction:

| from | allowed to import |
|------|-------------------|
| `handler` | `usecase` |
| `usecase` | `port`, `domain` |
| `port` | `domain` |
| `domain` | nothing |
| `adapter` | `port`, `domain` |

### Modular Monolith Layout

​```text
internal/
├── domain/
│   └── order/
│       ├── api/                  # module public interface
│       ├── application/          # use cases
│       ├── domain/               # entities, value objects
│       └── infrastructure/       # DB, external services
├── orchestration/
└── pkg/
​```

Modular Monolith direction:

| from | allowed to import |
|------|-------------------|
| `api` | `application` |
| `application` | `domain` |
| `domain` | nothing |
| `infrastructure` | `domain` |
```

- [ ] **Step 3: Add Quick Start examples for new presets in README.md**

After the "### Clean Architecture" quick start section, add:

```markdown
### Layered (Spring-style)

​```go
func TestArchitecture(t *testing.T) {
    pkgs, err := analyzer.Load(".", "internal/...", "cmd/...")
    if err != nil {
        t.Log(err)
    }
    if len(pkgs) == 0 {
        t.Fatalf("no packages loaded: %v", err)
    }

    m := rules.Layered()
    opts := []rules.Option{rules.WithModel(m)}

    t.Run("domain isolation", func(t *testing.T) {
        report.AssertNoViolations(t, rules.CheckDomainIsolation(pkgs, "", "", opts...))
    })
    t.Run("layer direction", func(t *testing.T) {
        report.AssertNoViolations(t, rules.CheckLayerDirection(pkgs, "", "", opts...))
    })
    t.Run("naming", func(t *testing.T) {
        report.AssertNoViolations(t, rules.CheckNaming(pkgs, opts...))
    })
    t.Run("structure", func(t *testing.T) {
        report.AssertNoViolations(t, rules.CheckStructure(".", opts...))
    })
}
​```

### Hexagonal (Ports & Adapters)

​```go
func TestArchitecture(t *testing.T) {
    pkgs, err := analyzer.Load(".", "internal/...", "cmd/...")
    if err != nil {
        t.Log(err)
    }
    if len(pkgs) == 0 {
        t.Fatalf("no packages loaded: %v", err)
    }

    m := rules.Hexagonal()
    opts := []rules.Option{rules.WithModel(m)}

    t.Run("domain isolation", func(t *testing.T) {
        report.AssertNoViolations(t, rules.CheckDomainIsolation(pkgs, "", "", opts...))
    })
    t.Run("layer direction", func(t *testing.T) {
        report.AssertNoViolations(t, rules.CheckLayerDirection(pkgs, "", "", opts...))
    })
    t.Run("naming", func(t *testing.T) {
        report.AssertNoViolations(t, rules.CheckNaming(pkgs, opts...))
    })
    t.Run("structure", func(t *testing.T) {
        report.AssertNoViolations(t, rules.CheckStructure(".", opts...))
    })
}
​```

### Modular Monolith

​```go
func TestArchitecture(t *testing.T) {
    pkgs, err := analyzer.Load(".", "internal/...", "cmd/...")
    if err != nil {
        t.Log(err)
    }
    if len(pkgs) == 0 {
        t.Fatalf("no packages loaded: %v", err)
    }

    m := rules.ModularMonolith()
    opts := []rules.Option{rules.WithModel(m)}

    t.Run("domain isolation", func(t *testing.T) {
        report.AssertNoViolations(t, rules.CheckDomainIsolation(pkgs, "", "", opts...))
    })
    t.Run("layer direction", func(t *testing.T) {
        report.AssertNoViolations(t, rules.CheckLayerDirection(pkgs, "", "", opts...))
    })
    t.Run("naming", func(t *testing.T) {
        report.AssertNoViolations(t, rules.CheckNaming(pkgs, opts...))
    })
    t.Run("structure", func(t *testing.T) {
        report.AssertNoViolations(t, rules.CheckStructure(".", opts...))
    })
}
​```
```

- [ ] **Step 4: Update API Reference table in README.md**

Add after `rules.CleanArch()` row:

```markdown
| `rules.Layered()` | Spring-style layered model |
| `rules.Hexagonal()` | Ports & Adapters model |
| `rules.ModularMonolith()` | Module-based layered model |
```

- [ ] **Step 5: Update README.md intro**

Change the first paragraph:
```
Ships with **DDD** and **Clean Architecture** presets
```
to:
```
Ships with **DDD**, **Clean Architecture**, **Layered**, **Hexagonal**, and **Modular Monolith** presets
```

- [ ] **Step 6: Mirror all changes to README.ko.md**

Apply the same structural changes to `README.ko.md` (Korean version), translating section headers and descriptions while keeping code blocks identical.

- [ ] **Step 7: Update SKILL.md**

In `plugins/go-arch-guard/skills/go-arch-guard/SKILL.md`:

1. Update description frontmatter: `DDD / Clean Architecture / Layered / Hexagonal / Modular Monolith 프리셋 또는 커스텀 모델 지원.`

2. Add Option C/D/E sections in "## 2. Choose Architecture Model" after Option B, following the same format with layout diagrams, direction tables, and constraints.

3. Add Quick Start templates in "## 3. architecture_test.go Template" section for each new preset.

4. Update "## 6. Banned Patterns" DTOs row to include new preset DTO layers.

- [ ] **Step 8: Run full test suite and lint**

```bash
cd /Users/kimtaeyun/go-arch-guard
goimports -w .
go fix ./...
go test ./... -count=1
make lint
```

Expected: All pass, 0 lint issues.

- [ ] **Step 9: Commit documentation**

```bash
git add README.md README.ko.md plugins/go-arch-guard/skills/go-arch-guard/SKILL.md
git commit -m "docs: add Layered, Hexagonal, ModularMonolith preset documentation"
```
