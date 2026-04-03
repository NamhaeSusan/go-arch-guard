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
	allowed := m.Direction["adapter"]
	if len(allowed) != 2 {
		t.Errorf("adapter allowed imports = %v, want [port domain]", allowed)
	}
}

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
	if m.ModelPath != "core" {
		t.Errorf("ModelPath = %q, want %q", m.ModelPath, "core")
	}
	if !m.PkgRestricted["core"] {
		t.Error("core sublayer must be pkg-restricted")
	}
	if m.DomainDir != "domain" {
		t.Errorf("DomainDir = %q, want %q", m.DomainDir, "domain")
	}
	allowed := m.Direction["infrastructure"]
	if len(allowed) != 1 || allowed[0] != "core" {
		t.Errorf("infrastructure allowed = %v, want [core]", allowed)
	}
}

func TestConsumerWorker_ReturnsValidModel(t *testing.T) {
	m := ConsumerWorker()
	if len(m.Sublayers) != 4 {
		t.Fatalf("ConsumerWorker sublayer count = %d, want 4", len(m.Sublayers))
	}
	if m.DomainDir != "" {
		t.Errorf("DomainDir = %q, want empty (flat layout)", m.DomainDir)
	}
	if m.OrchestrationDir != "" {
		t.Errorf("OrchestrationDir = %q, want empty", m.OrchestrationDir)
	}
	if m.SharedDir != "pkg" {
		t.Errorf("SharedDir = %q, want %q", m.SharedDir, "pkg")
	}
	if m.RequireAlias {
		t.Error("ConsumerWorker should not require alias")
	}
	if m.RequireModel {
		t.Error("ConsumerWorker should not require model")
	}
	if m.ModelPath != "model" {
		t.Errorf("ModelPath = %q, want %q", m.ModelPath, "model")
	}
	if !m.PkgRestricted["model"] {
		t.Error("model sublayer must be pkg-restricted")
	}
	for _, layer := range []string{"worker", "service", "store", "model", "pkg"} {
		if !m.InternalTopLevel[layer] {
			t.Errorf("InternalTopLevel missing %q", layer)
		}
	}
	if len(m.InternalTopLevel) != 5 {
		t.Errorf("InternalTopLevel has %d entries, want 5", len(m.InternalTopLevel))
	}
	workerAllowed := m.Direction["worker"]
	if len(workerAllowed) != 2 {
		t.Errorf("worker allowed = %v, want [service model]", workerAllowed)
	}
	modelAllowed := m.Direction["model"]
	if len(modelAllowed) != 0 {
		t.Errorf("model allowed = %v, want []", modelAllowed)
	}
	if len(m.TypePatterns) != 1 {
		t.Fatalf("TypePatterns count = %d, want 1", len(m.TypePatterns))
	}
	tp := m.TypePatterns[0]
	if tp.Dir != "worker" || tp.FilePrefix != "worker" || tp.TypeSuffix != "Worker" || tp.RequireMethod != "Process" {
		t.Errorf("TypePattern = %+v, unexpected", tp)
	}
}

func TestBatch_ReturnsValidModel(t *testing.T) {
	m := Batch()
	if len(m.Sublayers) != 4 {
		t.Fatalf("Batch sublayer count = %d, want 4", len(m.Sublayers))
	}
	if m.DomainDir != "" {
		t.Errorf("DomainDir = %q, want empty (flat layout)", m.DomainDir)
	}
	if m.OrchestrationDir != "" {
		t.Errorf("OrchestrationDir = %q, want empty", m.OrchestrationDir)
	}
	if m.SharedDir != "pkg" {
		t.Errorf("SharedDir = %q, want %q", m.SharedDir, "pkg")
	}
	if m.RequireAlias {
		t.Error("Batch should not require alias")
	}
	if m.RequireModel {
		t.Error("Batch should not require model")
	}
	if m.ModelPath != "model" {
		t.Errorf("ModelPath = %q, want %q", m.ModelPath, "model")
	}
	if !m.PkgRestricted["model"] {
		t.Error("model sublayer must be pkg-restricted")
	}
	for _, layer := range []string{"job", "service", "store", "model", "pkg"} {
		if !m.InternalTopLevel[layer] {
			t.Errorf("InternalTopLevel missing %q", layer)
		}
	}
	if len(m.InternalTopLevel) != 5 {
		t.Errorf("InternalTopLevel has %d entries, want 5", len(m.InternalTopLevel))
	}
	jobAllowed := m.Direction["job"]
	if len(jobAllowed) != 2 {
		t.Errorf("job allowed = %v, want [service model]", jobAllowed)
	}
	modelAllowed := m.Direction["model"]
	if len(modelAllowed) != 0 {
		t.Errorf("model allowed = %v, want []", modelAllowed)
	}
	if len(m.TypePatterns) != 1 {
		t.Fatalf("TypePatterns count = %d, want 1", len(m.TypePatterns))
	}
	tp := m.TypePatterns[0]
	if tp.Dir != "job" || tp.FilePrefix != "job" || tp.TypeSuffix != "Job" || tp.RequireMethod != "Run" {
		t.Errorf("TypePattern = %+v, unexpected", tp)
	}
}

