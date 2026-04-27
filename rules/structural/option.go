package structural

import "github.com/NamhaeSusan/go-arch-guard/core"

type Option func(*ruleConfig)

type ruleConfig struct {
	severity      core.Severity
	middlewareDir string
}

func WithSeverity(severity core.Severity) Option {
	return func(cfg *ruleConfig) {
		cfg.severity = severity
	}
}

// WithMiddlewareDir overrides the directory name MiddlewarePlacement looks
// for and demands to live under <SharedDir>/<name>/. Default: "middleware".
func WithMiddlewareDir(name string) Option {
	return func(cfg *ruleConfig) {
		cfg.middlewareDir = name
	}
}

func newConfig(opts []Option, severity core.Severity) ruleConfig {
	cfg := ruleConfig{severity: severity}
	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg
}
