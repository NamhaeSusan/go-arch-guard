package rules

import (
	"slices"
	"strings"
)

type internalKind int

const (
	kindDomain internalKind = iota
	kindOrchestration
	kindShared
	kindDomainRoot
	kindCmd
	kindUnclassified
	kindApp       // internal/<AppDir>/...
	kindTransport // internal/<ServerDir>/<proto>/...
)

type classified struct {
	Kind     internalKind
	Domain   string
	Sublayer string
	IsAlias  bool
}

// classifyInternalPackage classifies an internal package path into a single
// canonical kind, domain, and sublayer. It is the single source of truth for
// understanding where a package sits in the architecture.
func classifyInternalPackage(m Model, pkgPath, internalPrefix string) classified {
	if !strings.HasPrefix(pkgPath, internalPrefix) {
		return classified{Kind: kindUnclassified}
	}

	// Shared/pkg check first
	if m.SharedDir != "" && isUnderInternalDir(pkgPath, internalPrefix, m.SharedDir) {
		return classified{Kind: kindShared}
	}

	// App (composition root) check
	if m.AppDir != "" && isUnderInternalDir(pkgPath, internalPrefix, m.AppDir) {
		return classified{Kind: kindApp}
	}

	// Transport (server/<proto>) check — must be under ServerDir/<any-proto>/
	if m.ServerDir != "" {
		rel := strings.TrimPrefix(pkgPath, internalPrefix)
		serverPrefix := m.ServerDir + "/"
		if strings.HasPrefix(rel, serverPrefix) {
			// rel is now "server/<proto>[/...]" — must have at least one proto segment
			after := strings.TrimPrefix(rel, serverPrefix)
			if after != "" {
				return classified{Kind: kindTransport}
			}
		}
	}

	// Orchestration check
	if m.OrchestrationDir != "" && isUnderInternalDir(pkgPath, internalPrefix, m.OrchestrationDir) {
		return classified{Kind: kindOrchestration}
	}

	// Domain-based layout
	if m.DomainDir != "" {
		domain := identifyDomainWith(m, pkgPath, internalPrefix)
		if domain == "" {
			return classified{Kind: kindUnclassified}
		}
		sub := identifySublayerWith(m, pkgPath, internalPrefix, domain)
		if sub == "" {
			// Domain root (alias package)
			return classified{Kind: kindDomainRoot, Domain: domain, IsAlias: true}
		}
		return classified{Kind: kindDomain, Domain: domain, Sublayer: sub}
	}

	// Flat layout: check if it's a known sublayer
	rel := strings.TrimPrefix(pkgPath, internalPrefix)
	parts := strings.SplitN(rel, "/", 2)
	if parts[0] == "" {
		return classified{Kind: kindUnclassified}
	}
	if slices.Contains(m.Sublayers, parts[0]) {
		return classified{Kind: kindDomain, Sublayer: parts[0]}
	}
	return classified{Kind: kindUnclassified}
}
