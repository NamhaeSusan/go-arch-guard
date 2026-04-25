package structural

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/NamhaeSusan/go-arch-guard/core"
)

// hasInternalDir reports whether <root>/internal exists as a directory. The
// structural rules walk the filesystem rather than ctx.Pkgs(), so this is a
// pre-flight check used to decide whether the rule applies at all.
func hasInternalDir(root string) bool {
	if root == "" {
		return false
	}
	info, err := os.Stat(filepath.Join(root, "internal"))
	return err == nil && info.IsDir()
}

func metaLayoutNotSupported(ruleID string) core.Violation {
	return core.Violation{
		Rule:              "meta.layout-not-supported",
		Message:           fmt.Sprintf("%s requires an internal/-based layout; internal/ directory not found", ruleID),
		Fix:               "use this rule with internal/-based presets (DDD, CleanArch, Hexagonal, ModularMonolith), or remove it from your ruleset for flat layouts",
		DefaultSeverity:   core.Warning,
		EffectiveSeverity: core.Warning,
	}
}
