package rules_test

import (
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/rules"
)

func TestCheckNoSetters_FlagsClassicSetters(t *testing.T) {
	pkgs := loadSetters(t)
	got := rules.CheckNoSetters(pkgs)

	var modelViolations []rules.Violation
	for _, v := range got {
		if v.Rule == "setter.forbidden" {
			// only count violations from model/order.go
			modelViolations = append(modelViolations, v)
		}
	}

	if len(modelViolations) != 3 {
		t.Errorf("expected exactly 3 violations from model/order.go, got %d: %+v", len(modelViolations), modelViolations)
	}
}

func TestCheckNoSetters_SkipsFluentBuilder(t *testing.T) {
	pkgs := loadSetters(t)
	got := rules.CheckNoSetters(pkgs)

	for _, v := range got {
		if v.Rule == "setter.forbidden" && containsStr(v.File, "builder") {
			t.Errorf("fluent builder should not be flagged: %+v", v)
		}
	}
}

func TestCheckNoSetters_SkipsMocks(t *testing.T) {
	pkgs := loadSetters(t)
	got := rules.CheckNoSetters(pkgs)

	for _, v := range got {
		if v.Rule == "setter.forbidden" && containsStr(v.File, "mocks") {
			t.Errorf("mocks/ should be auto-excluded: %+v", v)
		}
	}
}

func TestCheckNoSetters_RespectsWithExclude(t *testing.T) {
	pkgs := loadSetters(t)
	got := rules.CheckNoSetters(pkgs, rules.WithExclude("internal/model/..."))

	for _, v := range got {
		if v.Rule == "setter.forbidden" {
			t.Errorf("WithExclude(internal/model/...) should suppress all setter violations, got: %+v", v)
		}
	}
}

func TestCheckNoSetters_RespectsSeverity(t *testing.T) {
	pkgs := loadSetters(t)
	got := rules.CheckNoSetters(pkgs, rules.WithSeverity(rules.Error))

	for _, v := range got {
		if v.Rule == "setter.forbidden" && v.Severity != rules.Error {
			t.Errorf("expected Error severity, got %v for %+v", v.Severity, v)
		}
	}
}

func TestCheckNoSetters_DefaultSeverityWarning(t *testing.T) {
	pkgs := loadSetters(t)
	got := rules.CheckNoSetters(pkgs)

	for _, v := range got {
		if v.Rule == "setter.forbidden" && v.Severity != rules.Warning {
			t.Errorf("expected Warning severity by default, got %v for %+v", v.Severity, v)
		}
	}
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
