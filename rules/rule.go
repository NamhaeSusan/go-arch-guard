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
	Sev             Severity
	ExcludePatterns []string
	archModel       *Model
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
	return func(c *Config) { c.Sev = s }
}

func WithModel(m Model) Option {
	return func(c *Config) { c.archModel = &m }
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
