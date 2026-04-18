package rules

import (
	"slices"
	"strings"
)

// isPortSublayer reports whether the sublayer name is a port/contract layer
// (pure interface definitions like repo, gateway).
func isPortSublayer(name string) bool {
	base := name
	if i := strings.LastIndex(name, "/"); i >= 0 {
		base = name[i+1:]
	}
	return base == "repo" || base == "gateway"
}

// isContractSublayer reports whether the sublayer name is a contract layer
// (port/repo + service interfaces like svc).
func isContractSublayer(name string) bool {
	if isPortSublayer(name) {
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
	return slices.ContainsFunc(m.Sublayers, isPortSublayer)
}

// matchPortSublayer returns the port sublayer name if pkgPath references one, "" otherwise.
func matchPortSublayer(m Model, pkgPath string) string {
	for _, sl := range m.Sublayers {
		if !isPortSublayer(sl) {
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
		if !isContractSublayer(sl) {
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
		if isPortSublayer(sl) {
			return sl
		}
	}
	return "core/repo"
}
