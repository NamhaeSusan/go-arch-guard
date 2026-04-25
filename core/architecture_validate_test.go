package core

import (
	"strings"
	"testing"
)

func validArchitecture() Architecture {
	return Architecture{
		Layers: LayerModel{
			Sublayers: []string{"handler", "app", "core/repo", "core/svc", "core/model"},
			Direction: map[string][]string{
				"handler":    {"app"},
				"app":        {"core/repo", "core/svc", "core/model"},
				"core/repo":  {"core/model"},
				"core/svc":   {"core/model"},
				"core/model": {},
			},
			PortLayers:     []string{"core/repo"},
			ContractLayers: []string{"core/repo", "core/svc"},
			PkgRestricted:  map[string]bool{"core/repo": true},
		},
		Structure: StructurePolicy{
			DTOAllowedLayers:        []string{"handler", "app"},
			InterfacePatternExclude: map[string]bool{"handler": true},
		},
	}
}

func TestValidateAcceptsValid(t *testing.T) {
	if err := validArchitecture().Validate(); err != nil {
		t.Fatalf("Validate() = %v, want nil", err)
	}
}

func TestValidateRejectsDirectionKeyMissing(t *testing.T) {
	a := validArchitecture()
	delete(a.Layers.Direction, "core/svc")
	err := a.Validate()
	if err == nil || !strings.Contains(err.Error(), "Direction") || !strings.Contains(err.Error(), "core/svc") {
		t.Errorf("expected Direction error mentioning core/svc, got %v", err)
	}
}

func TestValidateRejectsDirectionTargetUnknown(t *testing.T) {
	a := validArchitecture()
	a.Layers.Direction["app"] = append(a.Layers.Direction["app"], "ghost")
	err := a.Validate()
	if err == nil || !strings.Contains(err.Error(), "ghost") {
		t.Errorf("expected unknown-layer error mentioning ghost, got %v", err)
	}
}

func TestValidateRejectsPortNotSubsetOfContract(t *testing.T) {
	a := validArchitecture()
	a.Layers.PortLayers = []string{"core/repo", "core/model"} // model not in ContractLayers
	err := a.Validate()
	if err == nil || !strings.Contains(err.Error(), "PortLayers") || !strings.Contains(err.Error(), "ContractLayers") {
		t.Errorf("expected PortLayers/ContractLayers subset error, got %v", err)
	}
}

func TestValidateRejectsPortLayerNotInSublayers(t *testing.T) {
	a := validArchitecture()
	a.Layers.PortLayers = []string{"ghost"}
	a.Layers.ContractLayers = []string{"ghost"}
	err := a.Validate()
	if err == nil || !strings.Contains(err.Error(), "ghost") {
		t.Errorf("expected unknown-layer error, got %v", err)
	}
}

func TestValidateRejectsDTOAllowedLayerUnknown(t *testing.T) {
	a := validArchitecture()
	a.Structure.DTOAllowedLayers = []string{"ghost"}
	err := a.Validate()
	if err == nil || !strings.Contains(err.Error(), "DTOAllowedLayers") {
		t.Errorf("expected DTOAllowedLayers error, got %v", err)
	}
}

func TestValidateRejectsInterfacePatternExcludeUnknown(t *testing.T) {
	a := validArchitecture()
	a.Structure.InterfacePatternExclude = map[string]bool{"ghost": true}
	err := a.Validate()
	if err == nil || !strings.Contains(err.Error(), "InterfacePatternExclude") {
		t.Errorf("expected InterfacePatternExclude error, got %v", err)
	}
}

func TestValidateRejectsPkgRestrictedUnknown(t *testing.T) {
	a := validArchitecture()
	a.Layers.PkgRestricted = map[string]bool{"ghost": true}
	err := a.Validate()
	if err == nil || !strings.Contains(err.Error(), "PkgRestricted") {
		t.Errorf("expected PkgRestricted error, got %v", err)
	}
}

func TestValidateRejectsDirectionUnknownSourceLayer(t *testing.T) {
	a := validArchitecture()
	a.Layers.Direction["ghost"] = []string{"core/model"}
	err := a.Validate()
	if err == nil || !strings.Contains(err.Error(), "ghost") || !strings.Contains(err.Error(), "Direction") {
		t.Errorf("expected Direction unknown-source error, got %v", err)
	}
}

func TestValidateRejectsContractLayersUnknown(t *testing.T) {
	a := validArchitecture()
	a.Layers.ContractLayers = []string{"core/repo", "core/svc", "ghost"}
	err := a.Validate()
	if err == nil || !strings.Contains(err.Error(), "ContractLayers") || !strings.Contains(err.Error(), "ghost") {
		t.Errorf("expected ContractLayers unknown-layer error, got %v", err)
	}
}

func TestPackageLevelValidateMatchesMethod(t *testing.T) {
	a := validArchitecture()
	if Validate(a) != nil {
		t.Errorf("Validate(valid) should be nil")
	}
	a.Layers.PortLayers = []string{"ghost"}
	a.Layers.ContractLayers = []string{"ghost"}
	if Validate(a) == nil {
		t.Errorf("Validate(invalid) should return error")
	}
}
