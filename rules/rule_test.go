package rules_test

import (
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/rules"
)

func TestViolation_String(t *testing.T) {
	tests := []struct {
		name string
		v    rules.Violation
		want string
	}{
		{
			name: "error with line",
			v: rules.Violation{
				File:              "internal/domain/user/service.go",
				Line:              10,
				Rule:              "naming.no-stutter",
				Message:           `type "UserService" stutters with package "user"`,
				Fix:               `rename to "Service"`,
				DefaultSeverity:   rules.Error,
				EffectiveSeverity: rules.Error,
			},
			want: `[ERROR] violation: type "UserService" stutters with package "user" (file: internal/domain/user/service.go:10, rule: naming.no-stutter, fix: rename to "Service")`,
		},
		{
			name: "warning without line",
			v: rules.Violation{
				File:              "internal/util/",
				Line:              0,
				Rule:              "structure.banned-package",
				Message:           `package "util" is banned`,
				Fix:               "move to specific domain or pkg/",
				DefaultSeverity:   rules.Warning,
				EffectiveSeverity: rules.Warning,
			},
			want: `[WARNING] violation: package "util" is banned (file: internal/util/, rule: structure.banned-package, fix: move to specific domain or pkg/)`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.v.String(); got != tt.want {
				t.Errorf("got:\n%s\nwant:\n%s", got, tt.want)
			}
		})
	}
}

func TestOptions(t *testing.T) {
	t.Run("default severity is Error", func(t *testing.T) {
		cfg := rules.NewConfig()
		if cfg.Sev != rules.Error {
			t.Errorf("got %v, want Error", cfg.Sev)
		}
	})
	t.Run("WithSeverity sets level", func(t *testing.T) {
		cfg := rules.NewConfig(rules.WithSeverity(rules.Warning))
		if cfg.Sev != rules.Warning {
			t.Errorf("got %v, want Warning", cfg.Sev)
		}
	})
	t.Run("WithExclude sets patterns", func(t *testing.T) {
		cfg := rules.NewConfig(rules.WithExclude("internal/legacy/..."))
		if len(cfg.ExcludePatterns) != 1 || cfg.ExcludePatterns[0] != "internal/legacy/..." {
			t.Errorf("got %v", cfg.ExcludePatterns)
		}
	})
}

func TestIsExcluded(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		path    string
		want    bool
	}{
		{"exact match", "internal/legacy", "internal/legacy", true},
		{"wildcard matches dir", "internal/legacy/...", "internal/legacy", true},
		{"wildcard matches subdir", "internal/legacy/...", "internal/legacy/old", true},
		{"wildcard matches deep subdir", "internal/legacy/...", "internal/legacy/old/deep", true},
		{"wildcard must respect boundary", "internal/domain/foo/...", "internal/domain/foobar", false},
		{"wildcard must respect boundary subpath", "internal/domain/foo/...", "internal/domain/foobar/baz", false},
		{"no match on different path", "internal/legacy/...", "internal/domain/user", false},
		{"exact no partial", "internal/legacy", "internal/legacyv2", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := rules.NewConfig(rules.WithExclude(tt.pattern))
			if got := cfg.IsExcluded(tt.path); got != tt.want {
				t.Errorf("IsExcluded(%q, %q) = %v, want %v", tt.pattern, tt.path, got, tt.want)
			}
		})
	}
}

func TestWithTxBoundary_SetsConfig(t *testing.T) {
	cfg := rules.NewConfig(rules.WithTxBoundary(rules.TxBoundaryConfig{
		StartSymbols:  []string{"database/sql.(*DB).BeginTx"},
		Types:         []string{"database/sql.Tx"},
		AllowedLayers: []string{"app"},
	}))
	if len(cfg.TxBoundary.StartSymbols) != 1 {
		t.Fatalf("expected 1 start symbol, got %d", len(cfg.TxBoundary.StartSymbols))
	}
	if len(cfg.TxBoundary.Types) != 1 {
		t.Fatalf("expected 1 tx type, got %d", len(cfg.TxBoundary.Types))
	}
	if cfg.TxBoundary.AllowedLayers[0] != "app" {
		t.Fatalf("unexpected allowed layer: %s", cfg.TxBoundary.AllowedLayers[0])
	}
}
