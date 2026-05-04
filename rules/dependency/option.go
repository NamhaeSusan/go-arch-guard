package dependency

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
// domain infra adapters directly when using dependency.NewRootOnlyInfraUse.
// Patterns may end in "/..." to include descendants. Defaults include
// cmd/... and internal/<AppDir>/....
//
// Other dependency.New*() rules silently ignore this option to keep the
// option API uniform across the package.
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

// WithTestFiles controls whether dependency.NewRootOnlyInfraUse checks imports
// from _test.go files when the context includes packages loaded with tests.
// Use analyzer.LoadWithTests to include test files. The default is false so
// tests can construct adapters directly.
//
// Other dependency.New*() rules silently ignore this option to keep the
// option API uniform across the package.
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
