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
