package rules

import (
	"slices"
	"testing"
)

func TestDDD_ReturnsValidModel(t *testing.T) {
	m := DDD()
	if len(m.Sublayers) == 0 {
		t.Fatal("DDD model must have sublayers")
	}
	if len(m.Direction) == 0 {
		t.Fatal("DDD model must have direction map")
	}
	if !m.RequireAlias {
		t.Error("DDD model must require alias")
	}
	if !m.RequireModel {
		t.Error("DDD model must require domain model")
	}
	if m.DomainDir != "domain" {
		t.Errorf("DomainDir = %q, want %q", m.DomainDir, "domain")
	}
	if m.OrchestrationDir != "orchestration" {
		t.Errorf("OrchestrationDir = %q, want %q", m.OrchestrationDir, "orchestration")
	}
	if m.SharedDir != "pkg" {
		t.Errorf("SharedDir = %q, want %q", m.SharedDir, "pkg")
	}
}

func TestCleanArch_ReturnsValidModel(t *testing.T) {
	m := CleanArch()
	if len(m.Sublayers) == 0 {
		t.Fatal("CleanArch model must have sublayers")
	}
	if !m.InternalTopLevel["domain"] {
		t.Error("CleanArch must allow domain at top level")
	}
	if m.RequireAlias {
		t.Error("CleanArch should not require alias")
	}
	if m.RequireModel {
		t.Error("CleanArch should not require domain model")
	}
}

func TestNewModel_CustomOverrides(t *testing.T) {
	m := NewModel(
		WithDomainDir("module"),
		WithSharedDir("lib"),
		WithSublayers([]string{"handler", "usecase", "entity"}),
		WithDirection(map[string][]string{
			"handler": {"usecase"},
			"usecase": {"entity"},
			"entity":  {},
		}),
	)
	if m.DomainDir != "module" {
		t.Errorf("DomainDir = %q, want %q", m.DomainDir, "module")
	}
	if m.SharedDir != "lib" {
		t.Errorf("SharedDir = %q, want %q", m.SharedDir, "lib")
	}
	if len(m.Sublayers) != 3 {
		t.Errorf("Sublayers count = %d, want 3", len(m.Sublayers))
	}
	if !m.InternalTopLevel["module"] {
		t.Error("InternalTopLevel must include custom DomainDir")
	}
	if !m.InternalTopLevel["lib"] {
		t.Error("InternalTopLevel must include custom SharedDir")
	}
}

func TestNewModel_StartsFromDDD(t *testing.T) {
	m := NewModel()
	ddd := DDD()
	if len(m.Sublayers) != len(ddd.Sublayers) {
		t.Error("NewModel with no options must equal DDD()")
	}
}

func TestWithModel_SetsConfigModel(t *testing.T) {
	m := CleanArch()
	cfg := NewConfig(WithModel(m))
	got := cfg.model()
	if len(got.Sublayers) != len(m.Sublayers) {
		t.Errorf("WithModel did not apply: sublayers count %d vs %d", len(got.Sublayers), len(m.Sublayers))
	}
}

func TestConfig_DefaultModel_IsDDD(t *testing.T) {
	cfg := NewConfig()
	got := cfg.model()
	ddd := DDD()
	if len(got.Sublayers) != len(ddd.Sublayers) {
		t.Error("default model must be DDD")
	}
}

func TestModelConsistency(t *testing.T) {
	for _, tc := range []struct {
		name  string
		model Model
	}{
		{"DDD", DDD()},
		{"CleanArch", CleanArch()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			validateModelConsistency(t, tc.model)
		})
	}
}

func validateModelConsistency(t *testing.T, m Model) {
	t.Helper()
	for _, sl := range m.Sublayers {
		if _, ok := m.Direction[sl]; !ok {
			t.Errorf("sublayer %q not in Direction map", sl)
		}
	}
	for key := range m.Direction {
		found := slices.Contains(m.Sublayers, key)
		if !found {
			t.Errorf("Direction key %q not in Sublayers", key)
		}
	}
	if !m.InternalTopLevel[m.DomainDir] {
		t.Errorf("InternalTopLevel missing DomainDir %q", m.DomainDir)
	}
	if !m.InternalTopLevel[m.SharedDir] {
		t.Errorf("InternalTopLevel missing SharedDir %q", m.SharedDir)
	}
}
