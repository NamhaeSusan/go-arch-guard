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

// Item A: flat-layout NewModel must promote Sublayers into InternalTopLevel.
func TestNewModel_FlatLayout_PromotesSublayers(t *testing.T) {
	m := NewModel(
		WithDomainDir(""),
		WithOrchestrationDir(""),
		WithSharedDir("pkg"),
		WithSublayers([]string{"worker", "service", "store", "model"}),
		WithDirection(map[string][]string{
			"worker":  {"service", "model"},
			"service": {"store", "model"},
			"store":   {"model"},
			"model":   {},
		}),
	)
	for _, sl := range []string{"worker", "service", "store", "model", "pkg"} {
		if !m.InternalTopLevel[sl] {
			t.Errorf("flat-layout NewModel: InternalTopLevel missing %q", sl)
		}
	}
	if m.InternalTopLevel["domain"] {
		t.Error("flat-layout NewModel: InternalTopLevel must not contain old DomainDir")
	}
	if m.InternalTopLevel["orchestration"] {
		t.Error("flat-layout NewModel: InternalTopLevel must not contain old OrchestrationDir")
	}
}

// Item A: when DomainDir is set, domain-layout behavior is unchanged.
func TestNewModel_DomainLayout_UnchangedBehavior(t *testing.T) {
	m := NewModel(
		WithDomainDir("domain"),
		WithOrchestrationDir("orchestration"),
		WithSharedDir("pkg"),
		WithSublayers([]string{"handler", "app", "core"}),
		WithDirection(map[string][]string{
			"handler": {"app"},
			"app":     {"core"},
			"core":    {},
		}),
	)
	if !m.InternalTopLevel["domain"] {
		t.Error("domain-layout: InternalTopLevel must include DomainDir")
	}
	if !m.InternalTopLevel["orchestration"] {
		t.Error("domain-layout: InternalTopLevel must include OrchestrationDir")
	}
	if !m.InternalTopLevel["pkg"] {
		t.Error("domain-layout: InternalTopLevel must include SharedDir")
	}
	// Sublayers must NOT be promoted in domain layout.
	if m.InternalTopLevel["handler"] {
		t.Error("domain-layout: InternalTopLevel must not include sublayer 'handler'")
	}
}

// Item B: Model has PortLayers and ContractLayers fields.
func TestModel_PortLayersContractLayersFields(t *testing.T) {
	m := Model{
		PortLayers:     []string{"store", "port"},
		ContractLayers: []string{"store", "port", "svc"},
	}
	if len(m.PortLayers) != 2 {
		t.Errorf("PortLayers len = %d, want 2", len(m.PortLayers))
	}
	if len(m.ContractLayers) != 3 {
		t.Errorf("ContractLayers len = %d, want 3", len(m.ContractLayers))
	}
}

// Item B: WithPortLayers and WithContractLayers options exist.
func TestNewModel_WithPortLayersContractLayers(t *testing.T) {
	m := NewModel(
		WithPortLayers([]string{"store", "port"}),
		WithContractLayers([]string{"store", "port", "svc"}),
	)
	if !slices.Contains(m.PortLayers, "store") {
		t.Error("PortLayers must contain 'store'")
	}
	if !slices.Contains(m.ContractLayers, "svc") {
		t.Error("ContractLayers must contain 'svc'")
	}
}

// Item B: presets populate PortLayers/ContractLayers for DDD.
func TestDDD_HasPortAndContractLayers(t *testing.T) {
	m := DDD()
	if !slices.Contains(m.PortLayers, "core/repo") {
		t.Error("DDD PortLayers must contain 'core/repo'")
	}
	if !slices.Contains(m.ContractLayers, "core/svc") {
		t.Error("DDD ContractLayers must contain 'core/svc'")
	}
}

// Item B: presets populate PortLayers/ContractLayers for CleanArch.
func TestCleanArch_HasPortLayers(t *testing.T) {
	m := CleanArch()
	if !slices.Contains(m.PortLayers, "gateway") {
		t.Error("CleanArch PortLayers must contain 'gateway'")
	}
}

// Item B: Hexagonal uses no explicit PortLayers (falls back to hardcoded defaults).
// The port sublayer in Hexagonal does not follow the DDD repo-file-interface pattern.
func TestHexagonal_PortLayersEmpty(t *testing.T) {
	m := Hexagonal()
	if len(m.PortLayers) != 0 {
		t.Errorf("Hexagonal PortLayers must be empty (use fallback), got %v", m.PortLayers)
	}
}

