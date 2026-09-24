package testpolicy

import "github.com/NamhaeSusan/go-arch-guard/core"

type Option func(*ruleConfig)

type ruleConfig struct {
	severity core.Severity
	strict   bool
	allowed  []string
}

func WithSeverity(severity core.Severity) Option {
	return func(cfg *ruleConfig) {
		cfg.severity = severity
	}
}

func newConfig(opts []Option, severity core.Severity) ruleConfig {
	cfg := ruleConfig{severity: severity}
	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg
}

// WithStrictMocks forbids all handwritten test receiver methods, regardless of
// type name or underlying type. Use mockery-generated interface implementations.
// Pure functions and data-only test types remain allowed.
func WithStrictMocks() Option {
	return func(cfg *ruleConfig) { cfg.strict = true }
}

// WithAllowedTestReceivers permits genuine test helpers, not test doubles.
// Entries are exact module-relative "file.go:Receiver" pairs; no globs.
func WithAllowedTestReceivers(receivers ...string) Option {
	return func(cfg *ruleConfig) { cfg.allowed = append(cfg.allowed, receivers...) }
}
