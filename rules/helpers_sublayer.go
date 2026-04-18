package rules

import (
	"slices"
	"strings"
)

// isPortSublayerFor reports whether name is a port layer according to the model.
// When m.PortLayers is non-empty it is authoritative; otherwise falls back to
// hardcoded legacy names ("repo", "gateway") for backward compatibility.
func isPortSublayerFor(m Model, name string) bool {
	if len(m.PortLayers) > 0 {
		return slices.Contains(m.PortLayers, name)
	}
	base := name
	if i := strings.LastIndex(name, "/"); i >= 0 {
		base = name[i+1:]
	}
	return base == "repo" || base == "gateway"
}

// isContractSublayerFor reports whether name is a contract layer according to the model.
// When m.ContractLayers is non-empty it is authoritative; otherwise falls back to
// hardcoded legacy names ("repo", "gateway", "svc") for backward compatibility.
func isContractSublayerFor(m Model, name string) bool {
	if len(m.ContractLayers) > 0 {
		return slices.Contains(m.ContractLayers, name)
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
