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
	if idx := strings.IndexByte(name, '.'); idx > 0 {
		name = name[:idx]
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
	return pascalToSnake(strings.TrimSuffix(filename, ext)) + ext
}

// pascalToSnake converts a PascalCase / camelCase / mixed identifier to
// snake_case. Runs of uppercase are kept together as one word; an underscore
// is inserted only at a real word boundary:
//   - between a lowercase/digit and an uppercase ("createOrder" -> "create_order")
//   - between an uppercase and an uppercase that is followed by a lowercase
//     ("XMLParser" -> "xml_parser")
func pascalToSnake(name string) string {
	runes := []rune(name)
	var result []rune
	for i, r := range runes {
		if i > 0 && unicode.IsUpper(r) {
			prev := runes[i-1]
			nextLower := i+1 < len(runes) && unicode.IsLower(runes[i+1])
			if unicode.IsLower(prev) || unicode.IsDigit(prev) || (unicode.IsUpper(prev) && nextLower) {
				result = append(result, '_')
			}
		}
		result = append(result, unicode.ToLower(r))
	}
	return string(result)
}

var _ core.Rule = (*SnakeCaseFiles)(nil)
