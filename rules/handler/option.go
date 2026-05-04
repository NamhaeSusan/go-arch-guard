package handler

import (
	"strings"

	"github.com/NamhaeSusan/go-arch-guard/core"
)

type Option func(*ruleConfig)

type ruleConfig struct {
	severity          core.Severity
	allowedModelTypes map[string]bool
}

func WithSeverity(severity core.Severity) Option {
	return func(cfg *ruleConfig) {
		cfg.severity = severity
	}
}

// WithAllowedModelTypes allows selected domain model types to remain visible
// on handler/transport response boundaries. Values must be fully-qualified
// type names such as "example.com/shop/internal/domain/order/core/model.Order".
func WithAllowedModelTypes(types ...string) Option {
	return func(cfg *ruleConfig) {
		if cfg.allowedModelTypes == nil {
			cfg.allowedModelTypes = make(map[string]bool, len(types))
		}
		for _, typ := range types {
			typ = strings.TrimSpace(typ)
			if typ != "" {
				cfg.allowedModelTypes[typ] = true
			}
		}
	}
}

func newConfig(opts []Option, severity core.Severity) ruleConfig {
	cfg := ruleConfig{
		severity:          severity,
		allowedModelTypes: make(map[string]bool),
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg
}
