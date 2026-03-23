package analyzer_test

import (
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/analyzer"
)

func TestLoad(t *testing.T) {
	t.Run("returns error for nonexistent directory", func(t *testing.T) {
		_, err := analyzer.Load("/nonexistent", "internal/...")
		if err == nil {
			t.Error("expected error for nonexistent directory")
		}
	})
}
