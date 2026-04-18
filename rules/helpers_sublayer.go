package rules

import (
	"slices"
	"strings"
)

// isPortSublayerFor reports whether name is a port layer according to the model.
// The model's explicit PortLayers list is checked first; when it does not match,
// the basename fallback to legacy names ("repo", "gateway") is still applied so
// DDD-inherited custom models (which carry PortLayers=["core/repo"]) continue
// to recognize renamed sublayers like "core/ports/repo".
func isPortSublayerFor(m Model, name string) bool {
	if slices.Contains(m.PortLayers, name) {
		return true
	}
	base := name
	if i := strings.LastIndex(name, "/"); i >= 0 {
		base = name[i+1:]
	}
	return base == "repo" || base == "gateway"
}

// isContractSublayerFor reports whether name is a contract layer according to the model.
// Same layering as isPortSublayerFor: explicit ContractLayers match first, then port
// semantics, then basename fallback to "svc" so custom models keep working.
func isContractSublayerFor(m Model, name string) bool {
	if slices.Contains(m.ContractLayers, name) {
		return true
	}
	if isPortSublayerFor(m, name) {
		return true
	}
	base := name
	if i := strings.LastIndex(name, "/"); i >= 0 {
		base = name[i+1:]
	}
	return base == "svc"
}

// hasPortSublayer reports whether the model has any port sublayer.
func hasPortSublayer(m Model) bool {
	return slices.ContainsFunc(m.Sublayers, func(sl string) bool {
		return isPortSublayerFor(m, sl)
	})
}

// matchPortSublayer returns the port sublayer name if pkgPath references one, "" otherwise.
func matchPortSublayer(m Model, pkgPath string) string {
	for _, sl := range m.Sublayers {
		if !isPortSublayerFor(m, sl) {
			continue
		}
		if strings.HasSuffix(pkgPath, "/"+sl) || strings.Contains(pkgPath, "/"+sl+"/") {
			return sl
		}
	}
	return ""
}

// matchContractSublayer returns the contract sublayer name if pkgPath references one, "" otherwise.
func matchContractSublayer(m Model, pkgPath string) string {
	for _, sl := range m.Sublayers {
		if !isContractSublayerFor(m, sl) {
			continue
		}
		if strings.HasSuffix(pkgPath, "/"+sl) || strings.Contains(pkgPath, "/"+sl+"/") {
			return sl
		}
	}
	return ""
}

// portSublayerName returns the first port sublayer name from the model, or "core/repo" as fallback.
func portSublayerName(m Model) string {
	for _, sl := range m.Sublayers {
		if isPortSublayerFor(m, sl) {
			return sl
		}
	}
	return "core/repo"
}
