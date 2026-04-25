package structural

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/NamhaeSusan/go-arch-guard/core"
)

const (
	ruleInternalTopLevel = "structural.internal-top-level"
	internalTopLevel     = "structure.internal-top-level"
)

type InternalTopLevel struct {
	severity core.Severity
}

func NewInternalTopLevel(opts ...Option) *InternalTopLevel {
	cfg := newConfig(opts, core.Error)
	return &InternalTopLevel{severity: cfg.severity}
}

func (r *InternalTopLevel) Spec() core.RuleSpec {
	return withSeverity(core.RuleSpec{
		ID:              ruleInternalTopLevel,
		Description:     "internal top-level entries must match the configured allowlist",
		DefaultSeverity: r.severity,
		Violations: []core.ViolationSpec{
			{ID: internalTopLevel, Description: "unexpected internal top-level entry"},
		},
	}, r.severity)
}

func (r *InternalTopLevel) Check(ctx *core.Context) []core.Violation {
	if ctx == nil {
		return nil
	}
	arch := ctx.Arch()
	if !hasInternalDir(ctx.Root(), arch.Layout.InternalRoot) {
		return []core.Violation{metaLayoutNotSupported(ruleInternalTopLevel)}
	}
	internalDir := filepath.Join(ctx.Root(), arch.Layout.InternalRoot)
	entries, err := os.ReadDir(internalDir)
	if err != nil {
		return nil
	}

	allowedNames := make([]string, 0, len(arch.Layers.InternalTopLevel))
	for name := range arch.Layers.InternalTopLevel {
		allowedNames = append(allowedNames, arch.Layout.InternalRoot+"/"+name+"/")
	}
	sort.Strings(allowedNames)
	allowedHint := strings.Join(allowedNames, ", ")

	var violations []core.Violation
	for _, entry := range entries {
		relPath := filepath.ToSlash(filepath.Join(arch.Layout.InternalRoot, entry.Name()))
		if entry.IsDir() {
			if ctx.IsExcluded(relPath+"/") || arch.Layers.InternalTopLevel[entry.Name()] {
				continue
			}
			violations = append(violations, violation(r.severity, internalTopLevel, relPath+"/",
				arch.Layout.InternalRoot+`/ top-level package "`+entry.Name()+`" is not allowed`,
				"use only "+allowedHint+" at the "+arch.Layout.InternalRoot+"/ top level"))
			continue
		}

		if !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		if ctx.IsExcluded(relPath) {
			continue
		}
		violations = append(violations, violation(r.severity, internalTopLevel, relPath,
			arch.Layout.InternalRoot+`/ top-level Go file "`+entry.Name()+`" is not allowed`,
			"move code under "+allowedHint))
	}
	return violations
}
