package core

import "testing"

func TestNewContextStoresInputs(t *testing.T) {
	arch := validArchitecture()
	c := NewContext(nil, "github.com/example/foo", "/repo", arch, []string{"vendor/..."})

	if c.Module() != "github.com/example/foo" {
		t.Errorf("Module() = %q", c.Module())
	}
	if c.Root() != "/repo" {
		t.Errorf("Root() = %q", c.Root())
	}
	if got := c.Arch().Layers.Sublayers[0]; got != "handler" {
		t.Errorf("Arch().Layers.Sublayers[0] = %q", got)
	}
	if c.Pkgs() != nil {
		t.Errorf("Pkgs() = %v, want nil", c.Pkgs())
	}
}

func TestContextIsExcludedExactMatch(t *testing.T) {
	c := NewContext(nil, "", "", Architecture{}, []string{"internal/handler/foo.go"})
	if !c.IsExcluded("internal/handler/foo.go") {
		t.Errorf("IsExcluded(exact) = false")
	}
	if c.IsExcluded("internal/handler/bar.go") {
		t.Errorf("IsExcluded(non-match) = true")
	}
}

func TestContextIsExcludedRecursive(t *testing.T) {
	c := NewContext(nil, "", "", Architecture{}, []string{"internal/handler/..."})
	if !c.IsExcluded("internal/handler") {
		t.Errorf("IsExcluded(prefix base) = false")
	}
	if !c.IsExcluded("internal/handler/foo.go") {
		t.Errorf("IsExcluded(nested) = false")
	}
	if c.IsExcluded("internal/app/foo.go") {
		t.Errorf("IsExcluded(sibling) = true")
	}
}

func TestContextIsExcludedNormalizesLeadingDotSlash(t *testing.T) {
	c := NewContext(nil, "", "", Architecture{}, []string{"./internal/foo.go"})
	if !c.IsExcluded("internal/foo.go") {
		t.Errorf("IsExcluded should normalize leading ./")
	}
}

func TestContextIsExcludedNormalizesBackslashes(t *testing.T) {
	// A pattern entered with backslashes (e.g. copied from a Windows shell)
	// must match forward-slash paths emitted by rules on every platform.
	c := NewContext(nil, "", "", Architecture{}, []string{`internal\handler\...`})
	if !c.IsExcluded("internal/handler/foo.go") {
		t.Errorf("backslash exclude pattern must match forward-slash path")
	}
}
