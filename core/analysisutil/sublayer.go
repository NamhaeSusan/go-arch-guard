package analysisutil

import (
	"slices"
	"strings"

	"github.com/NamhaeSusan/go-arch-guard/core"
)

func IsKnownSublayer(layers core.LayerModel, name string) bool {
	return slices.Contains(layers.Sublayers, name)
}

func IsPortSublayer(layers core.LayerModel, name string) bool {
	if len(layers.PortLayers) > 0 {
		return slices.Contains(layers.PortLayers, name)
	}
	base := name
	if i := strings.LastIndex(name, "/"); i >= 0 {
		base = name[i+1:]
	}
	return base == "repo" || base == "gateway"
}

func IsContractSublayer(layers core.LayerModel, name string) bool {
	if len(layers.ContractLayers) > 0 || len(layers.PortLayers) > 0 {
		return slices.Contains(layers.ContractLayers, name) ||
			slices.Contains(layers.PortLayers, name)
	}
	if IsPortSublayer(layers, name) {
		return true
	}
	base := name
	if i := strings.LastIndex(name, "/"); i >= 0 {
		base = name[i+1:]
	}
	return base == "svc"
}

func HasPortSublayer(layers core.LayerModel) bool {
	return slices.ContainsFunc(layers.Sublayers, func(sl string) bool {
		return IsPortSublayer(layers, sl)
	})
}

func MatchPortSublayer(layers core.LayerModel, pkgPath string) string {
	return MatchPortSublayerInLayout(layers, core.LayoutModel{InternalRoot: "internal", DomainDir: "domain"}, pkgPath)
}

func MatchPortSublayerInLayout(layers core.LayerModel, layout core.LayoutModel, pkgPath string) string {
	for _, sl := range layers.Sublayers {
		if !IsPortSublayer(layers, sl) {
			continue
		}
		if matchesSublayerBoundary(layout, pkgPath, sl) {
			return sl
		}
	}
	return ""
}

func MatchContractSublayer(layers core.LayerModel, pkgPath string) string {
	return MatchContractSublayerInLayout(layers, core.LayoutModel{InternalRoot: "internal", DomainDir: "domain"}, pkgPath)
}

func MatchContractSublayerInLayout(layers core.LayerModel, layout core.LayoutModel, pkgPath string) string {
	for _, sl := range layers.Sublayers {
		if !IsContractSublayer(layers, sl) {
			continue
		}
		if matchesSublayerBoundary(layout, pkgPath, sl) {
			return sl
		}
	}
	return ""
}

func matchesSublayerBoundary(layout core.LayoutModel, pkgPath, sublayer string) bool {
	internalRoot := layout.InternalRoot
	if internalRoot == "" {
		internalRoot = "internal"
	}
	domainDir := layout.DomainDir
	if domainDir == "" {
		domainDir = "domain"
	}
	parts := strings.Split(pkgPath, "/")
	for i := 0; i+2 < len(parts); i++ {
		if parts[i] == internalRoot && parts[i+1] == domainDir {
			if matchesSublayerParts(parts[i+2:], sublayer) {
				return true
			}
		}
	}
	return false
}

func matchesSublayerParts(relParts []string, sublayer string) bool {
	slParts := strings.Split(sublayer, "/")
	if len(relParts) < 1+len(slParts) {
		return false
	}
	for i, want := range slParts {
		if relParts[1+i] != want {
			return false
		}
	}
	return true
}

func PortSublayerName(layers core.LayerModel) string {
	for _, sl := range layers.Sublayers {
		if IsPortSublayer(layers, sl) {
			return sl
		}
	}
	return "core/repo"
}
