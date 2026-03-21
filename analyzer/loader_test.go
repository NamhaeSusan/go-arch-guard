package analyzer_test

import (
	"testing"

	"github.com/kimtaeyun/go-arch-guard/analyzer"
)

func TestLoad(t *testing.T) {
	t.Run("loads valid project packages", func(t *testing.T) {
		pkgs, err := analyzer.Load("../testdata/valid", "internal/...")
		if err != nil {
			t.Fatal(err)
		}
		if len(pkgs) == 0 {
			t.Fatal("expected at least one package")
		}
		found := false
		for _, pkg := range pkgs {
			if pkg.Name == "user" {
				found = true
			}
		}
		if !found {
			t.Error("expected to find package 'user'")
		}
	})

	t.Run("returns error for nonexistent directory", func(t *testing.T) {
		_, err := analyzer.Load("/nonexistent", "internal/...")
		if err == nil {
			t.Error("expected error for nonexistent directory")
		}
	})
}
