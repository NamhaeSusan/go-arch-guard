package core

import (
	"fmt"
	"sort"
	"strings"
)

// Validate checks that every layer-referencing field names a layer present
// in Layers.Sublayers, that Direction has a key for every sublayer (no
// silent enforcement holes), that PortLayers ⊆ ContractLayers, that
// Sublayers entries are non-empty and unique, and that Direction is a DAG.
//
// Validate is called by Run before any rule executes; presets MAY call it
// at construction time to fail fast.
//
// A zero-value Architecture (no Sublayers, no Direction) IS accepted as
// valid — it represents "no layer policy." Pairing it with a non-empty
// RuleSet is generally a configuration mistake (no rule that reads
// Layers.Sublayers can fire) but Validate cannot detect this without
// inspecting the RuleSet, so it is tolerated.
//
// Layout-level non-empty checks (e.g. "AppDir must be set when an
// app-aware rule is enabled") are NOT performed here because Architecture
// does not know which rules will run. Presets that bundle layout-aware
// rules are responsible for refusing to construct an Architecture whose
// Layout is missing required directories.
func (a Architecture) Validate() error {
	known := make(map[string]bool, len(a.Layers.Sublayers))
	var errs []string

	// Sublayers must be non-empty distinct names (silently allowing
	// "" or duplicates lets typos pass and corrupts the Direction graph).
	for _, l := range a.Layers.Sublayers {
		if l == "" {
			errs = append(errs, "Sublayers contains empty-string entry")
			continue
		}
		if known[l] {
			errs = append(errs, fmt.Sprintf("Sublayers contains duplicate entry %q", l))
		}
		known[l] = true
	}

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

	// Direction must be a DAG. Cycles silently allow bidirectional imports
	// and defeat the purpose of layer-direction enforcement.
	if cycle := findDirectionCycle(a.Layers.Direction); cycle != nil {
		errs = append(errs, fmt.Sprintf("Direction contains cycle %s", strings.Join(cycle, " -> ")))
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

// findDirectionCycle does a DFS for cycles in the Direction adjacency.
// Returns the cycle as a node sequence (starting and ending at the same
// node), or nil if the graph is a DAG. Order of source iteration is
// stabilized by visiting layers in Sublayers order to keep error messages
// deterministic.
func findDirectionCycle(direction map[string][]string) []string {
	const (
		white = 0 // unvisited
		gray  = 1 // on stack
		black = 2 // fully explored
	)
	color := make(map[string]int, len(direction))
	var stack []string

	var dfs func(node string) []string
	dfs = func(node string) []string {
		color[node] = gray
		stack = append(stack, node)
		for _, next := range direction[node] {
			switch color[next] {
			case gray:
				// Cycle: extract the path from `next` to current top.
				start := -1
				for i, n := range stack {
					if n == next {
						start = i
						break
					}
				}
				if start < 0 {
					return nil
				}
				cycle := append([]string{}, stack[start:]...)
				cycle = append(cycle, next)
				return cycle
			case white:
				if c := dfs(next); c != nil {
					return c
				}
			}
		}
		stack = stack[:len(stack)-1]
		color[node] = black
		return nil
	}

	// Sort sources for deterministic error reporting.
	sources := make([]string, 0, len(direction))
	for src := range direction {
		sources = append(sources, src)
	}
	sort.Strings(sources)
	for _, src := range sources {
		if color[src] == white {
			if c := dfs(src); c != nil {
				return c
			}
		}
	}
	return nil
}
