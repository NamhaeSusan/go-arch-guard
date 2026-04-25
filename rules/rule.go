package rules

import (
	"strings"

	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/core/analysisutil"
)

type Severity = core.Severity

const (
	Error   = core.Error
	Warning = core.Warning
)

type Violation = core.Violation

type Option func(*Config)

type Config struct {
	Sev                     Severity
	sevExplicit             bool // true if WithSeverity was called explicitly
	ExcludePatterns         []string
	archModel               *Model
	MaxRepoInterfaceMethods int
	TxBoundary              TxBoundaryConfig
}

// TxBoundaryConfig configures CheckTxBoundary. Empty config → rule is a no-op.
type TxBoundaryConfig struct {
	// Fully-qualified symbols that start a transaction.
	// Examples: "database/sql.(*DB).BeginTx", "database/sql.(*DB).Begin"
	StartSymbols []string
	// Fully-qualified transaction type names.
	// Examples: "database/sql.Tx", "github.com/jackc/pgx/v5.Tx"
	Types []string
	// Layers allowed to both start tx and name tx types in signatures.
	// Uses the same notation as Model.Sublayers ("app", "core/repo", flat names).
	// Defaults to []string{"app"} when empty.
	AllowedLayers []string
}

func (c Config) model() Model {
	if c.archModel != nil {
		return *c.archModel
	}
	return defaultModel
}

func NewConfig(opts ...Option) Config {
	c := Config{Sev: Error}
	for _, o := range opts {
		o(&c)
	}
	return c
}

func WithSeverity(s Severity) Option {
	return func(c *Config) {
		c.Sev = s
		c.sevExplicit = true
	}
}

// SeverityExplicit reports whether WithSeverity was called explicitly by the
// caller. Rules that have a non-default severity (e.g. warnings) use this to
// decide whether to honor the caller's override or use their own default.
func (c Config) SeverityExplicit() bool {
	return c.sevExplicit
}

func WithModel(m Model) Option {
	return func(c *Config) { c.archModel = &m }
}

func WithMaxRepoInterfaceMethods(n int) Option {
	return func(c *Config) { c.MaxRepoInterfaceMethods = n }
}

func WithTxBoundary(cfg TxBoundaryConfig) Option {
	return func(c *Config) { c.TxBoundary = cfg }
}

func WithExclude(patterns ...string) Option {
	return func(c *Config) {
		for _, pattern := range patterns {
			c.ExcludePatterns = append(c.ExcludePatterns, normalizeMatchPath(pattern))
		}
	}
}

func (c Config) IsExcluded(path string) bool {
	path = normalizeMatchPath(path)
	for _, p := range c.ExcludePatterns {
		if matchPattern(p, path) {
			return true
		}
	}
	return false
}

func matchPattern(pattern, path string) bool {
	pattern = normalizeMatchPath(pattern)
	path = normalizeMatchPath(path)
	if len(pattern) > 3 && pattern[len(pattern)-3:] == "..." {
		prefix := strings.TrimRight(pattern[:len(pattern)-3], "/")
		return path == prefix || strings.HasPrefix(path, prefix+"/")
	}
	return pattern == path
}

func normalizeMatchPath(path string) string {
	return analysisutil.NormalizeMatchPath(path)
}
