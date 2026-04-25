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

func TestContextIsExcludedNormalizesLeadingSlash(t *testing.T) {
	c := NewContext(nil, "", "", Architecture{}, []string{"internal/handler/foo.go"})
	if !c.IsExcluded("/internal/handler/foo.go") {
		t.Errorf("IsExcluded should normalize leading / on the input path")
	}
}

func TestContextIsExcludedNormalizesTrailingSlash(t *testing.T) {
	// Pattern without trailing slash should match input with trailing slash
	// (placement rules emit dir paths as "internal/foo/").
	c := NewContext(nil, "", "", Architecture{}, []string{"internal/handler"})
	if !c.IsExcluded("internal/handler/") {
		t.Errorf("IsExcluded should match trailing-slash input against bare pattern")
	}

	// And the reverse: trailing-slash pattern matching bare input.
	c2 := NewContext(nil, "", "", Architecture{}, []string{"internal/handler/"})
	if !c2.IsExcluded("internal/handler") {
		t.Errorf("IsExcluded should match bare input against trailing-slash pattern")
	}
}

func TestContextIsExcludedRecursivePatternUnchanged(t *testing.T) {
	// Regression guard: trailing-slash trim must not break recursive "..." semantics.
	c := NewContext(nil, "", "", Architecture{}, []string{"internal/handler/..."})
	if !c.IsExcluded("internal/handler/foo.go") {
		t.Errorf("recursive pattern broke after normalization changes")
	}
	if !c.IsExcluded("internal/handler") {
		t.Errorf("recursive pattern should still match base path")
	}
	if c.IsExcluded("internal/app/foo.go") {
		t.Errorf("recursive pattern should not match siblings")
	}
}

func TestNewContextIsolatesCallerArchitectureMutation(t *testing.T) {
	arch := validArchitecture()
	arch.Layers.LayerDirNames = map[string]bool{"handler": true, "app": true}
	c := NewContext(nil, "m", "/r", arch, nil)

	// Mutate the original arch maps and slices after NewContext.
	arch.Layers.Sublayers = append(arch.Layers.Sublayers, "rogue")
	arch.Layers.LayerDirNames["rogue"] = true
	arch.Layers.Direction["rogue"] = []string{"handler"}

	got := c.Arch()
	for _, sl := range got.Layers.Sublayers {
		if sl == "rogue" {
			t.Fatalf("caller mutation leaked into Context: Sublayers contains rogue")
		}
	}
	if got.Layers.LayerDirNames["rogue"] {
		t.Fatalf("caller mutation leaked into Context: LayerDirNames contains rogue")
	}
	if _, ok := got.Layers.Direction["rogue"]; ok {
		t.Fatalf("caller mutation leaked into Context: Direction contains rogue")
	}
}
