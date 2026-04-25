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
	for _, sl := range layers.Sublayers {
		if !IsPortSublayer(layers, sl) {
			continue
		}
		if strings.HasSuffix(pkgPath, "/"+sl) || strings.Contains(pkgPath, "/"+sl+"/") {
			return sl
		}
	}
	return ""
}

func MatchContractSublayer(layers core.LayerModel, pkgPath string) string {
	for _, sl := range layers.Sublayers {
		if !IsContractSublayer(layers, sl) {
			continue
		}
		if strings.HasSuffix(pkgPath, "/"+sl) || strings.Contains(pkgPath, "/"+sl+"/") {
			return sl
		}
	}
	return ""
}

func PortSublayerName(layers core.LayerModel) string {
	for _, sl := range layers.Sublayers {
		if IsPortSublayer(layers, sl) {
			return sl
		}
	}
	return "core/repo"
}
