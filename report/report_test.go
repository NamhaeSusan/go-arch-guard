package report_test

import (
	"fmt"
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/report"
)

type fakeTB struct {
	testing.TB
	errors []string
	failed bool
}

func (f *fakeTB) Errorf(format string, args ...any) {
	f.errors = append(f.errors, fmt.Sprintf(format, args...))
	f.failed = true
}

func (f *fakeTB) Helper() {}

func (f *fakeTB) Log(args ...any) {}

func TestAssertNoViolations(t *testing.T) {
	t.Run("no violations passes", func(t *testing.T) {
		tb := &fakeTB{}
		report.AssertNoViolations(tb, nil)
		if tb.failed {
			t.Error("expected test to pass with no violations")
		}
	})

	t.Run("error violations fails test", func(t *testing.T) {
		tb := &fakeTB{}
		report.AssertNoViolations(tb, []core.Violation{
			{Rule: "test.rule", Message: "bad", EffectiveSeverity: core.Error},
		})
		if !tb.failed {
			t.Error("expected test to fail with error violations")
		}
	})

	t.Run("warning-only violations passes test", func(t *testing.T) {
		tb := &fakeTB{}
		report.AssertNoViolations(tb, []core.Violation{
			{Rule: "test.rule", Message: "warn", EffectiveSeverity: core.Warning},
		})
		if tb.failed {
			t.Error("expected test to pass with warning-only violations")
		}
	})
}
