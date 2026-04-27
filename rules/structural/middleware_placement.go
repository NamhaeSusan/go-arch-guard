package structural

import (
	"io/fs"
	"path/filepath"

	"github.com/NamhaeSusan/go-arch-guard/core"
)

const (
	ruleMiddlewarePlacement = "structural.middleware-placement"
	middlewarePlacement     = "structural.middleware-placement"
	defaultMiddlewareDir    = "middleware"
)

// MiddlewarePlacement flags directories named "<MiddlewareDir>" (default
// "middleware") found anywhere except internal/<SharedDir>/<MiddlewareDir>/.
// Override the directory name with WithMiddlewareDir.
type MiddlewarePlacement struct {
	severity      core.Severity
	middlewareDir string
}

func NewMiddlewarePlacement(opts ...Option) *MiddlewarePlacement {
	cfg := newConfig(opts, core.Error)
	dir := cfg.middlewareDir
	if dir == "" {
		dir = defaultMiddlewareDir
	}
	return &MiddlewarePlacement{severity: cfg.severity, middlewareDir: dir}
}

func (r *MiddlewarePlacement) Spec() core.RuleSpec {
	return withSeverity(core.RuleSpec{
		ID:              ruleMiddlewarePlacement,
		Description:     "middleware packages must live in the shared middleware directory",
		DefaultSeverity: r.severity,
	}, r.severity)
}

func (r *MiddlewarePlacement) Check(ctx *core.Context) []core.Violation {
	if ctx == nil {
		return nil
	}
	arch := ctx.Arch()
	if !hasInternalDir(ctx.Root(), arch.Layout.InternalRoot) {
		return []core.Violation{metaLayoutNotSupported(ruleMiddlewarePlacement)}
	}
	internalDir := filepath.Join(ctx.Root(), arch.Layout.InternalRoot)
	allowedPath := filepath.ToSlash(filepath.Join(arch.Layout.InternalRoot, arch.Layout.SharedDir, r.middlewareDir))
	var violations []core.Violation
	_ = filepath.WalkDir(internalDir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || !entry.IsDir() || entry.Name() != r.middlewareDir {
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
			r.middlewareDir+` found at "`+rel+`"`,
			"move "+r.middlewareDir+" to "+allowedPath+"/"))
		return nil
	})
	return violations
}

var _ core.Rule = (*MiddlewarePlacement)(nil)
