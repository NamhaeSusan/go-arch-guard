package dependency

import (
	"slices"

	"github.com/NamhaeSusan/go-arch-guard/core"
)

type Option func(*ruleConfig)

type ruleConfig struct {
	severity        core.Severity
	inspectedLayers []string
	deniedCalls     []string
	allowedCalls    []string
}

func WithSeverity(severity core.Severity) Option {
	return func(cfg *ruleConfig) {
		cfg.severity = severity
	}
}

// WithInspectedLayers sets the architecture sublayers checked by this rule,
// such as "core/model", "entity", "domain", or "event". Blank names are
// ignored. If omitted, the rule uses Architecture.Layers.PkgRestricted and
// falls back to common domain-core layer names.
func WithInspectedLayers(layers ...string) Option {
	return func(cfg *ruleConfig) {
		cfg.inspectedLayers = nonBlankStrings(layers)
	}
}

// WithDeniedCalls replaces the default side-effect call denylist. Entries are
// fully qualified call IDs such as "time.Now" or prefix patterns ending in
// "*", such as "math/rand.*". Blank entries are ignored.
func WithDeniedCalls(calls ...string) Option {
	return func(cfg *ruleConfig) {
		cfg.deniedCalls = nonBlankStrings(calls)
	}
}

// WithAllowedCalls exempts call IDs from the denylist. Entries use the same
// exact or trailing-* pattern syntax as WithDeniedCalls.
func WithAllowedCalls(calls ...string) Option {
	return func(cfg *ruleConfig) {
		cfg.allowedCalls = nonBlankStrings(calls)
	}
}

func newConfig(opts []Option, severity core.Severity) ruleConfig {
	cfg := ruleConfig{
		severity:    severity,
		deniedCalls: defaultDeniedCalls(),
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	cfg.inspectedLayers = slices.Clip(cfg.inspectedLayers)
	cfg.deniedCalls = slices.Clip(cfg.deniedCalls)
	cfg.allowedCalls = slices.Clip(cfg.allowedCalls)
	return cfg
}

func nonBlankStrings(in []string) []string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}
