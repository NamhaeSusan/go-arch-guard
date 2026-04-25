// arch.go defines the default architecture model and provides accessor
// functions used by rule implementations. The model can be overridden
// per-check via WithModel().
package rules

import "github.com/NamhaeSusan/go-arch-guard/core/analysisutil"

// defaultModel is the architecture model used when no WithModel option
// is provided. It equals DDD() — the original hardcoded behavior.
var defaultModel = DDD()

// isKnownSublayerIn reports whether s is a recognised domain sublayer
// in the given model.
func isKnownSublayerIn(m Model, s string) bool {
	return analysisutil.IsKnownSublayer(layerModelFromModel(m), s)
}
