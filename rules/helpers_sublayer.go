package rules

import (
	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/core/analysisutil"
)

// isPortSublayerFor reports whether name is a port layer according to the model.
// When m.PortLayers is non-empty it is authoritative (exact match only, no
// basename leakage) so callers that set WithPortLayers get exactly what they
// asked for. When it is empty, fall back to the hardcoded basename heuristic
// ("repo", "gateway").
func isPortSublayerFor(m Model, name string) bool {
	return analysisutil.IsPortSublayer(layerModelFromModel(m), name)
}

// isContractSublayerFor reports whether name is a contract layer according to
// the model. Contract ⊇ Port, so when either PortLayers or ContractLayers is
// non-empty the helper unions the two lists (authoritative, no basename leak).
// When both are empty, fall back to the hardcoded basename heuristic
// (port names plus "svc").
func isContractSublayerFor(m Model, name string) bool {
	return analysisutil.IsContractSublayer(layerModelFromModel(m), name)
}

// hasPortSublayer reports whether the model has any port sublayer.
func hasPortSublayer(m Model) bool {
	return analysisutil.HasPortSublayer(layerModelFromModel(m))
}

// matchPortSublayer returns the port sublayer name if pkgPath references one, "" otherwise.
func matchPortSublayer(m Model, pkgPath string) string {
	return analysisutil.MatchPortSublayer(layerModelFromModel(m), pkgPath)
}

// matchContractSublayer returns the contract sublayer name if pkgPath references one, "" otherwise.
func matchContractSublayer(m Model, pkgPath string) string {
	return analysisutil.MatchContractSublayer(layerModelFromModel(m), pkgPath)
}

// portSublayerName returns the first port sublayer name from the model, or "core/repo" as fallback.
func portSublayerName(m Model) string {
	return analysisutil.PortSublayerName(layerModelFromModel(m))
}

func layerModelFromModel(m Model) core.LayerModel {
	return core.LayerModel{
		Sublayers:      m.Sublayers,
		PortLayers:     m.PortLayers,
		ContractLayers: m.ContractLayers,
	}
}