func TestEventPipeline_ReturnsValidModel(t *testing.T) {
	m := EventPipeline()
	if len(m.Sublayers) != 7 {
		t.Fatalf("EventPipeline sublayer count = %d, want 7", len(m.Sublayers))
	}
	if m.DomainDir != "" {
		t.Errorf("DomainDir = %q, want empty (flat layout)", m.DomainDir)
	}
	if m.OrchestrationDir != "" {
		t.Errorf("OrchestrationDir = %q, want empty", m.OrchestrationDir)
	}
	if m.SharedDir != "pkg" {
		t.Errorf("SharedDir = %q, want %q", m.SharedDir, "pkg")
	}
	if m.RequireAlias {
		t.Error("EventPipeline should not require alias")
	}
	if m.RequireModel {
		t.Error("EventPipeline should not require model")
	}
	if m.ModelPath != "model" {
		t.Errorf("ModelPath = %q, want %q", m.ModelPath, "model")
	}
	if !m.PkgRestricted["model"] {
		t.Error("model sublayer must be pkg-restricted")
	}
	if !m.PkgRestricted["event"] {
		t.Error("event sublayer must be pkg-restricted")
	}
	for _, layer := range []string{"command", "aggregate", "event", "projection", "eventstore", "readstore", "model", "pkg"} {
		if !m.InternalTopLevel[layer] {
			t.Errorf("InternalTopLevel missing %q", layer)
		}
	}
	if len(m.InternalTopLevel) != 8 {
		t.Errorf("InternalTopLevel has %d entries, want 8", len(m.InternalTopLevel))
	}
	cmdAllowed := m.Direction["command"]
	if len(cmdAllowed) != 3 {
		t.Errorf("command allowed = %v, want [aggregate eventstore model]", cmdAllowed)
	}
	aggAllowed := m.Direction["aggregate"]
	if len(aggAllowed) != 2 {
		t.Errorf("aggregate allowed = %v, want [event model]", aggAllowed)
	}
	modelAllowed := m.Direction["model"]
	if len(modelAllowed) != 0 {
		t.Errorf("model allowed = %v, want []", modelAllowed)
	}
	if len(m.TypePatterns) != 2 {
		t.Fatalf("TypePatterns count = %d, want 2", len(m.TypePatterns))
	}
	tp0 := m.TypePatterns[0]
	if tp0.Dir != "command" || tp0.FilePrefix != "command" || tp0.TypeSuffix != "Command" || tp0.RequireMethod != "Execute" {
		t.Errorf("TypePattern[0] = %+v, unexpected", tp0)
	}
	tp1 := m.TypePatterns[1]
	if tp1.Dir != "aggregate" || tp1.FilePrefix != "aggregate" || tp1.TypeSuffix != "Aggregate" || tp1.RequireMethod != "Apply" {
		t.Errorf("TypePattern[1] = %+v, unexpected", tp1)
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
		t.Error("NewModel with no options must have same sublayer count as DDD()")
	}
	if m.DomainDir != ddd.DomainDir {
		t.Errorf("DomainDir = %q, want %q", m.DomainDir, ddd.DomainDir)
	}
	if m.RequireAlias != ddd.RequireAlias {
		t.Errorf("RequireAlias = %v, want %v", m.RequireAlias, ddd.RequireAlias)
	}
}

func TestWithModel_SetsConfigModel(t *testing.T) {
	m := CleanArch()
	cfg := NewConfig(WithModel(m))
	got := cfg.model()
	if got.RequireAlias != false {
		t.Error("WithModel did not apply: expected RequireAlias=false for CleanArch")
	}
	if !slices.Contains(got.Sublayers, "usecase") {
		t.Error("WithModel did not apply: expected 'usecase' sublayer for CleanArch")
	}
}

func TestConfig_DefaultModel_IsDDD(t *testing.T) {
	cfg := NewConfig()
	got := cfg.model()
	if !got.RequireAlias {
		t.Error("default model must require alias (DDD)")
	}
	if !slices.Contains(got.Sublayers, "core/model") {
		t.Error("default model must have core/model sublayer (DDD)")
	}
}

func TestNewModel_OrchestrationDirPropagation(t *testing.T) {
	m := NewModel(WithOrchestrationDir("workflow"))
	if !m.InternalTopLevel["workflow"] {
		t.Error("InternalTopLevel must include custom OrchestrationDir")
	}
	if m.InternalTopLevel["orchestration"] {
		t.Error("InternalTopLevel must not include old OrchestrationDir after override")
	}
}

func TestInterfacePatternExclude(t *testing.T) {
	tests := []struct {
		name     string
		model    Model
		expected []string
	}{
		{"DDD", DDD(), []string{"handler", "app", "core/model", "event"}},
		{"CleanArch", CleanArch(), []string{"handler", "entity"}},
		{"Layered", Layered(), []string{"handler", "model"}},
		{"Hexagonal", Hexagonal(), []string{"handler", "domain"}},
		{"ModularMonolith", ModularMonolith(), []string{"api", "core"}},
		{"ConsumerWorker", ConsumerWorker(), []string{"model", "worker"}},
		{"Batch", Batch(), []string{"model", "job"}},
		{"EventPipeline", EventPipeline(), []string{"model", "event", "command", "aggregate"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for _, layer := range tc.expected {
				if !tc.model.InterfacePatternExclude[layer] {
					t.Errorf("InterfacePatternExclude missing %q", layer)
				}
			}
		})
	}
}

func TestModelConsistency(t *testing.T) {
	for _, tc := range []struct {
		name  string
		model Model
	}{
		{"DDD", DDD()},
		{"CleanArch", CleanArch()},
		{"Layered", Layered()},
		{"Hexagonal", Hexagonal()},
		{"ModularMonolith", ModularMonolith()},
		{"ConsumerWorker", ConsumerWorker()},
		{"Batch", Batch()},
		{"EventPipeline", EventPipeline()},
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
	if m.DomainDir != "" && !m.InternalTopLevel[m.DomainDir] {
		t.Errorf("InternalTopLevel missing DomainDir %q", m.DomainDir)
	}
	if !m.InternalTopLevel[m.SharedDir] {
		t.Errorf("InternalTopLevel missing SharedDir %q", m.SharedDir)
	}
}
