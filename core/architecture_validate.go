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
			errs = append(errs, fmt.Sprintf("Direction has unknown source layer %q", src))
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

	if len(errs) == 0 {
		return nil
	}
	sort.Strings(errs)
	return fmt.Errorf("invalid Architecture: %s", strings.Join(errs, "; "))
}
