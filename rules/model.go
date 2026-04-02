package rules

// Model defines the architecture model used by all rule checks.
// Use DDD(), CleanArch(), Layered(), Hexagonal(), ModularMonolith(), or NewModel() to create one.
type Model struct {
	Sublayers        []string
	Direction        map[string][]string
	PkgRestricted    map[string]bool
	InternalTopLevel map[string]bool
	DomainDir        string
	OrchestrationDir string
	SharedDir        string
	RequireAlias     bool
	AliasFileName    string
	RequireModel     bool
	ModelPath        string
	DTOAllowedLayers []string
	BannedPkgNames   []string
	LegacyPkgNames   []string
	LayerDirNames    map[string]bool
	TypePatterns     []TypePattern
}

// TypePattern defines an AST-based naming/structure convention for a directory.
type TypePattern struct {
	Dir           string // target directory under internal/, e.g. "worker"
	FilePrefix    string // required file prefix, e.g. "worker"
	TypeSuffix    string // required exported type suffix, e.g. "Worker"
	RequireMethod string // required method name, e.g. "Process"
}

// ModelOption configures a Model via NewModel.
type ModelOption func(*Model)

// DDD returns the default Domain-Driven Design architecture model.
func DDD() Model {
	return Model{
		Sublayers: []string{
			"handler", "app", "core", "core/model",
			"core/repo", "core/svc", "event", "infra",
		},
		Direction: map[string][]string{
			"handler":    {"app"},
			"app":        {"core/model", "core/repo", "core/svc", "event"},
			"core":       {"core/model"},
			"core/model": {},
			"core/repo":  {"core/model"},
			"core/svc":   {"core/model"},
			"event":      {"core/model"},
			"infra":      {"core/repo", "core/model", "event"},
		},
		PkgRestricted: map[string]bool{
			"core": true, "core/model": true,
			"core/repo": true, "core/svc": true, "event": true,
		},
		InternalTopLevel: map[string]bool{
			"domain": true, "orchestration": true, "pkg": true,
		},
		DomainDir:        "domain",
		OrchestrationDir: "orchestration",
		SharedDir:        "pkg",
		RequireAlias:     true,
		AliasFileName:    "alias.go",
		RequireModel:     true,
		ModelPath:        "core/model",
		DTOAllowedLayers: []string{"handler", "app"},
		BannedPkgNames:   []string{"util", "common", "misc", "helper", "shared", "services"},
		LegacyPkgNames:   []string{"router", "bootstrap"},
		LayerDirNames: map[string]bool{
			"handler": true, "app": true, "core": true,
			"model": true, "repo": true, "svc": true,
			"event": true, "infra": true,
			"service": true, "controller": true,
			"entity": true, "store": true, "persistence": true,
			"domain": true,
		},
	}
}

