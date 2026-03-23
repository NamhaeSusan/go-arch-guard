// arch.go is the single source of truth for architecture model constants
// shared across rule implementations. When adding a new sublayer or
// changing the import graph, update this file — not the individual rules.
package rules

import "slices"

// knownDomainSublayers is the exhaustive set of sublayer paths recognised
// inside a domain (e.g. "handler", "core/model"). Everything else is
// flagged as unknown by CheckLayerDirection.
var knownDomainSublayers = []string{
	"handler",
	"app",
	"core",
	"core/model",
	"core/repo",
	"core/svc",
	"event",
	"infra",
}

// allowedLayerImports defines the intra-domain import graph.
// Key = source sublayer, value = allowed target sublayers.
var allowedLayerImports = map[string][]string{
	"handler":    {"app"},
	"app":        {"core/model", "core/repo", "core/svc", "event"},
	"core":       {"core/model"},
	"core/model": {},
	"core/repo":  {"core/model"},
	"core/svc":   {"core/model"},
	"event":      {"core/model"},
	"infra":      {"core/repo", "core/model", "event"},
}

// pkgRestrictedSublayers are sublayers that must not import internal/pkg.
var pkgRestrictedSublayers = map[string]bool{
	"core":       true,
	"core/model": true,
	"core/repo":  true,
	"core/svc":   true,
	"event":      true,
}

// allowedInternalTopLevel lists the only packages permitted directly under
// internal/.
var allowedInternalTopLevel = map[string]bool{
	"domain":        true,
	"orchestration": true,
	"pkg":           true,
}

// bannedPackageNames are package names rejected anywhere under internal/.
var bannedPackageNames = []string{
	"util", "common", "misc", "helper", "shared", "services",
}

// legacyPackageNames are package names that trigger a migration warning.
var legacyPackageNames = []string{"router", "bootstrap"}

// layerDirNames identifies directory names that are "layer-like" for the
// naming.no-layer-suffix check. Derived from knownDomainSublayers plus
// common synonyms (controller, entity, persistence, store, service, etc.).
var layerDirNames = map[string]bool{
	// from knownDomainSublayers
	"handler": true, "app": true, "core": true,
	"model": true, "repo": true, "svc": true,
	"event": true, "infra": true,
	// common synonyms
	"service": true, "controller": true,
	"entity": true, "store": true, "persistence": true,
	"domain": true,
}

// isKnownSublayer reports whether s is a recognised domain sublayer.
func isKnownSublayer(s string) bool {
	return slices.Contains(knownDomainSublayers, s)
}
