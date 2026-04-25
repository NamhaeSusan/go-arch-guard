package naming

import (
	"go/ast"
	"strings"
	"unicode"

	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/core/analysisutil"
)

type Option func(*ruleConfig)

type ruleConfig struct {
	severity core.Severity
}

func WithSeverity(severity core.Severity) Option {
	return func(cfg *ruleConfig) {
		cfg.severity = severity
	}
}

func newConfig(opts []Option, severity core.Severity) ruleConfig {
	cfg := ruleConfig{severity: severity}
	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg
}

type NoStutter struct {
	severity core.Severity
}

func NewNoStutter(opts ...Option) *NoStutter {
	cfg := newConfig(opts, core.Warning)
	return &NoStutter{severity: cfg.severity}
}

func (r *NoStutter) Spec() core.RuleSpec {
	return core.RuleSpec{
		ID:              "naming.no-stutter",
		Description:     "exported type names must not repeat their package name",
		DefaultSeverity: r.severity,
	}
}

func (r *NoStutter) Check(ctx *core.Context) []core.Violation {
	var violations []core.Violation
	for _, pkg := range ctx.Pkgs() {
		pkgName := pkg.Name
		pkgNameLen := len([]rune(pkgName))
		for _, file := range pkg.Syntax {
			filePath := analysisutil.RelativePathForPackage(pkg, pkg.Fset.Position(file.Pos()).Filename)
			if ctx.IsExcluded(filePath) {
				continue
			}
			for _, decl := range file.Decls {
				gd, ok := decl.(*ast.GenDecl)
				if !ok {
					continue
				}
				for _, spec := range gd.Specs {
					ts, ok := spec.(*ast.TypeSpec)
					if !ok || !ts.Name.IsExported() {
						continue
					}
					name := ts.Name.Name
					if !stutters(pkgName, name) {
						continue
					}
					suggested := string([]rune(name)[pkgNameLen:])
					pos := pkg.Fset.Position(ts.Name.Pos())
					violations = append(violations, r.violation(
						analysisutil.RelativePathForPackage(pkg, pos.Filename),
						pos.Line,
						`type "`+name+`" stutters with package "`+pkgName+`"`,
						`rename to "`+suggested+`"`,
					))
				}
			}
		}
	}
	return violations
}

func (r *NoStutter) violation(file string, line int, message, fix string) core.Violation {
	return core.Violation{
		File:              file,
		Line:              line,
		Rule:              "naming.no-stutter",
		Message:           message,
		Fix:               fix,
		DefaultSeverity:   r.severity,
		EffectiveSeverity: r.severity,
	}
}

func stutters(pkgName, typeName string) bool {
	pkgRunes := []rune(pkgName)
	typeRunes := []rune(typeName)
	if len(typeRunes) <= len(pkgRunes) {
		return false
	}
	prefix := strings.ToLower(string(typeRunes[:len(pkgRunes)]))
	if prefix != strings.ToLower(pkgName) {
		return false
	}
	return unicode.IsUpper(typeRunes[len(pkgRunes)])
}

var _ core.Rule = (*NoStutter)(nil)
