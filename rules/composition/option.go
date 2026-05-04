package composition

import (
	"strings"

	"github.com/NamhaeSusan/go-arch-guard/core"
)

type Option func(*ruleConfig)

type ruleConfig struct {
	severity         core.Severity
	compositionRoots []string
	includeTestFiles bool
}

func WithSeverity(severity core.Severity) Option {
	return func(cfg *ruleConfig) {
		cfg.severity = severity
	}
}

// WithCompositionRoots adds project-relative package roots that may import
// domain infra adapters directly. Patterns may end in "/..." to include
// descendants. Defaults include cmd/... and internal/<AppDir>/....
func WithCompositionRoots(roots ...string) Option {
	return func(cfg *ruleConfig) {
		for _, root := range roots {
			root = strings.TrimSpace(root)
			if root != "" {
				cfg.compositionRoots = append(cfg.compositionRoots, root)
			}
		}
	}
}

// WithTestFiles controls whether imports from _test.go files are checked.
// The default is false so tests can construct adapters directly.
func WithTestFiles(include bool) Option {
	return func(cfg *ruleConfig) {
		cfg.includeTestFiles = include
	}
}

func newConfig(opts []Option, severity core.Severity) ruleConfig {
	cfg := ruleConfig{severity: severity}
	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg
}
