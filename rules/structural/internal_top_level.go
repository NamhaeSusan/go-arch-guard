package structural

import (
	"fmt"
	"os"
	"path/filepath"
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
	r := &InternalTopLevel{severity: core.Error}
	applyOptions(r, opts)
	return r
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
	internalDir := filepath.Join(ctx.Root(), "internal")
	entries, err := os.ReadDir(internalDir)
	if err != nil {
		return nil
	}

	arch := ctx.Arch()
	allowedNames := make([]string, 0, len(arch.Layers.InternalTopLevel))
	for name := range arch.Layers.InternalTopLevel {
		allowedNames = append(allowedNames, "internal/"+name+"/")
	}

	var violations []core.Violation
	for _, entry := range entries {
		relPath := filepath.ToSlash(filepath.Join("internal", entry.Name()))
		if entry.IsDir() {
			if ctx.IsExcluded(relPath+"/") || arch.Layers.InternalTopLevel[entry.Name()] {
				continue
			}
			violations = append(violations, violation(r.severity, internalTopLevel, relPath+"/",
				`internal/ top-level package "`+entry.Name()+`" is not allowed`,
				fmt.Sprintf("use only %v at the internal/ top level", allowedNames)))
			continue
		}

		if !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		if ctx.IsExcluded(relPath) {
			continue
		}
		violations = append(violations, violation(r.severity, internalTopLevel, relPath,
			`internal/ top-level Go file "`+entry.Name()+`" is not allowed`,
			fmt.Sprintf("move code under %v", allowedNames)))
	}
	return violations
}
