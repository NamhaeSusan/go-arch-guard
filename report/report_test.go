package report_test

import (
	"fmt"
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/report"
	"github.com/NamhaeSusan/go-arch-guard/rules"
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
		report.AssertNoViolations(tb, []rules.Violation{
			{Rule: "test.rule", Message: "bad", Severity: rules.Error},
		})
		if !tb.failed {
			t.Error("expected test to fail with error violations")
		}
	})

	t.Run("warning-only violations passes test", func(t *testing.T) {
		tb := &fakeTB{}
		report.AssertNoViolations(tb, []rules.Violation{
			{Rule: "test.rule", Message: "warn", Severity: rules.Warning},
		})
		if tb.failed {
			t.Error("expected test to pass with warning-only violations")
		}
	})
}
