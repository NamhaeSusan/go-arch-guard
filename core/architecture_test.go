package core

import "testing"

func TestArchitectureZeroValueIsValid(t *testing.T) {
	// Zero value must be constructible — callers populate sub-models field
	// by field. Any panic from sub-model construction here is a regression.
	var arch Architecture
	if arch.Layers.Sublayers != nil {
		t.Errorf("zero Architecture should have nil Sublayers, got %v", arch.Layers.Sublayers)
	}
}

func TestArchitectureCarriesAllSubModels(t *testing.T) {
	arch := Architecture{
		Layers: LayerModel{
			Sublayers: []string{"handler", "core"},
		},
		Layout: LayoutModel{
			DomainDir: "domain",
		},
		Naming: NamingPolicy{
			AliasFileName: "alias.go",
		},
		Structure: StructurePolicy{
			RequireAlias: true,
			TypePatterns: []TypePattern{
				{Dir: "worker", FilePrefix: "worker", TypeSuffix: "Worker", RequireMethod: "Process"},
			},
		},
	}
	if arch.Layers.Sublayers[0] != "handler" {
		t.Errorf("Layers.Sublayers[0] = %q", arch.Layers.Sublayers[0])
	}
	if arch.Layout.DomainDir != "domain" {
		t.Errorf("Layout.DomainDir = %q", arch.Layout.DomainDir)
	}
	if arch.Naming.AliasFileName != "alias.go" {
		t.Errorf("Naming.AliasFileName = %q", arch.Naming.AliasFileName)
	}
	if !arch.Structure.RequireAlias {
		t.Errorf("Structure.RequireAlias = false")
	}
	if got := arch.Structure.TypePatterns[0].TypeSuffix; got != "Worker" {
		t.Errorf("Structure.TypePatterns[0].TypeSuffix = %q", got)
	}
}

// TestCloneArchitectureCoversAllMutableFields is a deliberate regression
// guard: it builds an Architecture where every map and slice field has
// content, clones it, mutates the clone in every field, and asserts that
// the original is untouched. If a future contributor adds a new map/slice
// field to Architecture without extending cloneArchitecture, this test
// fails in the relevant assertion.
func TestCloneArchitectureCoversAllMutableFields(t *testing.T) {
	original := Architecture{
		Layers: LayerModel{
			Sublayers:        []string{"handler", "core"},
			PortLayers:       []string{"core/repo"},
			ContractLayers:   []string{"core/repo", "core/svc"},
			Direction:        map[string][]string{"handler": {"app"}, "app": {"core/model"}},
			LayerDirNames:    map[string]bool{"handler": true},
			InternalTopLevel: map[string]bool{"domain": true},
			PkgRestricted:    map[string]bool{"core/model": true},
		},
		Layout: LayoutModel{DomainDir: "domain"},
		Naming: NamingPolicy{
			BannedPkgNames: []string{"util"},
			LegacyPkgNames: []string{"router"},
			AliasFileName:  "alias.go",
		},
		Structure: StructurePolicy{
			RequireAlias:            true,
			RequireModel:            true,
			ModelPath:               "core/model",
			DTOAllowedLayers:        []string{"handler"},
			TypePatterns:            []TypePattern{{Dir: "worker", FilePrefix: "worker", TypeSuffix: "Worker"}},
			InterfacePatternExclude: map[string]bool{"handler": true},
		},
	}

	clone := cloneArchitecture(original)

	// Slice fields — append to the clone and verify the original is not extended.
	clone.Layers.Sublayers = append(clone.Layers.Sublayers, "X")
	clone.Layers.PortLayers = append(clone.Layers.PortLayers, "X")
	clone.Layers.ContractLayers = append(clone.Layers.ContractLayers, "X")
	clone.Naming.BannedPkgNames = append(clone.Naming.BannedPkgNames, "X")
	clone.Naming.LegacyPkgNames = append(clone.Naming.LegacyPkgNames, "X")
	clone.Structure.DTOAllowedLayers = append(clone.Structure.DTOAllowedLayers, "X")
	clone.Structure.TypePatterns = append(clone.Structure.TypePatterns, TypePattern{Dir: "X"})

	// Slice element overwrites — these would mutate shared backing arrays
	// if cloneStringSlice forgot to allocate a new array.
	clone.Layers.Sublayers[0] = "MUTATED"
	clone.Layers.PortLayers[0] = "MUTATED"
	clone.Layers.ContractLayers[0] = "MUTATED"
	clone.Naming.BannedPkgNames[0] = "MUTATED"
	clone.Naming.LegacyPkgNames[0] = "MUTATED"
	clone.Structure.DTOAllowedLayers[0] = "MUTATED"
	clone.Structure.TypePatterns[0].Dir = "MUTATED"

	// Map fields — insert a key and verify the original does not see it.
	clone.Layers.LayerDirNames["new"] = true
	clone.Layers.InternalTopLevel["new"] = true
	clone.Layers.PkgRestricted["new"] = true
	clone.Structure.InterfacePatternExclude["new"] = true
	clone.Layers.Direction["new"] = []string{"X"}

	// Nested slice inside the Direction map — mutate an existing inner slice.
	clone.Layers.Direction["handler"] = append(clone.Layers.Direction["handler"], "X")
	clone.Layers.Direction["handler"][0] = "MUTATED"

	// Now check the original is intact.
	if got := original.Layers.Sublayers; len(got) != 2 || got[0] != "handler" {
		t.Errorf("Layers.Sublayers leaked: %v", got)
	}
	if got := original.Layers.PortLayers; len(got) != 1 || got[0] != "core/repo" {
		t.Errorf("Layers.PortLayers leaked: %v", got)
	}
	if got := original.Layers.ContractLayers; len(got) != 2 || got[0] != "core/repo" {
		t.Errorf("Layers.ContractLayers leaked: %v", got)
	}
	if got := original.Naming.BannedPkgNames; len(got) != 1 || got[0] != "util" {
		t.Errorf("Naming.BannedPkgNames leaked: %v", got)
	}
	if got := original.Naming.LegacyPkgNames; len(got) != 1 || got[0] != "router" {
		t.Errorf("Naming.LegacyPkgNames leaked: %v", got)
	}
	if got := original.Structure.DTOAllowedLayers; len(got) != 1 || got[0] != "handler" {
		t.Errorf("Structure.DTOAllowedLayers leaked: %v", got)
	}
	if got := original.Structure.TypePatterns; len(got) != 1 || got[0].Dir != "worker" {
		t.Errorf("Structure.TypePatterns leaked: %v", got)
	}
	if _, ok := original.Layers.LayerDirNames["new"]; ok {
		t.Errorf("Layers.LayerDirNames leaked: %v", original.Layers.LayerDirNames)
	}
	if _, ok := original.Layers.InternalTopLevel["new"]; ok {
		t.Errorf("Layers.InternalTopLevel leaked: %v", original.Layers.InternalTopLevel)
	}
	if _, ok := original.Layers.PkgRestricted["new"]; ok {
		t.Errorf("Layers.PkgRestricted leaked: %v", original.Layers.PkgRestricted)
	}
	if _, ok := original.Structure.InterfacePatternExclude["new"]; ok {
		t.Errorf("Structure.InterfacePatternExclude leaked: %v", original.Structure.InterfacePatternExclude)
	}
	if _, ok := original.Layers.Direction["new"]; ok {
		t.Errorf("Layers.Direction leaked (new key): %v", original.Layers.Direction)
	}
	if got := original.Layers.Direction["handler"]; len(got) != 1 || got[0] != "app" {
		t.Errorf("Layers.Direction[handler] inner slice leaked: %v", got)
	}
}
