package tui

import (
	"github.com/NamhaeSusan/go-arch-guard/rules"
	"golang.org/x/tools/go/packages"
)

// Severity level for a path in the tree.
type severity int

const (
	sevNone    severity = iota
	sevWarning          // only warnings
	sevError            // at least one error
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

// Severity returns the worst severity for a path and all sub-paths.
func (vi ViolationIndex) Severity(relPath string) severity {
	worst := sevNone
	vi.walkPath(relPath, func(viols []rules.Violation) {
		for _, v := range viols {
			if v.Severity == rules.Error {
				worst = sevError
				return
			}
			if v.Severity == rules.Warning && worst < sevWarning {
				worst = sevWarning
			}
		}
	})
	return worst
}

func (vi ViolationIndex) walkPath(relPath string, fn func([]rules.Violation)) {
	if viols, ok := vi[relPath]; ok {
		fn(viols)
	}
	prefix := relPath + "/"
	for k, viols := range vi {
		if len(k) > len(prefix) && k[:len(prefix)] == prefix {
			fn(viols)
		}
	}
}
