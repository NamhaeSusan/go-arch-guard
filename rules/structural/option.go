package structural

import "github.com/NamhaeSusan/go-arch-guard/core"

type Option func(*ruleConfig)

type ruleConfig struct {
	severity            core.Severity
	middlewareDir       string
	dtoFilenames        []string
	dtoFilenameSuffixes []string
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

// WithDTOFilenames overrides the exact filenames DTOPlacement flags. Default:
// "dto.go".
func WithDTOFilenames(names ...string) Option {
	return func(cfg *ruleConfig) {
		cfg.dtoFilenames = append([]string(nil), names...)
	}
}

// WithDTOFilenameSuffixes overrides the filename suffixes (other than ".go")
// DTOPlacement flags. Default: "_dto.go".
func WithDTOFilenameSuffixes(suffixes ...string) Option {
	return func(cfg *ruleConfig) {
		cfg.dtoFilenameSuffixes = append([]string(nil), suffixes...)
	}
}

func newConfig(opts []Option, severity core.Severity) ruleConfig {
	cfg := ruleConfig{severity: severity}
	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg
}
