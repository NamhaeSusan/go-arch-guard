package core

import (
	"path/filepath"
	"strings"

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
// stripped, OS path separators converted) at construction.
func NewContext(pkgs []*packages.Package, module, root string, arch Architecture, exclude []string) *Context {
	norm := make([]string, len(exclude))
	for i, p := range exclude {
		norm[i] = normalizeMatchPath(p)
	}
	return &Context{
		pkgs:    pkgs,
		module:  module,
		root:    root,
		arch:    arch,
		exclude: norm,
	}
}

// Pkgs returns the loaded packages.
//
// IMPORTANT: This is a slice-header copy only. Reslicing or appending to
// the returned slice cannot affect other rules. However, the
// *packages.Package values it points at are SHARED across rules — Go does
// not let us deep-clone them cheaply, and a true copy would re-walk the
// type system per rule. Mutating any field of a *packages.Package
// (Imports, Types, Syntax, Errors, …) is therefore a contract violation:
// rules MUST be pure functions of their input. Violating this corrupts
// later rules' view of the world and is undefined behavior under any
// future parallel runner.
func (c *Context) Pkgs() []*packages.Package {
	if c.pkgs == nil {
		return nil
	}
	out := make([]*packages.Package, len(c.pkgs))
	copy(out, c.pkgs)
	return out
}
func (c *Context) Module() string            { return c.module }
func (c *Context) Root() string              { return c.root }
func (c *Context) Arch() Architecture        { return c.arch }

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

func normalizeMatchPath(path string) string {
	path = filepath.ToSlash(path)
	for strings.HasPrefix(path, "./") {
		path = strings.TrimPrefix(path, "./")
	}
	return path
}
