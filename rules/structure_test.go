package rules_test

import (
	"testing"

	"github.com/kimtaeyun/go-arch-guard/rules"
)

func findViolation(violations []rules.Violation, rule string) *rules.Violation {
	for i := range violations {
		if violations[i].Rule == rule {
			return &violations[i]
		}
	}
	return nil
}

func TestCheckStructure(t *testing.T) {
	t.Run("valid project has no violations", func(t *testing.T) {
		violations := rules.CheckStructure("../testdata/valid")
		if len(violations) > 0 {
			t.Errorf("expected no violations, got %d: %v", len(violations), violations)
		}
	})

	t.Run("detects banned package names", func(t *testing.T) {
		violations := rules.CheckStructure("../testdata/invalid")
		found := findViolation(violations, "structure.banned-package")
		if found == nil {
			t.Error("expected banned-package violation for 'util'")
		}
	})

	t.Run("detects missing model.go in domain", func(t *testing.T) {
		violations := rules.CheckStructure("../testdata/invalid")
		found := findViolation(violations, "structure.domain-model-required")
		if found == nil {
			t.Error("expected domain-model-required violation for 'order'")
		}
	})

	t.Run("detects DTO in domain", func(t *testing.T) {
		violations := rules.CheckStructure("../testdata/invalid")
		found := findViolation(violations, "structure.dto-placement")
		if found == nil {
			t.Error("expected dto-placement violation in domain/user/")
		}
	})

	t.Run("exclude skips matching paths", func(t *testing.T) {
		violations := rules.CheckStructure("../testdata/invalid",
			rules.WithExclude("internal/util/..."))
		for _, v := range violations {
			if v.Rule == "structure.banned-package" {
				t.Error("expected util violation to be excluded")
			}
		}
	})

	t.Run("warning severity sets violation level", func(t *testing.T) {
		violations := rules.CheckStructure("../testdata/invalid",
			rules.WithSeverity(rules.Warning))
		for _, v := range violations {
			if v.Severity != rules.Warning {
				t.Errorf("expected Warning, got %v", v.Severity)
			}
		}
	})
}
