package interfaces

import "github.com/NamhaeSusan/go-arch-guard/core"

type Option func(*ruleConfig)

type ruleConfig struct {
	severity   core.Severity
	maxMethods int
}

func WithSeverity(severity core.Severity) Option {
	return func(cfg *ruleConfig) {
		cfg.severity = severity
	}
}

func WithMaxMethods(n int) Option {
	return func(cfg *ruleConfig) {
		cfg.maxMethods = n
	}
}

func newConfig(opts []Option, severity core.Severity) ruleConfig {
	cfg := ruleConfig{severity: severity}
	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg
}
