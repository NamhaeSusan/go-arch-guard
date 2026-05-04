package orchestration

import (
	"slices"

	"github.com/NamhaeSusan/go-arch-guard/core"
)

type Option func(*ruleConfig)

type ruleConfig struct {
	severity           core.Severity
	maxBranches        int
	maxStatements      int
	maxCyclomatic      int
	countErrorBranches bool
	ignoredFunctions   []string
	ignoredPaths       []string
	orchestrationDirs  []string
}

func WithSeverity(severity core.Severity) Option {
	return func(cfg *ruleConfig) {
		cfg.severity = severity
	}
}

func WithMaxBranches(n int) Option {
	return func(cfg *ruleConfig) {
		cfg.maxBranches = n
	}
}

func WithMaxStatements(n int) Option {
	return func(cfg *ruleConfig) {
		cfg.maxStatements = n
	}
}

func WithMaxCyclomatic(n int) Option {
	return func(cfg *ruleConfig) {
		cfg.maxCyclomatic = n
	}
}

// WithCountErrorBranches makes simple `if err != nil { return err }`
// branches count toward the branch, statement, and cyclomatic budgets. The
// default discounts these branches so ordinary Go error flow does not drown
// out real orchestration decisions.
func WithCountErrorBranches() Option {
	return func(cfg *ruleConfig) {
		cfg.countErrorBranches = true
	}
}

func WithIgnoredFunctions(names ...string) Option {
	return func(cfg *ruleConfig) {
		cfg.ignoredFunctions = nonBlankStrings(names)
	}
}

// WithIgnoredPaths excludes project-relative files or directories from this
// rule. Patterns ending in "..." match the base path and descendants.
func WithIgnoredPaths(paths ...string) Option {
	return func(cfg *ruleConfig) {
		cfg.ignoredPaths = nonBlankStrings(paths)
	}
}

// WithOrchestrationDirs sets project-relative directories checked by the
// rule. Values may be full paths such as "internal/orchestration" or layout
// names such as "orchestration". Blank names are ignored.
func WithOrchestrationDirs(dirs ...string) Option {
	return func(cfg *ruleConfig) {
		cfg.orchestrationDirs = nonBlankStrings(dirs)
	}
}

func newConfig(opts []Option, severity core.Severity) ruleConfig {
	cfg := ruleConfig{
		severity:      severity,
		maxBranches:   8,
		maxStatements: 40,
		maxCyclomatic: 10,
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	cfg.ignoredFunctions = slices.Clip(cfg.ignoredFunctions)
	cfg.ignoredPaths = slices.Clip(cfg.ignoredPaths)
	cfg.orchestrationDirs = slices.Clip(cfg.orchestrationDirs)
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