// CleanArch returns a Clean Architecture model.
func CleanArch() Model {
	return Model{
		Sublayers: []string{
			"handler", "usecase", "entity", "gateway", "infra",
		},
		Direction: map[string][]string{
			"handler": {"usecase"},
			"usecase": {"entity", "gateway"},
			"entity":  {},
			"gateway": {"entity"},
			"infra":   {"gateway", "entity"},
		},
		PkgRestricted: map[string]bool{
			"entity": true,
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
		ModelPath:        "entity",
		DTOAllowedLayers: []string{"handler", "usecase"},
		BannedPkgNames:   []string{"util", "common", "misc", "helper", "shared", "services"},
		LegacyPkgNames:   []string{"router", "bootstrap"},
		LayerDirNames: map[string]bool{
			"handler": true, "usecase": true, "entity": true,
			"gateway": true, "infra": true,
			"service": true, "controller": true,
			"store": true, "persistence": true, "domain": true,
		},
	}
}

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

// ModularMonolith returns a module-based layered architecture model.
func ModularMonolith() Model {
	return Model{
		Sublayers: []string{
			"api", "application", "core", "infrastructure",
		},
		Direction: map[string][]string{
			"api":            {"application"},
			"application":    {"core"},
			"core":           {},
			"infrastructure": {"core"},
		},
		PkgRestricted: map[string]bool{
			"core": true,
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
		ModelPath:        "core",
		DTOAllowedLayers: []string{"api", "application"},
		BannedPkgNames:   []string{"util", "common", "misc", "helper", "shared", "services"},
		LegacyPkgNames:   []string{"router", "bootstrap"},
		LayerDirNames: map[string]bool{
			"api": true, "application": true, "core": true,
			"infrastructure": true,
			"controller":     true, "service": true, "entity": true,
			"store": true, "persistence": true,
		},
	}
}

// ConsumerWorker returns a flat-layout model for Kafka/RabbitMQ consumer projects.
// Flat layout means layers live directly under internal/ (no domain/ directory).
func ConsumerWorker() Model {
	return Model{
		Sublayers: []string{"worker", "service", "store", "model"},
		Direction: map[string][]string{
			"worker":  {"service", "model"},
			"service": {"store", "model"},
			"store":   {"model"},
			"model":   {},
		},
		PkgRestricted: map[string]bool{"model": true},
		InternalTopLevel: map[string]bool{
			"worker": true, "service": true,
			"store": true, "model": true, "pkg": true,
		},
		DomainDir:        "",
		OrchestrationDir: "",
		SharedDir:        "pkg",
		RequireAlias:     false,
		AliasFileName:    "",
		RequireModel:     false,
		ModelPath:        "model",
		DTOAllowedLayers: []string{"worker", "service"},
		BannedPkgNames:   []string{"util", "common", "misc", "helper", "shared", "services"},
		LegacyPkgNames:   []string{"router", "bootstrap"},
		LayerDirNames: map[string]bool{
			"worker": true, "service": true,
			"store": true, "model": true,
		},
		TypePatterns: []TypePattern{
			{Dir: "worker", FilePrefix: "worker", TypeSuffix: "Worker", RequireMethod: "Process"},
		},
	}
}

// NewModel creates a Model starting from DDD defaults, then applies options.
func NewModel(opts ...ModelOption) Model {
	m := DDD()
	for _, o := range opts {
		o(&m)
	}
	m.InternalTopLevel = map[string]bool{
		m.DomainDir:        true,
		m.OrchestrationDir: true,
		m.SharedDir:        true,
	}
	return m
}

func WithSublayers(sublayers []string) ModelOption {
	return func(m *Model) { m.Sublayers = sublayers }
}

func WithDirection(direction map[string][]string) ModelOption {
	return func(m *Model) { m.Direction = direction }
}

func WithPkgRestricted(restricted map[string]bool) ModelOption {
	return func(m *Model) { m.PkgRestricted = restricted }
}

func WithDomainDir(dir string) ModelOption {
	return func(m *Model) { m.DomainDir = dir }
}

func WithOrchestrationDir(dir string) ModelOption {
	return func(m *Model) { m.OrchestrationDir = dir }
}

func WithSharedDir(dir string) ModelOption {
	return func(m *Model) { m.SharedDir = dir }
}

func WithRequireAlias(b bool) ModelOption {
	return func(m *Model) { m.RequireAlias = b }
}

func WithRequireModel(b bool) ModelOption {
	return func(m *Model) { m.RequireModel = b }
}

func WithModelPath(path string) ModelOption {
	return func(m *Model) { m.ModelPath = path }
}

func WithDTOAllowedLayers(layers []string) ModelOption {
	return func(m *Model) { m.DTOAllowedLayers = layers }
}

func WithBannedPkgNames(names []string) ModelOption {
	return func(m *Model) { m.BannedPkgNames = names }
}

func WithLegacyPkgNames(names []string) ModelOption {
	return func(m *Model) { m.LegacyPkgNames = names }
}

func WithAliasFileName(name string) ModelOption {
	return func(m *Model) { m.AliasFileName = name }
}

func WithLayerDirNames(names map[string]bool) ModelOption {
	return func(m *Model) { m.LayerDirNames = names }
}
