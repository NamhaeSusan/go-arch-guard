package naming

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/core/analysisutil"
)

type NoLayerSuffix struct {
	severity core.Severity
}

func NewNoLayerSuffix(opts ...Option) *NoLayerSuffix {
	cfg := newConfig(opts, core.Warning)
	return &NoLayerSuffix{severity: cfg.severity}
}

func (r *NoLayerSuffix) Spec() core.RuleSpec {
	return core.RuleSpec{
		ID:              "naming.no-layer-suffix",
		Description:     "layer files must not repeat their layer name as a filename suffix",
		DefaultSeverity: r.severity,
	}
}

func (r *NoLayerSuffix) Check(ctx *core.Context) []core.Violation {
	banned := layerSuffixes(ctx.Arch().Layers.LayerDirNames)
	var violations []core.Violation
	seen := make(map[string]bool)
	for _, pkg := range ctx.Pkgs() {
		for _, file := range pkg.GoFiles {
			if seen[file] {
				continue
			}
			seen[file] = true
			base := filepath.Base(file)
			if strings.HasSuffix(base, "_test.go") {
				continue
			}
			relPath := analysisutil.RelativePathForPackage(pkg, file)
			if ctx.IsExcluded(relPath) {
				continue
			}
			dir := filepath.Base(filepath.Dir(file))
			if !ctx.Arch().Layers.LayerDirNames[dir] || isTypePatternFile(ctx.Arch().Structure.TypePatterns, dir, base) {
				continue
			}
			name := strings.TrimSuffix(base, ".go")
			for _, suffix := range banned {
				trimmed, ok := strings.CutSuffix(name, suffix)
				if !ok {
					continue
				}
				violations = append(violations, core.Violation{
					File:              relPath,
					Rule:              "naming.no-layer-suffix",
					Message:           `filename "` + base + `" has redundant layer suffix "` + suffix + `"`,
					Fix:               `rename to "` + trimmed + `.go"`,
					DefaultSeverity:   r.severity,
					EffectiveSeverity: r.severity,
				})
				break
			}
		}
	}
	return violations
}

func layerSuffixes(layerDirNames map[string]bool) []string {
	out := make([]string, 0, len(layerDirNames))
	for dir, enabled := range layerDirNames {
		if enabled && dir != "" {
			out = append(out, "_"+dir)
		}
	}
	sort.Strings(out)
	return out
}

func isTypePatternFile(patterns []core.TypePattern, dir, filename string) bool {
	name := strings.TrimSuffix(filename, ".go")
	for _, pattern := range patterns {
		if pattern.Dir == dir && strings.HasPrefix(name, pattern.FilePrefix+"_") {
			return true
		}
	}
	return false
}

var _ core.Rule = (*NoLayerSuffix)(nil)
