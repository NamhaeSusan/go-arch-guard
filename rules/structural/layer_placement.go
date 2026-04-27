package structural

import (
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/NamhaeSusan/go-arch-guard/core"
)

const (
	ruleLayerPlacement = "structural.layer-placement"
	misplacedLayer     = "structural.misplaced-layer"
)

// LayerPlacement flags layer-named directories (currently "app", "infra",
// "handler") that appear outside their configured slice. The rule still
// hard-codes these three names; full configurability via LayerDirNames is
// tracked in #93.
type LayerPlacement struct {
	severity core.Severity
}

func NewLayerPlacement(opts ...Option) *LayerPlacement {
	cfg := newConfig(opts, core.Error)
	return &LayerPlacement{severity: cfg.severity}
}

func (r *LayerPlacement) Spec() core.RuleSpec {
	return withSeverity(core.RuleSpec{
		ID:              ruleLayerPlacement,
		Description:     "layer packages must stay in configured structural locations",
		DefaultSeverity: r.severity,
		Violations: []core.ViolationSpec{
			{ID: misplacedLayer, Description: "layer package is outside a valid slice"},
		},
	}, r.severity)
}

func (r *LayerPlacement) Check(ctx *core.Context) []core.Violation {
	if ctx == nil {
		return nil
	}
	arch := ctx.Arch()
	if !hasInternalDir(ctx.Root(), arch.Layout.InternalRoot) {
		return []core.Violation{metaLayoutNotSupported(ruleLayerPlacement)}
	}
	internalDir := filepath.Join(ctx.Root(), arch.Layout.InternalRoot)
	var violations []core.Violation
	_ = filepath.WalkDir(internalDir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || !entry.IsDir() {
			return nil
		}
		rel, relErr := filepath.Rel(filepath.Dir(internalDir), path)
		if relErr != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		if rel == arch.Layout.InternalRoot {
			return nil
		}
		if ctx.IsExcluded(rel + "/") {
			return filepath.SkipDir
		}
		if !hasNonTestGoFiles(path) {
			return nil
		}
		name := entry.Name()
		if !isMisplacedLayerDir(arch, rel, name) {
			return nil
		}
		violations = append(violations, violation(r.severity, misplacedLayer, rel+"/",
			`layer package "`+name+`" is misplaced`,
			"place layer packages only in configured domain slices or "+arch.Layout.OrchestrationDir+" handler"))
		return nil
	})
	return violations
}

func isMisplacedLayerDir(arch core.Architecture, rel, name string) bool {
	if len(arch.Layers.LayerDirNames) > 0 && !arch.Layers.LayerDirNames[name] {
		return false
	}
	switch name {
	case "app", "infra":
		if name == "app" && arch.Layout.AppDir != "" && rel == filepath.ToSlash(filepath.Join(arch.Layout.InternalRoot, arch.Layout.AppDir)) {
			return false
		}
		return !matchesDomainLayer(arch, rel, name)
	case "handler":
		if arch.Layout.ServerDir != "" {
			serverBase := filepath.ToSlash(filepath.Join(arch.Layout.InternalRoot, arch.Layout.ServerDir))
			if strings.HasPrefix(rel, serverBase+"/") || rel == serverBase {
				return false
			}
		}
		return !matchesDomainLayer(arch, rel, name) &&
			rel != filepath.ToSlash(filepath.Join(arch.Layout.InternalRoot, arch.Layout.OrchestrationDir, "handler"))
	default:
		return false
	}
}

func matchesDomainLayer(arch core.Architecture, rel, name string) bool {
	parts := strings.Split(rel, "/")
	return len(parts) == 4 && parts[0] == arch.Layout.InternalRoot && parts[1] == arch.Layout.DomainDir && parts[2] != "" && parts[3] == name
}

var _ core.Rule = (*LayerPlacement)(nil)
