package core

import "maps"

// Architecture is the team-defined description of a project's layering and
// naming conventions. Presets construct Architecture instances; rules read
// from Context.Arch(). The vocabulary that names layers lives on
// LayerModel — see its godoc for the Sublayers vs LayerDirNames split.
type Architecture struct {
	Layers    LayerModel  // layer vocabulary (paths and basenames)
	Layout    LayoutModel // internal/ directory topology
	Naming    NamingPolicy
	Structure StructurePolicy
}

// LayerModel owns layer vocabulary. Two complementary fields name layers
// for different consumers:
//
//   - Sublayers carries full layer paths ("core/repo", "core/svc", "handler").
//     This is authoritative for direction-aware rules, port/contract sublayer
//     matching, and domain isolation. Direction, PortLayers, ContractLayers,
//     PkgRestricted, and StructurePolicy.InterfacePatternExclude entries MUST
//     reference values that appear here.
//
//   - LayerDirNames carries basenames ("repo", "svc", "model"). File and
//     directory placement rules use these to recognize a layer directory
//     regardless of nesting depth. The basename "repo" recognizes both
//     internal/<domain>/core/repo/... and a flat internal/repo/... layout.
//
// The two are complementary, not redundant. A typical preset declares
// "core/repo" in Sublayers AND "repo" in LayerDirNames so both kinds of rule
// have the data they need. LayerDirNames entries deliberately do NOT have to
// appear in Sublayers — they are basename hints, not full paths.
type LayerModel struct {
	// Sublayers is the authoritative list of layer paths.
	Sublayers []string
	// Direction maps each Sublayer to the Sublayers it may import.
	Direction map[string][]string
	// PortLayers lists Sublayers that are pure-interface ports (e.g. repo).
	// Every entry must appear in Sublayers.
	PortLayers []string
	// ContractLayers lists Sublayers exposed as cross-domain contracts
	// (typically PortLayers ∪ svc-style layers). Every entry must appear in
	// Sublayers.
	ContractLayers []string
	// PkgRestricted marks Sublayers that the shared pkg/ tree must not
	// import from. Keys must appear in Sublayers.
	PkgRestricted map[string]bool
	// InternalTopLevel lists directory names allowed directly under
	// internal/. Keys are top-level dir names (e.g. "domain", "pkg") and
	// are NOT required to appear in Sublayers.
	InternalTopLevel map[string]bool
	// LayerDirNames is the set of layer basenames recognized by file and
	// directory placement rules. Keys are basenames, NOT full Sublayer
	// paths.
	LayerDirNames map[string]bool
}

// LayoutModel describes the package-root directory topology. Empty fields
// disable the corresponding classification (e.g. flat layouts leave
// DomainDir == "").
type LayoutModel struct {
	// InternalRoot is the project-relative directory under which all
	// rule-managed packages live. Defaults to "internal" when empty;
	// cloneArchitecture normalizes the empty value at construction so
	// rules read this field directly without a per-call default check.
	InternalRoot     string
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
// callers handling the result cannot influence the original. It also
// normalizes Layout.InternalRoot to "internal" when empty, which is the
// single source of truth for that default — every consumer reads the
// normalized value, so no per-call conditional is needed.
func cloneArchitecture(a Architecture) Architecture {
	if a.Layout.InternalRoot == "" {
		a.Layout.InternalRoot = "internal"
	}
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
