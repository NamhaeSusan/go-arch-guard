package naming

import (
	"path/filepath"
	"strings"
	"unicode"

	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/core/analysisutil"
)

type SnakeCaseFiles struct {
	severity core.Severity
}

func NewSnakeCaseFiles(opts ...Option) *SnakeCaseFiles {
	cfg := newConfig(opts, core.Warning)
	return &SnakeCaseFiles{severity: cfg.severity}
}

func (r *SnakeCaseFiles) Spec() core.RuleSpec {
	return core.RuleSpec{
		ID:              "naming.snake-case-file",
		Description:     "Go filenames must use snake_case",
		DefaultSeverity: r.severity,
	}
}

func (r *SnakeCaseFiles) Check(ctx *core.Context) []core.Violation {
	var violations []core.Violation
	seen := make(map[string]bool)
	for _, pkg := range ctx.Pkgs() {
		for _, file := range pkg.GoFiles {
			if seen[file] {
				continue
			}
			seen[file] = true
			relPath := analysisutil.RelativePathForPackage(pkg, file)
			if ctx.IsExcluded(relPath) {
				continue
			}
			base := filepath.Base(file)
			if isSnakeCase(base) {
				continue
			}
			violations = append(violations, core.Violation{
				File:              relPath,
				Rule:              "naming.snake-case-file",
				Message:           `filename "` + base + `" must be snake_case`,
				Fix:               `rename to "` + toSnakeCase(base) + `"`,
				DefaultSeverity:   r.severity,
				EffectiveSeverity: r.severity,
			})
		}
	}
	return violations
}

func isSnakeCase(filename string) bool {
	name := strings.TrimSuffix(filename, ".go")
	name = strings.TrimSuffix(name, "_test")
	// Generated suffixes such as .pb.go are valid, but every segment must
	// obey the same naming policy.
	if strings.Contains(name, ".") {
		for _, part := range strings.Split(name, ".") {
			if part == "" || !isSnakeCase(part) {
				return false
			}
		}
		return true
	}
	if name == "" {
		return false
	}
	for _, r := range name {
		if !unicode.IsLower(r) && !unicode.IsDigit(r) && r != '_' {
			return false
		}
	}
	return true
}

func toSnakeCase(filename string) string {
	ext := filepath.Ext(filename)
	return analysisutil.PascalToSnake(strings.TrimSuffix(filename, ext)) + ext
}

var _ core.Rule = (*SnakeCaseFiles)(nil)
