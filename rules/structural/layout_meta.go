package structural

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/rules/internal/rulemeta"
)

// hasInternalDir reports whether <root>/<internalRoot> exists as a
// directory. The structural rules walk the filesystem rather than
// ctx.Pkgs(), so this is a pre-flight check used to decide whether the
// rule applies at all.
func hasInternalDir(root, internalRoot string) bool {
	if root == "" {
		return false
	}
	info, err := os.Stat(filepath.Join(root, internalRoot))
	return err == nil && info.IsDir()
}

func metaLayoutNotSupported(ruleID string) core.Violation {
	return rulemeta.LayoutNotSupported(
		fmt.Sprintf("%s requires an internal/-based layout; internal/ directory not found", ruleID),
		"use this rule with internal/-based presets (DDD, CleanArch, Hexagonal, ModularMonolith), or remove it from your ruleset for flat layouts")
}

// metaRuleDisabledByConfig signals that a rule is registered in the RuleSet
// but the supplied core.Architecture configuration prevents it from running
// (whole rule) or makes a sub-check inert (partial). Severity defaults to
// Warning via the runner's meta.* prefix handling.
func metaRuleDisabledByConfig(ruleID, reason, fix string) core.Violation {
	return rulemeta.RuleDisabledByConfig(ruleID, reason, fix)
}
