package structural

import (
	"slices"

	"github.com/NamhaeSusan/go-arch-guard/core"
)

type Option func(*ruleConfig)

type ruleConfig struct {
	severity                  core.Severity
	repoPortSuffixes          []string
	requirePlaceholderAliases bool
}

func WithSeverity(severity core.Severity) Option {
	return func(cfg *ruleConfig) {
		cfg.severity = severity
	}
}

// WithRepoPortSuffixes sets the suffix list used by
// structural.NewRepoFileInterface to detect repository-port interface names.
// Default is ["Repository", "Repo"]; pass alternates such as "Gateway",
// "Adapter", or "Port" to match a different vocabulary.
//
// Other structural.New*() rules silently ignore this option to keep the
// option API uniform across the package. Empty/nil suffixes are treated as
// "use the default".
func WithRepoPortSuffixes(suffixes ...string) Option {
	return func(cfg *ruleConfig) {
		cfg.repoPortSuffixes = nonBlankStrings(suffixes)
	}
}

// WithRequirePlaceholderAliases also flags empty alias files for domains
// without an app package when using structural.NewNonEmptyAlias. By default,
// placeholder/model-only domains are allowed until they grow an app package.
//
// Other structural.New*() rules silently ignore this option to keep the
// option API uniform across the package.
func WithRequirePlaceholderAliases(require bool) Option {
	return func(cfg *ruleConfig) {
		cfg.requirePlaceholderAliases = require
	}
}

func nonBlankStrings(in []string) []string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		if s != "" {
			out = append(out, s)
		}
	}
	return slices.Clip(out)
}

func newConfig(opts []Option, severity core.Severity) ruleConfig {
	cfg := ruleConfig{severity: severity}
	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg
}
