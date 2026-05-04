package orchestration_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/analyzer"
	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/presets"
	"github.com/NamhaeSusan/go-arch-guard/rules/orchestration"
)

func TestLogicBudgetSpec(t *testing.T) {
	spec := orchestration.NewLogicBudget(orchestration.WithSeverity(core.Error)).Spec()

	if spec.ID != "orchestration.logic-budget" {
		t.Fatalf("ID = %q, want orchestration.logic-budget", spec.ID)
	}
	if spec.DefaultSeverity != core.Error {
		t.Fatalf("DefaultSeverity = %v, want Error", spec.DefaultSeverity)
	}
}

func TestLogicBudgetFlagsBranchHeavyOrchestration(t *testing.T) {
	ctx := orchestrationContext(t, map[string]string{
		"internal/orchestration/checkout.go": `package orchestration

import "errors"

type Checkout struct{}

func (c *Checkout) Place(total int, country string, stock int, quantity int) error {
	if total > 10000 {
		total -= total / 10
	}
	if country == "KR" {
		total = total
	}
	if stock < quantity {
		return errors.New("stock")
	}
	return nil
}
`,
	})

	got := orchestration.NewLogicBudget(
		orchestration.WithMaxBranches(2),
		orchestration.WithMaxStatements(100),
		orchestration.WithMaxCyclomatic(100),
	).Check(ctx)

	assertLogicBudgetViolation(t, got, "Place", "branches 3 > 2")
}

func TestLogicBudgetFlagsStatementHeavyOrchestration(t *testing.T) {
	ctx := orchestrationContext(t, map[string]string{
		"internal/orchestration/checkout.go": `package orchestration

func step() {}

func Place() {
	step()
	step()
	step()
	step()
}
`,
	})

	got := orchestration.NewLogicBudget(
		orchestration.WithMaxBranches(100),
		orchestration.WithMaxStatements(3),
		orchestration.WithMaxCyclomatic(100),
	).Check(ctx)

	assertLogicBudgetViolation(t, got, "Place", "statements 4 > 3")
}

func TestLogicBudgetDiscountsSimpleErrorHandlingBranches(t *testing.T) {
	ctx := orchestrationContext(t, map[string]string{
		"internal/orchestration/checkout.go": `package orchestration

import "fmt"

type Step interface{ Do() error }

type Checkout struct {
	a Step
	b Step
	c Step
}

func (c *Checkout) Place() error {
	if err := c.a.Do(); err != nil {
		return fmt.Errorf("a: %w", err)
	}
	if err := c.b.Do(); err != nil {
		return err
	}
	if err := c.c.Do(); err != nil {
		return err
	}
	return nil
}
`,
	})

	got := orchestration.NewLogicBudget(
		orchestration.WithMaxBranches(0),
		orchestration.WithMaxStatements(4),
		orchestration.WithMaxCyclomatic(1),
	).Check(ctx)

	if len(got) != 0 {
		t.Fatalf("simple error-handling coordination should pass, got %+v", got)
	}

	got = orchestration.NewLogicBudget(
		orchestration.WithMaxBranches(0),
		orchestration.WithMaxStatements(4),
		orchestration.WithMaxCyclomatic(1),
		orchestration.WithCountErrorBranches(),
	).Check(ctx)

	assertLogicBudgetViolation(t, got, "Place", "branches 3 > 0")
}

func TestLogicBudgetDoesNotDiscountPolicyReturnsThatMentionErr(t *testing.T) {
	ctx := orchestrationContext(t, map[string]string{
		"internal/orchestration/checkout.go": `package orchestration

import (
	"errors"
	"io"
)

func step() error { return nil }

func Place() bool {
	if err := step(); err != nil {
		return errors.Is(err, io.EOF)
	}
	return true
}
`,
	})

	got := orchestration.NewLogicBudget(
		orchestration.WithMaxBranches(0),
		orchestration.WithMaxStatements(100),
		orchestration.WithMaxCyclomatic(100),
	).Check(ctx)

	assertLogicBudgetViolation(t, got, "Place", "branches 1 > 0")
}

func TestLogicBudgetOnlyDiscountsFmtErrorfWhenCheckedErrIsWrapped(t *testing.T) {
	ctx := orchestrationContext(t, map[string]string{
		"internal/orchestration/checkout.go": `package orchestration

import (
	"errors"
	"fmt"
)

var ErrPolicy = errors.New("policy")

func step() error { return nil }

func Place() error {
	if err := step(); err != nil {
		return fmt.Errorf("validation failed: %v: %w", err, ErrPolicy)
	}
	return nil
}
`,
	})

	got := orchestration.NewLogicBudget(
		orchestration.WithMaxBranches(0),
		orchestration.WithMaxStatements(100),
		orchestration.WithMaxCyclomatic(100),
	).Check(ctx)

	assertLogicBudgetViolation(t, got, "Place", "branches 1 > 0")
}

func TestLogicBudgetDoesNotDiscountErrorsJoinWithPolicyError(t *testing.T) {
	ctx := orchestrationContext(t, map[string]string{
		"internal/orchestration/checkout.go": `package orchestration

import "errors"

var ErrPolicy = errors.New("policy")

func step() error { return nil }

func Place() error {
	if err := step(); err != nil {
		return errors.Join(ErrPolicy, err)
	}
	return nil
}
`,
	})

	got := orchestration.NewLogicBudget(
		orchestration.WithMaxBranches(0),
		orchestration.WithMaxStatements(100),
		orchestration.WithMaxCyclomatic(100),
	).Check(ctx)

	assertLogicBudgetViolation(t, got, "Place", "branches 1 > 0")
}

