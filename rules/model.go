package rules

// Model defines the architecture model used by all rule checks.
// Use DDD(), CleanArch(), or NewModel() to create one.
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
