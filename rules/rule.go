package rules

import "fmt"

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

func (v Violation) String() string {
	sev := "ERROR"
	if v.Severity == Warning {
		sev = "WARNING"
	}
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

func WithExclude(patterns ...string) Option {
	return func(c *Config) { c.ExcludePatterns = append(c.ExcludePatterns, patterns...) }
}

func (c Config) IsExcluded(path string) bool {
	for _, p := range c.ExcludePatterns {
		if matchPattern(p, path) {
			return true
		}
	}
	return false
}

func matchPattern(pattern, path string) bool {
	if len(pattern) > 3 && pattern[len(pattern)-3:] == "..." {
		prefix := pattern[:len(pattern)-3]
		if len(path) >= len(prefix) && path[:len(prefix)] == prefix {
			return true
		}
		for len(prefix) > 0 && prefix[len(prefix)-1] == '/' {
			prefix = prefix[:len(prefix)-1]
		}
		return path == prefix
	}
	return pattern == path
}
