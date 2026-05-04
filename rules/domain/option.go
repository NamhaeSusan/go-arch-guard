package domain

import "github.com/NamhaeSusan/go-arch-guard/core"

type Option func(*ruleConfig)

type ruleConfig struct {
	severity                  core.Severity
	requirePlaceholderAliases bool
}

func WithSeverity(severity core.Severity) Option {
	return func(cfg *ruleConfig) {
		cfg.severity = severity
	}
}

// WithRequirePlaceholderAliases also flags empty alias files for domains
// without an app package. By default, placeholder/model-only domains are
// allowed until they grow an app package.
func WithRequirePlaceholderAliases(require bool) Option {
	return func(cfg *ruleConfig) {
		cfg.requirePlaceholderAliases = require
	}
}

func newConfig(opts []Option, severity core.Severity) ruleConfig {
	cfg := ruleConfig{severity: severity}
	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg
}
