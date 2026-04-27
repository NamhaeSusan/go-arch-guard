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

// WithMaxMethods sets the per-interface method count cap. It applies only
// to interfaces.NewTooManyMethods — passing it to other interfaces.New*()
// rules (NewPattern, NewContainer, NewCrossDomainAnonymous) is silently
// ignored to keep the option API uniform across the package.
//
// Values <= 0 are treated as "use the default" (currently 10).
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