func TestLogicBudgetCountsFunctionLiteralBodies(t *testing.T) {
	ctx := orchestrationContext(t, map[string]string{
		"internal/orchestration/checkout.go": `package orchestration

func run(fn func() error) error { return fn() }

func Place(total int, country string) error {
	return run(func() error {
		if total > 10000 {
			total -= total / 10
		}
		if country == "KR" {
			total = total
		}
		return nil
	})
}
`,
	})

	got := orchestration.NewLogicBudget(
		orchestration.WithMaxBranches(1),
		orchestration.WithMaxStatements(100),
		orchestration.WithMaxCyclomatic(100),
	).Check(ctx)

	assertLogicBudgetViolation(t, got, "Place", "branches 2 > 1")
}

func TestLogicBudgetCountsFunctionLiteralBodiesInControlExpressions(t *testing.T) {
	ctx := orchestrationContext(t, map[string]string{
		"internal/orchestration/checkout.go": `package orchestration

func run(fn func() bool) bool { return fn() }

func Place(total int, country string) error {
	if run(func() bool {
		if total > 10000 {
			total -= total / 10
		}
		if country == "KR" {
			total = total
		}
		return true
	}) {
		return nil
	}
	return nil
}
`,
	})

	got := orchestration.NewLogicBudget(
		orchestration.WithMaxBranches(2),
		orchestration.WithMaxStatements(100),
		orchestration.WithMaxCyclomatic(100),
	).Check(ctx)

	assertLogicBudgetViolation(t, got, "Place", "branches 3 > 2")
}

func TestLogicBudgetThresholdsAndIgnoredFunctionsAreConfigurable(t *testing.T) {
	ctx := orchestrationContext(t, map[string]string{
		"internal/orchestration/checkout.go": `package orchestration

func Place(total int, country string, stock int, quantity int) error {
	if total > 10000 {
		total -= total / 10
	}
	if country == "KR" {
		total = total
	}
	if stock < quantity {
		return nil
	}
	return nil
}
`,
	})

	got := orchestration.NewLogicBudget(
		orchestration.WithMaxBranches(3),
		orchestration.WithMaxStatements(100),
		orchestration.WithMaxCyclomatic(4),
	).Check(ctx)
	if len(got) != 0 {
		t.Fatalf("custom thresholds should allow Place, got %+v", got)
	}

	got = orchestration.NewLogicBudget(
		orchestration.WithMaxBranches(0),
		orchestration.WithIgnoredFunctions("Place"),
	).Check(ctx)
	if len(got) != 0 {
		t.Fatalf("ignored function should pass, got %+v", got)
	}
}

func TestLogicBudgetIgnoredPathsAreConfigurable(t *testing.T) {
	ctx := orchestrationContext(t, map[string]string{
		"internal/orchestration/handler/http/checkout.go": `package http

func Place(total int, country string, stock int, quantity int) error {
	if total > 10000 {
		total -= total / 10
	}
	if country == "KR" {
		total = total
	}
	if stock < quantity {
		return nil
	}
	return nil
}
`,
	})

	got := orchestration.NewLogicBudget(
		orchestration.WithMaxBranches(0),
		orchestration.WithIgnoredPaths("internal/orchestration/handler/..."),
	).Check(ctx)
	if len(got) != 0 {
		t.Fatalf("ignored path should pass, got %+v", got)
	}
}

func TestLogicBudgetOnlyChecksOrchestrationPackages(t *testing.T) {
	ctx := orchestrationContext(t, map[string]string{
		"internal/domain/order/app/service.go": `package app

func Place(total int, country string, stock int, quantity int) error {
	if total > 10000 {
		total -= total / 10
	}
	if country == "KR" {
		total = total
	}
	if stock < quantity {
		return nil
	}
	return nil
}
`,
	})

	got := orchestration.NewLogicBudget(orchestration.WithMaxBranches(0)).Check(ctx)
	if len(got) != 0 {
		t.Fatalf("non-orchestration package should not be checked, got %+v", got)
	}
}

func orchestrationContext(t *testing.T, files map[string]string) *core.Context {
	t.Helper()
	root := t.TempDir()
	writeOrchestrationFile(t, filepath.Join(root, "go.mod"), "module example.com/shop\n\ngo 1.25.0\n")
	for name, content := range files {
		writeOrchestrationFile(t, filepath.Join(root, name), content)
	}
	pkgs, err := analyzer.Load(root, "internal/...")
	if err != nil {
		t.Fatal(err)
	}
	return core.NewContext(pkgs, "example.com/shop", root, presets.DDD(), nil)
}

func writeOrchestrationFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func assertLogicBudgetViolation(t *testing.T, violations []core.Violation, funcName, want string) {
	t.Helper()
	for _, v := range violations {
		if v.Rule == "orchestration.logic-budget" &&
			strings.Contains(v.Message, funcName) &&
			strings.Contains(v.Message, want) {
			if v.DefaultSeverity != core.Warning || v.EffectiveSeverity != core.Warning {
				t.Fatalf("severity = default %v effective %v, want Warning/Warning", v.DefaultSeverity, v.EffectiveSeverity)
			}
			return
		}
	}
	t.Fatalf("expected orchestration.logic-budget violation for %s containing %q, got %+v", funcName, want, violations)
}
