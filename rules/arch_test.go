package rules

import (
	"slices"
	"strings"
	"testing"
)

func TestArchModelConsistency(t *testing.T) {
	// Every domain sublayer must be a key in allowedLayerImports.
	for _, sl := range knownDomainSublayers {
		if _, ok := allowedLayerImports[sl]; !ok {
			t.Errorf("knownDomainSublayers contains %q but allowedLayerImports does not", sl)
		}
	}

	// layerDirNames must be a superset of knownDomainSublayers leaf names.
	for _, sl := range knownDomainSublayers {
		leaf := sl
		if i := strings.LastIndex(sl, "/"); i >= 0 {
			leaf = sl[i+1:]
		}
		if !layerDirNames[leaf] {
			t.Errorf("knownDomainSublayers leaf %q (from %q) missing in layerDirNames", leaf, sl)
		}
	}

	// Every allowedLayerImports key must be in knownDomainSublayers.
	for key := range allowedLayerImports {
		if !slices.Contains(knownDomainSublayers, key) {
			t.Errorf("allowedLayerImports has key %q not present in knownDomainSublayers", key)
		}
	}
}
