package core

// Architecture is the team-defined description of a project's layering and
// naming conventions. Presets construct Architecture instances; rules read
// from Context.Arch().
type Architecture struct {
	Layers    LayerModel  // single source of truth for layer vocabulary
	Layout    LayoutModel // internal/ directory topology
	Naming    NamingPolicy
	Structure StructurePolicy
}

// LayerModel owns layer vocabulary. Sublayers is authoritative; every
// other field that names a layer (here or in StructurePolicy) MUST appear
// in Sublayers — Architecture.Validate enforces this.
type LayerModel struct {
	Sublayers        []string
	Direction        map[string][]string
	PortLayers       []string        // pure interface layers (repo, gateway)
	ContractLayers   []string        // ⊇ PortLayers (port + svc-like layers)
	PkgRestricted    map[string]bool // sublayers that the shared pkg/ tree must not import from
	InternalTopLevel map[string]bool // top-level dirs allowed under internal/
	LayerDirNames    map[string]bool // basename hints used by placement rules
}

// LayoutModel describes the internal/ directory topology. Empty fields
// disable the corresponding classification (e.g. flat layouts leave
// DomainDir == "").
type LayoutModel struct {
	DomainDir        string
	OrchestrationDir string
	SharedDir        string
	AppDir           string
	ServerDir        string
}

// NamingPolicy carries naming-only conventions. Layer names are NOT
// duplicated here — rules read ctx.Arch().Layers.Sublayers.
type NamingPolicy struct {
	BannedPkgNames []string
	LegacyPkgNames []string
	AliasFileName  string
}

// StructurePolicy carries placement and structure conventions.
type StructurePolicy struct {
	RequireAlias            bool
	RequireModel            bool
	ModelPath               string
	DTOAllowedLayers        []string
	TypePatterns            []TypePattern
	InterfacePatternExclude map[string]bool // sublayer names; validated against Layers.Sublayers
}

// TypePattern is an AST-based naming/structure convention for a directory.
type TypePattern struct {
	Dir           string
	FilePrefix    string
	TypeSuffix    string
	RequireMethod string
}
