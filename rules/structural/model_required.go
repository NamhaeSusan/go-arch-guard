package structural

import (
	"os"
	"path/filepath"

	"github.com/NamhaeSusan/go-arch-guard/core"
)

const (
	ruleModelRequired = "structural.model-required"
	modelRequired     = "structural.domain-model-required"
)

type ModelRequired struct {
	severity core.Severity
}

func NewModelRequired(opts ...Option) *ModelRequired {
	cfg := newConfig(opts, core.Error)
	return &ModelRequired{severity: cfg.severity}
}

func (r *ModelRequired) Spec() core.RuleSpec {
	return withSeverity(core.RuleSpec{
		ID:              ruleModelRequired,
		Description:     "domain roots must contain a direct model file in the configured model path",
		DefaultSeverity: r.severity,
		Violations: []core.ViolationSpec{
			{ID: modelRequired, Description: "domain root is missing a direct model Go file"},
		},
	}, r.severity)
}

func (r *ModelRequired) Check(ctx *core.Context) []core.Violation {
	if ctx == nil {
		return nil
	}
	arch := ctx.Arch()
	if !hasInternalDir(ctx.Root(), arch.Layout.InternalRoot) {
		return []core.Violation{metaLayoutNotSupported(ruleModelRequired)}
	}
	if arch.Layout.DomainDir == "" {
		return []core.Violation{metaRuleDisabledByConfig(ruleModelRequired,
			"Layout.DomainDir is empty (flat layout); model enforcement requires a domain directory",
			"set Layout.DomainDir to your domain root, or remove structural.NewModelRequired() from your RuleSet")}
	}
	if !arch.Structure.RequireModel {
		return []core.Violation{metaRuleDisabledByConfig(ruleModelRequired,
			"Structure.RequireModel is false; model enforcement skipped",
			"set Structure.RequireModel = true, or remove structural.NewModelRequired() from your RuleSet")}
	}

	domainDir := filepath.Join(ctx.Root(), arch.Layout.InternalRoot, filepath.FromSlash(arch.Layout.DomainDir))
	entries, err := os.ReadDir(domainDir)
	if err != nil {
		return nil
	}
	var violations []core.Violation
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		relPath := filepath.ToSlash(filepath.Join(arch.Layout.InternalRoot, arch.Layout.DomainDir, entry.Name()))
		if ctx.IsExcluded(relPath + "/") {
			continue
		}
		modelDir := filepath.Join(domainDir, entry.Name(), filepath.FromSlash(arch.Structure.ModelPath))
		if hasNonTestGoFiles(modelDir) {
			continue
		}
		violations = append(violations, violation(r.severity, modelRequired, relPath+"/",
			`domain "`+entry.Name()+`" missing a direct non-test Go file in `+arch.Structure.ModelPath+`/`,
			"add at least one non-test Go file directly under "+arch.Structure.ModelPath+"/"))
	}
	return violations
}
