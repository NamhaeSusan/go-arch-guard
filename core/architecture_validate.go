package core

import (
	"fmt"
	"sort"
	"strings"
)

// Validate checks that every layer-referencing field names a layer present
// in Layers.Sublayers, that Direction has a key for every sublayer (no
// silent enforcement holes), and that PortLayers ⊆ ContractLayers.
//
// Validate is called by Run before any rule executes; presets MAY call it
// at construction time to fail fast.
//
// Layout-level non-empty checks (e.g. "AppDir must be set when an
// app-aware rule is enabled") are NOT performed here because Architecture
// does not know which rules will run. Presets that bundle layout-aware
// rules are responsible for refusing to construct an Architecture whose
// Layout is missing required directories.
func (a Architecture) Validate() error {
	known := make(map[string]bool, len(a.Layers.Sublayers))
	for _, l := range a.Layers.Sublayers {
		known[l] = true
	}

	var errs []string

	// Direction: every Sublayer must be a key.
	for _, l := range a.Layers.Sublayers {
		if _, ok := a.Layers.Direction[l]; !ok {
			errs = append(errs, fmt.Sprintf("Direction missing key for sublayer %q (silent enforcement hole)", l))
		}
	}
	// Direction values must be known sublayers.
	for src, targets := range a.Layers.Direction {
		if !known[src] {
			errs = append(errs, fmt.Sprintf("Direction[%q] has unknown source layer (key not in Sublayers)", src))
		}
		for _, t := range targets {
			if !known[t] {
				errs = append(errs, fmt.Sprintf("Direction[%q] references unknown layer %q", src, t))
			}
		}
	}

	// PortLayers / ContractLayers membership.
	for _, l := range a.Layers.PortLayers {
		if !known[l] {
			errs = append(errs, fmt.Sprintf("PortLayers references unknown layer %q", l))
		}
	}
	for _, l := range a.Layers.ContractLayers {
		if !known[l] {
			errs = append(errs, fmt.Sprintf("ContractLayers references unknown layer %q", l))
		}
	}
	// PortLayers ⊆ ContractLayers.
	contract := make(map[string]bool, len(a.Layers.ContractLayers))
	for _, l := range a.Layers.ContractLayers {
		contract[l] = true
	}
	for _, l := range a.Layers.PortLayers {
		if !contract[l] {
			errs = append(errs, fmt.Sprintf("PortLayers must be a subset of ContractLayers (missing %q)", l))
		}
	}

	// PkgRestricted keys.
	for l := range a.Layers.PkgRestricted {
		if !known[l] {
			errs = append(errs, fmt.Sprintf("PkgRestricted references unknown layer %q", l))
		}
	}

	// DTOAllowedLayers / InterfacePatternExclude.
	for _, l := range a.Structure.DTOAllowedLayers {
		if !known[l] {
			errs = append(errs, fmt.Sprintf("DTOAllowedLayers references unknown layer %q", l))
		}
	}
	for l := range a.Structure.InterfacePatternExclude {
		if !known[l] {
			errs = append(errs, fmt.Sprintf("InterfacePatternExclude references unknown layer %q", l))
		}
	}

	// TypePatterns: each pattern's Dir must be a declared sublayer.
	for _, tp := range a.Structure.TypePatterns {
		if tp.Dir == "" {
			errs = append(errs, "TypePatterns entry has empty Dir")
			continue
		}
		if !known[tp.Dir] {
			errs = append(errs, fmt.Sprintf("TypePatterns references unknown layer %q", tp.Dir))
		}
	}

	if len(errs) == 0 {
		return nil
	}
	sort.Strings(errs)
	return fmt.Errorf("invalid Architecture: %s", strings.Join(errs, "; "))
}

// Validate is the package-level form of Architecture.Validate. Presets that
// prefer the functional form (e.g. `core.Validate(arch)`) can call this.
func Validate(a Architecture) error { return a.Validate() }
