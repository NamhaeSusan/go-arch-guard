package report_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/report"
)

type fakeTB struct {
	testing.TB
	errors []string
	logs   []string
	failed bool
}

func (f *fakeTB) Errorf(format string, args ...any) {
	f.errors = append(f.errors, fmt.Sprintf(format, args...))
	f.failed = true
}

func (f *fakeTB) Helper() {}

func (f *fakeTB) Log(args ...any) {
	f.logs = append(f.logs, fmt.Sprint(args...))
}

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

	t.Run("logs every violation with file/line/rule/message/fix and severity", func(t *testing.T) {
		// Regression guard for the human/bot debugging output. AssertNoViolations
		// is the primary signal a CI log carries, so the log line must keep
		// every field that helps a reader (or remediation bot) locate the
		// offending code: severity, file:line, rule, message, fix.
		tb := &fakeTB{}
		report.AssertNoViolations(tb, []core.Violation{
			{
				File:              "internal/order/app/service.go",
				Line:              42,
				Rule:              "test.rule",
				Message:           "domain leak",
				Fix:               "move to orchestration",
				EffectiveSeverity: core.Error,
			},
			{
				Rule:              "meta.no-matching-packages",
				Message:           "module mismatch",
				EffectiveSeverity: core.Warning,
			},
		})
		if len(tb.logs) != 2 {
			t.Fatalf("expected 2 log lines (one per violation), got %d: %v", len(tb.logs), tb.logs)
		}
		first := tb.logs[0]
		for _, frag := range []string{"ERROR", "internal/order/app/service.go", "42", "test.rule", "domain leak", "move to orchestration"} {
			if !strings.Contains(first, frag) {
				t.Errorf("log[0] missing %q\n%s", frag, first)
			}
		}
		// meta violation: severity, rule, message must be present even with empty file.
		second := tb.logs[1]
		for _, frag := range []string{"WARNING", "meta.no-matching-packages", "module mismatch"} {
			if !strings.Contains(second, frag) {
				t.Errorf("log[1] missing %q\n%s", frag, second)
			}
		}
	})
}
