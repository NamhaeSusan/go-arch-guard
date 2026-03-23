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

	t.Run("loads packages without requiring successful type checking", func(t *testing.T) {
		pkgs, err := analyzer.Load("../testdata/load_type_error", "internal/...")
		if err != nil {
			t.Fatalf("expected type-check errors to be ignored, got %v", err)
		}
		if len(pkgs) == 0 {
			t.Fatal("expected packages to be loaded")
		}
	})
}
