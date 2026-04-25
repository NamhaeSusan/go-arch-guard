package structural

import (
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/NamhaeSusan/go-arch-guard/core"
)

const (
	rulePlacement       = "structural.placement"
	misplacedLayer      = "structure.misplaced-layer"
	middlewarePlacement = "structure.middleware-placement"
	dtoPlacement        = "structure.dto-placement"
)

type Placement struct {
	severity core.Severity
}

func NewPlacement(opts ...Option) *Placement {
	cfg := newConfig(opts, core.Error)
	return &Placement{severity: cfg.severity}
}

func (r *Placement) Spec() core.RuleSpec {
	return withSeverity(core.RuleSpec{
		ID:              rulePlacement,
		Description:     "layer, middleware, and DTO files must stay in configured structural locations",
		DefaultSeverity: r.severity,
		Violations: []core.ViolationSpec{
			{ID: misplacedLayer, Description: "layer package is outside a valid slice"},
			{ID: middlewarePlacement, Description: "middleware package is outside the shared middleware directory"},
			{ID: dtoPlacement, Description: "DTO file is outside an allowed layer"},
		},
	}, r.severity)
}

func (r *Placement) Check(ctx *core.Context) []core.Violation {
	if ctx == nil {
		return nil
	}
	if !hasInternalDir(ctx.Root()) {
		return []core.Violation{metaLayoutNotSupported(rulePlacement)}
	}
	internalDir := filepath.Join(ctx.Root(), "internal")
	arch := ctx.Arch()
	violations := r.checkLayerPlacement(ctx, internalDir, arch)
	violations = append(violations, r.checkMiddlewarePlacement(ctx, internalDir, arch)...)
	if arch.Layout.DomainDir != "" {
		violations = append(violations, r.checkDTOPlacement(ctx, internalDir, arch)...)
	}
	return violations
}

func (r *Placement) checkLayerPlacement(ctx *core.Context, internalDir string, arch core.Architecture) []core.Violation {
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
		if rel == "internal" {
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

func (r *Placement) checkMiddlewarePlacement(ctx *core.Context, internalDir string, arch core.Architecture) []core.Violation {
	allowedPath := filepath.ToSlash(filepath.Join("internal", arch.Layout.SharedDir, "middleware"))
	var violations []core.Violation
	_ = filepath.WalkDir(internalDir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || !entry.IsDir() || entry.Name() != "middleware" {
			return nil
		}
		if !hasNonTestGoFiles(path) {
			return nil
		}
		rel, relErr := filepath.Rel(filepath.Dir(internalDir), path)
		if relErr != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		if ctx.IsExcluded(rel+"/") || rel == allowedPath {
			return nil
		}
		violations = append(violations, violation(r.severity, middlewarePlacement, rel+"/",
			`middleware found at "`+rel+`"`,
			"move middleware to "+allowedPath+"/"))
		return nil
	})
	return violations
}

func (r *Placement) checkDTOPlacement(ctx *core.Context, internalDir string, arch core.Architecture) []core.Violation {
	domainDir := filepath.Join(internalDir, filepath.FromSlash(arch.Layout.DomainDir))
	if _, err := os.Stat(domainDir); err != nil {
		return nil
	}
	var violations []core.Violation
	_ = filepath.WalkDir(domainDir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return nil
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".go") {
			return nil
		}
		if name != "dto.go" && (!strings.HasSuffix(name, "_dto.go") || strings.HasSuffix(name, "_test.go")) {
			return nil
		}
		rel, relErr := filepath.Rel(filepath.Dir(internalDir), path)
		if relErr != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		if ctx.IsExcluded(rel) || isDTOAllowedSublayer(arch, rel) {
			return nil
		}
		violations = append(violations, violation(r.severity, dtoPlacement, rel,
			`"`+name+`" found in forbidden layer`,
			"DTOs belong in "+strings.Join(arch.Structure.DTOAllowedLayers, ", ")))
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
		if name == "app" && arch.Layout.AppDir != "" && rel == filepath.ToSlash(filepath.Join("internal", arch.Layout.AppDir)) {
			return false
		}
		return !matchesDomainLayer(arch, rel, name)
	case "handler":
		if arch.Layout.ServerDir != "" {
			serverBase := filepath.ToSlash(filepath.Join("internal", arch.Layout.ServerDir))
			if strings.HasPrefix(rel, serverBase+"/") || rel == serverBase {
				return false
			}
		}
		return !matchesDomainLayer(arch, rel, name) &&
			rel != filepath.ToSlash(filepath.Join("internal", arch.Layout.OrchestrationDir, "handler"))
	default:
		return false
	}
}

func matchesDomainLayer(arch core.Architecture, rel, name string) bool {
	parts := strings.Split(rel, "/")
	return len(parts) == 4 && parts[0] == "internal" && parts[1] == arch.Layout.DomainDir && parts[2] != "" && parts[3] == name
}

func isDTOAllowedSublayer(arch core.Architecture, rel string) bool {
	domainDepth := len(strings.Split(arch.Layout.DomainDir, "/"))
	parts := strings.Split(rel, "/")
	sublayerIdx := 1 + domainDepth + 1
	if len(parts) <= sublayerIdx {
		return false
	}
	return slices.Contains(arch.Structure.DTOAllowedLayers, parts[sublayerIdx])
}
