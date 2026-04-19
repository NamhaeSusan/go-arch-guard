package rules

import "testing"

func TestClassifyInternalPackage(t *testing.T) {
	m := DDD()
	internalPrefix := "example.com/myapp/internal/"

	tests := []struct {
		name     string
		pkgPath  string
		wantKind internalKind
		wantDom  string
		wantSub  string
		wantAls  bool
	}{
		{
			name:     "domain sublayer",
			pkgPath:  "example.com/myapp/internal/domain/order/app",
			wantKind: kindDomain,
			wantDom:  "order",
			wantSub:  "app",
		},
		{
			name:     "domain root (alias package)",
			pkgPath:  "example.com/myapp/internal/domain/order",
			wantKind: kindDomainRoot,
			wantDom:  "order",
			wantSub:  "",
			wantAls:  true,
		},
		{
			name:     "orchestration package",
			pkgPath:  "example.com/myapp/internal/orchestration",
			wantKind: kindOrchestration,
		},
		{
			name:     "orchestration sub-package",
			pkgPath:  "example.com/myapp/internal/orchestration/handler",
			wantKind: kindOrchestration,
		},
		{
			name:     "shared/pkg package",
			pkgPath:  "example.com/myapp/internal/pkg",
			wantKind: kindShared,
		},
		{
			name:     "shared sub-package",
			pkgPath:  "example.com/myapp/internal/pkg/errors",
			wantKind: kindShared,
		},
		{
			name:     "unclassified package",
			pkgPath:  "example.com/myapp/internal/unknown",
			wantKind: kindUnclassified,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := classifyInternalPackage(m, tc.pkgPath, internalPrefix)
			if got.Kind != tc.wantKind {
				t.Errorf("Kind = %v, want %v", got.Kind, tc.wantKind)
			}
			if got.Domain != tc.wantDom {
				t.Errorf("Domain = %q, want %q", got.Domain, tc.wantDom)
			}
			if got.Sublayer != tc.wantSub {
				t.Errorf("Sublayer = %q, want %q", got.Sublayer, tc.wantSub)
			}
			if got.IsAlias != tc.wantAls {
				t.Errorf("IsAlias = %v, want %v", got.IsAlias, tc.wantAls)
			}
		})
	}
}

func TestClassifyInternalPackage_FlatLayout(t *testing.T) {
	m := ConsumerWorker()
	internalPrefix := "example.com/myapp/internal/"

	tests := []struct {
		name     string
		pkgPath  string
		wantKind internalKind
		wantSub  string
	}{
		{
			name:     "worker layer",
			pkgPath:  "example.com/myapp/internal/worker",
			wantKind: kindDomain,
			wantSub:  "worker",
		},
		{
			name:     "service layer",
			pkgPath:  "example.com/myapp/internal/service",
			wantKind: kindDomain,
			wantSub:  "service",
		},
		{
			name:     "shared/pkg",
			pkgPath:  "example.com/myapp/internal/pkg",
			wantKind: kindShared,
		},
		{
			name:     "unknown layer",
			pkgPath:  "example.com/myapp/internal/mystery",
			wantKind: kindUnclassified,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := classifyInternalPackage(m, tc.pkgPath, internalPrefix)
			if got.Kind != tc.wantKind {
				t.Errorf("Kind = %v, want %v", got.Kind, tc.wantKind)
			}
			if got.Sublayer != tc.wantSub {
				t.Errorf("Sublayer = %q, want %q", got.Sublayer, tc.wantSub)
			}
		})
	}
}

func TestClassifyInternalPackage_EventPipeline(t *testing.T) {
	m := EventPipeline()
	internalPrefix := "example.com/myapp/internal/"

	tests := []struct {
		name     string
		pkgPath  string
		wantKind internalKind
		wantSub  string
	}{
		{
			name:     "command layer",
			pkgPath:  "example.com/myapp/internal/command",
			wantKind: kindDomain,
			wantSub:  "command",
		},
		{
			name:     "aggregate layer",
			pkgPath:  "example.com/myapp/internal/aggregate",
			wantKind: kindDomain,
			wantSub:  "aggregate",
		},
		{
			name:     "shared/pkg",
			pkgPath:  "example.com/myapp/internal/pkg",
			wantKind: kindShared,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := classifyInternalPackage(m, tc.pkgPath, internalPrefix)
			if got.Kind != tc.wantKind {
				t.Errorf("Kind = %v, want %v", got.Kind, tc.wantKind)
			}
			if got.Sublayer != tc.wantSub {
				t.Errorf("Sublayer = %q, want %q", got.Sublayer, tc.wantSub)
			}
		})
	}
}
