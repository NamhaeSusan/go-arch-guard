package structural_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/rules/structural"
)

func TestMiddlewarePlacement(t *testing.T) {
	t.Run("valid fixture has no middleware-placement violations", func(t *testing.T) {
		violations := runRule(t, "../../testdata/valid", structural.NewMiddlewarePlacement())
		assertNoRulePrefix(t, violations, "structural.middleware-placement")
	})

	t.Run("detects invalid fixture middleware in wrong location", func(t *testing.T) {
		violations := runRule(t, "../../testdata/invalid", structural.NewMiddlewarePlacement())
		assertViolation(t, violations, "structural.middleware-placement", "internal/handler/middleware/")
	})

	t.Run("WithMiddlewareDir overrides the directory name", func(t *testing.T) {
		root := t.TempDir()
		// "interceptor/" outside <SharedDir>/ should be flagged when WithMiddlewareDir("interceptor") is set,
		// while the same project's "middleware/" directory must be ignored.
		writeTestFile(t, filepath.Join(root, "internal", "handler", "interceptor", "auth.go"), "package interceptor\n")
		writeTestFile(t, filepath.Join(root, "internal", "domain", "order", "middleware", "logging.go"), "package middleware\n")

		ctx := core.NewContext(nil, "github.com/example/app", root, dddArch(), nil)
		got := core.Run(ctx, core.NewRuleSet(structural.NewMiddlewarePlacement(structural.WithMiddlewareDir("interceptor"))))

		var sawInterceptor, sawMiddleware bool
		for _, v := range got {
			if v.Rule != "structural.middleware-placement" {
				continue
			}
			if strings.Contains(v.File, "interceptor") {
				sawInterceptor = true
			}
			if strings.Contains(v.File, "middleware") {
				sawMiddleware = true
			}
		}
		if !sawInterceptor {
			t.Fatalf("expected interceptor/ to be flagged when WithMiddlewareDir is interceptor; got %+v", got)
		}
		if sawMiddleware {
			t.Fatalf("middleware/ should not be flagged when MiddlewareDir is interceptor; got %+v", got)
		}
	})
}
