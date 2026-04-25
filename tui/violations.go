package tui

import (
	"strings"

	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/presets"
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
type ViolationIndex map[string][]core.Violation

// BuildViolationIndex runs all rules and indexes violations by package path.
func BuildViolationIndex(pkgs []*packages.Package, module, root string) ViolationIndex {
	arch := presets.DDD()
	ctx := core.NewContext(pkgs, module, root, arch, nil)
	ruleSet := presets.RecommendedDDD()
	all := core.Run(ctx, ruleSet)

	idx := make(ViolationIndex)
	for _, v := range all {
		key := strings.TrimRight(v.File, "/")
		idx[key] = append(idx[key], v)
	}
	return idx
}

// Severity returns the worst severity for a path and all sub-paths.
func (vi ViolationIndex) Severity(relPath string) severity {
	worst := sevNone
	vi.walkPath(relPath, func(viols []core.Violation) {
		for _, v := range viols {
			if v.EffectiveSeverity == core.Error {
				worst = sevError
				return
			}
			if v.EffectiveSeverity == core.Warning && worst < sevWarning {
				worst = sevWarning
			}
		}
	})
	return worst
}

func (vi ViolationIndex) walkPath(relPath string, fn func([]core.Violation)) {
	if viols, ok := vi[relPath]; ok {
		fn(viols)
	}
	prefix := relPath + "/"
	for k, viols := range vi {
		if strings.HasPrefix(k, prefix) {
			fn(viols)
		}
	}
}
