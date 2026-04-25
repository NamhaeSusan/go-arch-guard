package core

import "maps"

// Architecture is the team-defined description of a project's layering and
// naming conventions. Presets construct Architecture instances; rules read
// from Context.Arch().
type Architecture struct {
	Layers    LayerModel  // single source of truth for layer vocabulary
	Layout    LayoutModel // internal/ directory topology
	Naming    NamingPolicy
	Structure StructurePolicy
}

// LayerModel owns layer vocabulary. Sublayers is the authoritative single
// source of truth for layer names: every other field that names a layer
// (here, in StructurePolicy, or in any rule) MUST appear in Sublayers, and
// rules read layer names exclusively through ctx.Arch().Layers.Sublayers.
// Architecture.Validate enforces both invariants.
type LayerModel struct {
	// Sublayers is the authoritative single source of truth for layer names.
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

// cloneArchitecture deep-copies every slice and map in an Architecture so
// callers handling the result cannot influence the original.
func cloneArchitecture(a Architecture) Architecture {
	return Architecture{
		Layers: LayerModel{
			Sublayers:        cloneStringSlice(a.Layers.Sublayers),
			Direction:        cloneStringSliceMap(a.Layers.Direction),
			PortLayers:       cloneStringSlice(a.Layers.PortLayers),
			ContractLayers:   cloneStringSlice(a.Layers.ContractLayers),
			PkgRestricted:    cloneBoolMap(a.Layers.PkgRestricted),
			InternalTopLevel: cloneBoolMap(a.Layers.InternalTopLevel),
			LayerDirNames:    cloneBoolMap(a.Layers.LayerDirNames),
		},
		Layout: a.Layout, // value type with only string fields
		Naming: NamingPolicy{
			BannedPkgNames: cloneStringSlice(a.Naming.BannedPkgNames),
			LegacyPkgNames: cloneStringSlice(a.Naming.LegacyPkgNames),
			AliasFileName:  a.Naming.AliasFileName,
		},
		Structure: StructurePolicy{
			RequireAlias:            a.Structure.RequireAlias,
			RequireModel:            a.Structure.RequireModel,
			ModelPath:               a.Structure.ModelPath,
			DTOAllowedLayers:        cloneStringSlice(a.Structure.DTOAllowedLayers),
			TypePatterns:            cloneTypePatterns(a.Structure.TypePatterns),
			InterfacePatternExclude: cloneBoolMap(a.Structure.InterfacePatternExclude),
		},
	}
}

func cloneStringSlice(in []string) []string {
	if in == nil {
		return nil
	}
	out := make([]string, len(in))
	copy(out, in)
	return out
}

func cloneStringSliceMap(in map[string][]string) map[string][]string {
	if in == nil {
		return nil
	}
	out := make(map[string][]string, len(in))
	for k, v := range in {
		out[k] = cloneStringSlice(v)
	}
	return out
}

func cloneBoolMap(in map[string]bool) map[string]bool {
	if in == nil {
		return nil
	}
	out := make(map[string]bool, len(in))
	maps.Copy(out, in)
	return out
}

func cloneTypePatterns(in []TypePattern) []TypePattern {
	if in == nil {
		return nil
	}
	out := make([]TypePattern, len(in))
	copy(out, in)
	return out
}
