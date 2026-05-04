package naming

import (
	"slices"

	"github.com/NamhaeSusan/go-arch-guard/core"
)

type Option func(*ruleConfig)

type ruleConfig struct {
	severity                core.Severity
	allowedConstructorNames []string
	infraSublayers          []string
}

func WithSeverity(severity core.Severity) Option {
	return func(cfg *ruleConfig) {
		cfg.severity = severity
	}
}

// WithAllowedConstructorNames sets the package-level constructor names that
// infra.constructor-name accepts. Blank names are ignored. If all provided
// names are blank, the rule falls back to "New".
func WithAllowedConstructorNames(names ...string) Option {
	return func(cfg *ruleConfig) {
		cfg.allowedConstructorNames = nonBlankStrings(names)
	}
}

// WithInfraSublayers sets the sublayers checked by infra.constructor-name.
// Values are architecture sublayer paths such as "infra", "adapter", or
// "persistence". Blank names are ignored. If omitted, the rule infers common
// infra-like sublayers from the configured architecture and falls back to
// "infra".
func WithInfraSublayers(sublayers ...string) Option {
	return func(cfg *ruleConfig) {
		cfg.infraSublayers = nonBlankStrings(sublayers)
	}
}

func newConfig(opts []Option, severity core.Severity) ruleConfig {
	cfg := ruleConfig{
		severity:                severity,
		allowedConstructorNames: []string{"New"},
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	if len(cfg.allowedConstructorNames) == 0 {
		cfg.allowedConstructorNames = []string{"New"}
	}
	cfg.allowedConstructorNames = slices.Clip(cfg.allowedConstructorNames)
	cfg.infraSublayers = slices.Clip(cfg.infraSublayers)
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
