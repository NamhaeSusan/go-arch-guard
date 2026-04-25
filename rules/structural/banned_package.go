package structural

import (
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/NamhaeSusan/go-arch-guard/core"
)

const (
	ruleBannedPackage = "structural.banned-package"
	bannedPackage     = "structure.banned-package"
	legacyPackage     = "structure.legacy-package"
)

type BannedPackage struct {
	severity core.Severity
}

func NewBannedPackage(opts ...Option) *BannedPackage {
	cfg := newConfig(opts, core.Warning)
	return &BannedPackage{severity: cfg.severity}
}

func (r *BannedPackage) Spec() core.RuleSpec {
	return withSeverity(core.RuleSpec{
		ID:              ruleBannedPackage,
		Description:     "banned and legacy package names must not appear under internal",
		DefaultSeverity: r.severity,
		Violations: []core.ViolationSpec{
			{ID: bannedPackage, Description: "package name is banned"},
			{ID: legacyPackage, Description: "legacy package should be migrated"},
		},
	}, r.severity)
}

func (r *BannedPackage) Check(ctx *core.Context) []core.Violation {
	if ctx == nil {
		return nil
	}
	if !hasInternalDir(ctx.Root()) {
		return []core.Violation{metaLayoutNotSupported(ruleBannedPackage)}
	}
	internalDir := filepath.Join(ctx.Root(), "internal")
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
		arch := ctx.Arch()
		for _, banned := range arch.Naming.BannedPkgNames {
			if name == banned {
				violations = append(violations, violation(r.severity, bannedPackage, rel+"/",
					`package "`+name+`" is banned`,
					fmt.Sprintf("move to specific domain or %s/", arch.Layout.SharedDir)))
			}
		}
		for _, legacy := range arch.Naming.LegacyPkgNames {
			if name == legacy {
				violations = append(violations, violation(r.severity, legacyPackage, rel+"/",
					`legacy package "`+name+`" should be migrated`,
					fmt.Sprintf("move app-specific wiring to cmd/ and shared helpers to internal/%s/", arch.Layout.SharedDir)))
			}
		}
		return nil
	})
	return violations
}
