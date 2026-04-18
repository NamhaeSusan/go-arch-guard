package rules

import (
	"fmt"
	"path/filepath"
	"strings"
)

type Severity int

const (
	Error Severity = iota
	Warning
)

type Violation struct {
	File     string
	Line     int
	Rule     string
	Message  string
	Fix      string
	Severity Severity
}

func (s Severity) String() string {
	if s == Warning {
		return "WARNING"
	}
	return "ERROR"
}

func (v Violation) String() string {
	sev := v.Severity.String()
	fileStr := v.File
	if v.Line > 0 {
		fileStr = fmt.Sprintf("%s:%d", v.File, v.Line)
	}
	return fmt.Sprintf("[%s] violation: %s (file: %s, rule: %s, fix: %s)",
		sev, v.Message, fileStr, v.Rule, v.Fix)
}

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
	// The special layer name "cmd" covers every package under <module>/cmd/...
	// Defaults to []string{"app"} when empty.
	AllowedLayers []string
	// EnforceUnclassified controls how internal packages that do not map to any
	// known sublayer (e.g. internal/testutil, internal/generic, codegen output)
	// are treated.
	//   false (default): skip — preserves old behavior, avoids noise on ad-hoc
	//                    helper packages.
	//   true:            treat as non-allowed — a forbidden call there produces
	//                    a violation. Use when the team wants strict coverage
	//                    and is willing to add explicit WithExclude for
	//                    legitimate helper packages.
	//
	// Packages under <module>/cmd/... are always scanned regardless of this
	// flag because cmd/ is a well-defined composition-root layer.
	EnforceUnclassified bool
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
	path = filepath.ToSlash(path)
	for strings.HasPrefix(path, "./") {
		path = strings.TrimPrefix(path, "./")
	}
	return path
}
