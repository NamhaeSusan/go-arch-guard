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

// LayerPlacement flags layer-named directories that appear outside their
// allowed slice. LayerModel.LayerLocations can define custom placement
// templates for any layer basename. When no explicit template exists for a
// name, the rule keeps its legacy fallback vocabulary:
//
//   - "app": allowed at internal/<Layout.AppDir> or per-domain
//     internal/<DomainDir>/<X>/app/
//   - "infra": allowed only as per-domain
//     internal/<DomainDir>/<X>/infra/. Global internal/infra/ is rejected
//     by design (the library opts the team into per-domain infra).
//   - "handler": allowed under internal/<ServerDir>/, in
//     internal/<OrchestrationDir>/handler/, or as per-domain
//     internal/<DomainDir>/<X>/handler/.
//
// LayerDirNames can narrow the active set (only names present in the map
// are checked). LayerLocations adds protection for additional names such as
// "controller", "usecase", or "port" without changing the legacy fallback.
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
	if locations, ok := arch.Layers.LayerLocations[name]; ok {
		return !matchesAnyLayerLocation(arch, rel, locations)
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

func matchesAnyLayerLocation(arch core.Architecture, rel string, locations []string) bool {
	for _, location := range locations {
		if matchLayerLocation(arch, rel, location) {
			return true
		}
	}
	return false
}

func matchLayerLocation(arch core.Architecture, rel, location string) bool {
	location = strings.Trim(strings.TrimSpace(location), "/")
	location = strings.NewReplacer(
		"{InternalRoot}", arch.Layout.InternalRoot,
		"{DomainDir}", arch.Layout.DomainDir,
		"{OrchestrationDir}", arch.Layout.OrchestrationDir,
		"{SharedDir}", arch.Layout.SharedDir,
		"{AppDir}", arch.Layout.AppDir,
		"{ServerDir}", arch.Layout.ServerDir,
	).Replace(location)
	patternParts := strings.Split(filepath.ToSlash(location), "/")
	relParts := strings.Split(strings.Trim(filepath.ToSlash(rel), "/"), "/")
	return matchLayerLocationParts(patternParts, relParts)
}

func matchLayerLocationParts(pattern, rel []string) bool {
	if len(pattern) == 0 {
		return len(rel) == 0
	}
	if pattern[0] == "**" {
		return len(pattern) == 1 || matchLayerLocationParts(pattern[1:], rel) ||
			(len(rel) > 0 && matchLayerLocationParts(pattern, rel[1:]))
	}
	if len(rel) == 0 {
		return false
	}
	if pattern[0] == "*" {
		return rel[0] != "" && matchLayerLocationParts(pattern[1:], rel[1:])
	}
	return pattern[0] == rel[0] && matchLayerLocationParts(pattern[1:], rel[1:])
}

func matchesDomainLayer(arch core.Architecture, rel, name string) bool {
	parts := strings.Split(rel, "/")
	return len(parts) == 4 && parts[0] == arch.Layout.InternalRoot && parts[1] == arch.Layout.DomainDir && parts[2] != "" && parts[3] == name
}

var _ core.Rule = (*LayerPlacement)(nil)