// Item B: isPortSublayer consults model PortLayers when non-empty.
func TestIsPortSublayer_ConsultsModel(t *testing.T) {
	m := Model{
		Sublayers:  []string{"store", "model"},
		PortLayers: []string{"store"},
	}
	if !isPortSublayerFor(m, "store") {
		t.Error("isPortSublayerFor must return true for 'store' when PortLayers=['store']")
	}
	if isPortSublayerFor(m, "model") {
		t.Error("isPortSublayerFor must return false for 'model' when PortLayers=['store']")
	}
	// legacy default: model without PortLayers falls back to hardcoded names
	mFallback := Model{Sublayers: []string{"core/repo"}}
	if !isPortSublayerFor(mFallback, "core/repo") {
		t.Error("fallback: isPortSublayerFor must return true for 'core/repo' when PortLayers is empty")
	}
}

// Item B: isContractSublayer consults model ContractLayers when non-empty.
func TestIsContractSublayer_ConsultsModel(t *testing.T) {
	m := Model{
		Sublayers:      []string{"store", "svc", "model"},
		PortLayers:     []string{"store"},
		ContractLayers: []string{"store", "svc"},
	}
	if !isContractSublayerFor(m, "store") {
		t.Error("isContractSublayerFor must return true for 'store'")
	}
	if !isContractSublayerFor(m, "svc") {
		t.Error("isContractSublayerFor must return true for 'svc'")
	}
	if isContractSublayerFor(m, "model") {
		t.Error("isContractSublayerFor must return false for 'model'")
	}
	// legacy default: model without ContractLayers falls back to hardcoded names
	mFallback := Model{Sublayers: []string{"core/repo"}}
	if !isContractSublayerFor(mFallback, "core/repo") {
		t.Error("fallback: isContractSublayerFor must return true for 'core/repo'")
	}
}

// Item B: matchPortSublayer uses PortLayers from the model.
func TestMatchPortSublayer_CustomLayer(t *testing.T) {
	m := Model{
		Sublayers:  []string{"store", "model"},
		PortLayers: []string{"store"},
	}
	got := matchPortSublayer(m, "github.com/example/myapp/internal/store")
	if got != "store" {
		t.Errorf("matchPortSublayer = %q, want %q", got, "store")
	}
	got = matchPortSublayer(m, "github.com/example/myapp/internal/model")
	if got != "" {
		t.Errorf("matchPortSublayer = %q, want %q", got, "")
	}
}

// Regression (review #1): custom flat model with a non-empty OrchestrationDir
// must keep that directory allowed at internal/ top level.
func TestNewModel_FlatLayout_PreservesOrchestrationDir(t *testing.T) {
	m := NewModel(
		WithDomainDir(""),
		WithOrchestrationDir("workflow"),
		WithSharedDir("pkg"),
		WithSublayers([]string{"worker", "service", "store", "model"}),
		WithDirection(map[string][]string{
			"worker":  {"service", "model"},
			"service": {"store", "model"},
			"store":   {"model"},
			"model":   {},
		}),
	)
	if !m.InternalTopLevel["workflow"] {
		t.Error("flat-layout NewModel must preserve non-empty OrchestrationDir in InternalTopLevel")
	}
	for _, sl := range []string{"worker", "service", "store", "model", "pkg"} {
		if !m.InternalTopLevel[sl] {
			t.Errorf("flat-layout NewModel: InternalTopLevel missing %q", sl)
		}
	}
}

// Regression (review #2): NewModel inherits PortLayers/ContractLayers from DDD.
// A custom model that renames core/repo to core/ports/repo (without explicitly
// setting the new fields) must still get port/contract semantics via the
// basename fallback.
func TestNewModel_InheritedPortLayers_BasenameFallback(t *testing.T) {
	m := NewModel(
		WithSublayers([]string{"handler", "app", "core/ports/repo", "core/svcs/svc", "core/model"}),
		WithDirection(map[string][]string{
			"handler":         {"app"},
			"app":             {"core/ports/repo", "core/svcs/svc", "core/model"},
			"core/ports/repo": {"core/model"},
			"core/svcs/svc":   {"core/model"},
			"core/model":      {},
		}),
	)
	if !isPortSublayerFor(m, "core/ports/repo") {
		t.Error("basename fallback: 'core/ports/repo' must still be a port sublayer")
	}
	if !isContractSublayerFor(m, "core/ports/repo") {
		t.Error("basename fallback: 'core/ports/repo' must still be a contract sublayer")
	}
	if !isContractSublayerFor(m, "core/svcs/svc") {
		t.Error("basename fallback: 'core/svcs/svc' must still be a contract sublayer")
	}
	if got := matchPortSublayer(m, "example.com/app/internal/domain/order/core/ports/repo"); got != "core/ports/repo" {
		t.Errorf("matchPortSublayer with basename fallback = %q, want %q", got, "core/ports/repo")
	}
}
