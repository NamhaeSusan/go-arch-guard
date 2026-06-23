package orchestration

import "github.com/NamhaeSusan/go-arch-guard/core"

type Option func(*ruleConfig)

type ruleConfig struct {
	severity                       core.Severity
	allowConstructorServiceAliases bool
}

func WithSeverity(severity core.Severity) Option {
	return func(cfg *ruleConfig) {
		cfg.severity = severity
	}
}

func WithConstructorServiceAliases(allow bool) Option {
	return func(cfg *ruleConfig) {
		cfg.allowConstructorServiceAliases = allow
	}
}

func newConfig(opts []Option, severity core.Severity) ruleConfig {
	cfg := ruleConfig{
		severity:                       severity,
		allowConstructorServiceAliases: true,
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg
}
