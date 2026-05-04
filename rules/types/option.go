package types

import (
	"slices"

	"github.com/NamhaeSusan/go-arch-guard/core"
)

type Option func(*ruleConfig)

type ruleConfig struct {
	severity         core.Severity
	inspectedLayers  []string
	allowedPaths     []string
	allowedFunctions []string
}

func WithSeverity(severity core.Severity) Option {
	return func(cfg *ruleConfig) {
		cfg.severity = severity
	}
}

// WithInspectedLayers sets the architecture sublayers checked by this rule.
// Examples include "core/model", "core/svc", "event", "app", "entity",
// "usecase", "domain", "core", and "application". Blank names are ignored.
func WithInspectedLayers(layers ...string) Option {
	return func(cfg *ruleConfig) {
		cfg.inspectedLayers = nonBlankStrings(layers)
	}
}

// WithAllowedPaths exempts project-relative file or subtree patterns. Patterns
// may end in "/..." to match a directory and all descendants.
func WithAllowedPaths(paths ...string) Option {
	return func(cfg *ruleConfig) {
		cfg.allowedPaths = nonBlankStrings(paths)
	}
}

// WithAllowedFunctions exempts function names. Patterns may end in "*" to
// allow families such as Must* helpers.
func WithAllowedFunctions(names ...string) Option {
	return func(cfg *ruleConfig) {
		cfg.allowedFunctions = nonBlankStrings(names)
	}
}

func newConfig(opts []Option, severity core.Severity) ruleConfig {
	cfg := ruleConfig{severity: severity}
	for _, opt := range opts {
		opt(&cfg)
	}
	cfg.inspectedLayers = slices.Clip(cfg.inspectedLayers)
	cfg.allowedPaths = slices.Clip(cfg.allowedPaths)
	cfg.allowedFunctions = slices.Clip(cfg.allowedFunctions)
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
