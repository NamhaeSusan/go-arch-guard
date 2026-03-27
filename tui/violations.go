package tui

import (
	"github.com/NamhaeSusan/go-arch-guard/rules"
	"golang.org/x/tools/go/packages"
)

// ViolationIndex maps relative package paths to their violations.
type ViolationIndex map[string][]rules.Violation

// BuildViolationIndex runs all rules and indexes violations by package path.
func BuildViolationIndex(pkgs []*packages.Package, module, root string) ViolationIndex {
	var all []rules.Violation
	all = append(all, rules.CheckDomainIsolation(pkgs, module, root)...)
	all = append(all, rules.CheckLayerDirection(pkgs, module, root)...)
	all = append(all, rules.CheckNaming(pkgs)...)
	all = append(all, rules.CheckStructure(root)...)
	all = append(all, rules.AnalyzeBlastRadius(pkgs, module, root)...)

	idx := make(ViolationIndex)
	for _, v := range all {
		idx[v.File] = append(idx[v.File], v)
	}
	return idx
}

// HasViolations returns true if the path or any sub-path has violations.
func (vi ViolationIndex) HasViolations(relPath string) bool {
	if _, ok := vi[relPath]; ok {
		return true
	}
	prefix := relPath + "/"
	for k := range vi {
		if len(k) > len(prefix) && k[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}
