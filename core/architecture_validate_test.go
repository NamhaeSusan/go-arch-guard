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

// TestValidateZeroValueArchitecture locks down the contract that an empty
// Architecture is a legal "no-op" config: callers building one field at a
// time must not see a panic from nil maps/slices, and Validate() must
// succeed when no rules are configured. If a future change makes
// validation stricter for zero values, that change must update this test
// (and document the migration).
func TestValidateZeroValueArchitecture(t *testing.T) {
	var arch Architecture
	if err := arch.Validate(); err != nil {
		t.Fatalf("zero-value Architecture.Validate() = %v, want nil", err)
	}
}

func TestValidateRejectsEmptyLayerDirNamesKey(t *testing.T) {
	a := validArchitecture()
	a.Layers.LayerDirNames = map[string]bool{"": true}
	err := a.Validate()
	if err == nil || !strings.Contains(err.Error(), "LayerDirNames") || !strings.Contains(err.Error(), "empty") {
		t.Errorf("expected empty-key error for LayerDirNames, got %v", err)
	}
}

func TestValidateRejectsEmptyInternalTopLevelKey(t *testing.T) {
	a := validArchitecture()
	a.Layers.InternalTopLevel = map[string]bool{"": true}
	err := a.Validate()
	if err == nil || !strings.Contains(err.Error(), "InternalTopLevel") || !strings.Contains(err.Error(), "empty") {
		t.Errorf("expected empty-key error for InternalTopLevel, got %v", err)
	}
}

func TestValidateAcceptsLayerDirNamesNotInSublayers(t *testing.T) {
	// LayerDirNames carries basenames that intentionally do NOT have to
	// match Sublayers entries — "repo" is a basename even though only
	// "core/repo" lives in Sublayers. Validate must NOT reject this.
	a := validArchitecture()
	a.Layers.LayerDirNames = map[string]bool{"repo": true, "svc": true, "model": true, "handler": true, "app": true}
	if err := a.Validate(); err != nil {
		t.Fatalf("LayerDirNames basenames must be allowed regardless of Sublayers, got %v", err)
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

func TestValidateRejectsEmptyStringSublayer(t *testing.T) {
	a := Architecture{
		Layers: LayerModel{
			Sublayers: []string{"", "svc"},
			Direction: map[string][]string{"": {}, "svc": {}},
		},
	}
	if err := a.Validate(); err == nil || !strings.Contains(err.Error(), "empty-string") {
		t.Fatalf("empty-string sublayer must be rejected, got %v", err)
	}
}

func TestValidateRejectsDuplicateSublayers(t *testing.T) {
	a := Architecture{
		Layers: LayerModel{
			Sublayers: []string{"svc", "svc"},
			Direction: map[string][]string{"svc": {}},
		},
	}
	if err := a.Validate(); err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("duplicate sublayer name must be rejected, got %v", err)
	}
}

func TestValidateRejectsDirectionCycle(t *testing.T) {
	a := Architecture{
		Layers: LayerModel{
			Sublayers: []string{"a", "b"},
			Direction: map[string][]string{
				"a": {"b"},
				"b": {"a"},
			},
		},
	}
	err := a.Validate()
	if err == nil {
		t.Fatalf("Validate() should reject Direction cycle")
	}
	msg := err.Error()
	if !strings.Contains(msg, "cycle") || !strings.Contains(msg, "a") || !strings.Contains(msg, "b") {
		t.Errorf("error should mention cycle and offending layers, got %q", msg)
	}
}

func TestValidateRejectsSelfLoopInDirection(t *testing.T) {
	a := Architecture{
		Layers: LayerModel{
			Sublayers: []string{"a"},
			Direction: map[string][]string{"a": {"a"}},
		},
	}
	if err := a.Validate(); err == nil || !strings.Contains(err.Error(), "cycle") {
		t.Errorf("self-loop Direction[a]=[a] should be rejected as cycle, got %v", err)
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
