package core

import (
	"strings"

	"github.com/NamhaeSusan/go-arch-guard/internal/pathmatch"
	"golang.org/x/tools/go/packages"
)

// Context is the read-only view a Rule.Check sees. Fields are unexported
// to enforce immutability — rules access them through accessor methods.
// Run constructs Context once per invocation and shares it across rules.
type Context struct {
	pkgs    []*packages.Package
	module  string
	root    string
	arch    Architecture
	exclude []string
}

// NewContext builds a Context. Excludes are normalized (leading "./"
// stripped, OS path separators converted) at construction. The Architecture
// is deep-cloned so caller-side mutations to its maps and slices after this
// call cannot leak into the Context's view; Arch() also returns a defensive
// clone, but cloning here closes the construction-time window where a
// caller could otherwise mutate shared state before the first Arch() read.
func NewContext(pkgs []*packages.Package, module, root string, arch Architecture, exclude []string) *Context {
	module, root = resolveContextDefaults(pkgs, module, root)

	norm := make([]string, len(exclude))
	for i, p := range exclude {
		norm[i] = normalizeMatchPath(p)
	}
	return &Context{
		pkgs:    pkgs,
		module:  module,
		root:    root,
		arch:    cloneArchitecture(arch),
		exclude: norm,
	}
}

func resolveContextDefaults(pkgs []*packages.Package, module, root string) (string, string) {
	if module != "" && root != "" {
		return module, root
	}
	for _, pkg := range pkgs {
		if pkg == nil || pkg.Module == nil {
			continue
		}
		if module == "" && pkg.Module.Path != "" {
			module = pkg.Module.Path
		}
		if root == "" && pkg.Module.Dir != "" {
			root = pkg.Module.Dir
		}
		if module != "" && root != "" {
			return module, root
		}
	}
	return module, root
}

// Pkgs returns the loaded packages. Treat the returned *packages.Package
// values as read-only — mutation is undefined behavior.
//
// The returned slice is a header copy: reslicing or appending to it
// cannot affect other rules. However, the *packages.Package values it
// points at are SHARED across rules — Go does not let us deep-clone them
// cheaply, and a true copy would re-walk the type system per rule.
// Mutating any field of a *packages.Package (Imports, Types, Syntax,
// Errors, …) is therefore a contract violation: rules MUST be pure
// functions of their input. Violating this corrupts later rules' view of
// the world and is undefined behavior under any future parallel runner.
func (c *Context) Pkgs() []*packages.Package {
	if c.pkgs == nil {
		return nil
	}
	out := make([]*packages.Package, len(c.pkgs))
	copy(out, c.pkgs)
	return out
}
func (c *Context) Module() string { return c.module }
func (c *Context) Root() string   { return c.root }

// Arch returns a defensive deep copy of the architecture so a rule that
// mutates returned slices/maps cannot corrupt later rules' view of the
// policy. The cost is small (Architecture is bounded in size) and it
// preserves the "single source of truth" guarantee on Layers.Sublayers
// when several rules read from the context concurrently in some future
// runner.
func (c *Context) Arch() Architecture {
	return cloneArchitecture(c.arch)
}

// IsExcluded reports whether path matches any configured exclude pattern.
// Patterns ending in "..." match the base directory and any descendant;
// other patterns require an exact match. Both pattern and path are
// normalized to forward slashes with leading "./" stripped.
func (c *Context) IsExcluded(path string) bool {
	path = normalizeMatchPath(path)
	for _, p := range c.exclude {
		if matchExcludePattern(p, path) {
			return true
		}
	}
	return false
}

func matchExcludePattern(pattern, path string) bool {
	if len(pattern) > 3 && pattern[len(pattern)-3:] == "..." {
		prefix := strings.TrimRight(pattern[:len(pattern)-3], "/")
		return path == prefix || strings.HasPrefix(path, prefix+"/")
	}
	return pattern == path
}

// normalizeMatchPath canonicalizes both exclude patterns and the file paths
// rules emit so they match consistently. The supported equivalences are:
//
//   - Backslashes are converted to forward slashes (Windows-shell paste).
//   - Leading "/" is trimmed (absolute-path fallback in analysisutil).
//   - Leading "./" is trimmed (rule-emitted relative paths).
//   - Trailing "/" is trimmed for non-recursive patterns; a recursive
//     pattern keeps its "..." suffix intact since matchExcludePattern
//     reads it.
//
// As a result "internal/foo", "/internal/foo", "./internal/foo",
// "internal\\foo", and "internal/foo/" are all the same key.
func normalizeMatchPath(path string) string {
	return pathmatch.Normalize(path)
}
