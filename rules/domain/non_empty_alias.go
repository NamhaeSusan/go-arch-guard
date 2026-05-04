package domain

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"

	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/core/analysisutil"
)

const ruleNonEmptyAlias = "domain.non-empty-alias"

type NonEmptyAlias struct {
	cfg ruleConfig
}

func NewNonEmptyAlias(opts ...Option) *NonEmptyAlias {
	return &NonEmptyAlias{cfg: newConfig(opts, core.Warning)}
}

func (r *NonEmptyAlias) Spec() core.RuleSpec {
	return core.RuleSpec{
		ID:              ruleNonEmptyAlias,
		Description:     "domain alias files should expose a non-empty public surface when the domain has an app package",
		DefaultSeverity: r.cfg.severity,
		Violations: []core.ViolationSpec{
			{
				ID:              ruleNonEmptyAlias,
				Description:     "domain alias file has no exported public surface",
				DefaultSeverity: r.cfg.severity,
			},
		},
	}
}

func (r *NonEmptyAlias) Check(ctx *core.Context) []core.Violation {
	if ctx == nil {
		return nil
	}
	arch := ctx.Arch()
	root := analysisutil.ResolveRootFromContext(ctx, "")
	if root == "" {
		root = ctx.Root()
	}
	if !hasDir(filepath.Join(root, arch.Layout.InternalRoot)) {
		return []core.Violation{metaLayoutNotSupported(ruleNonEmptyAlias, root, arch.Layout.InternalRoot)}
	}
	if arch.Layout.DomainDir == "" {
		return []core.Violation{metaRuleDisabledByConfig(ruleNonEmptyAlias,
			"Layout.DomainDir is empty (flat layout); alias surface detection requires a domain directory",
			"set Layout.DomainDir to your domain root, or remove domain.NewNonEmptyAlias() from your RuleSet")}
	}

	domainRoot := filepath.Join(root, arch.Layout.InternalRoot, filepath.FromSlash(arch.Layout.DomainDir))
	entries, err := os.ReadDir(domainRoot)
	if err != nil {
		return nil
	}

	var violations []core.Violation
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		domainName := entry.Name()
		domainRel := filepath.ToSlash(filepath.Join(arch.Layout.InternalRoot, arch.Layout.DomainDir, domainName))
		if ctx.IsExcluded(domainRel + "/") {
			continue
		}
		domainDir := filepath.Join(domainRoot, domainName)
		if !r.shouldRequireSurface(domainDir, arch) {
			continue
		}
		aliasName := aliasFileName(arch)
		aliasPath := filepath.Join(domainDir, aliasName)
		aliasRel := domainRel + "/" + aliasName
		if !hasFile(aliasPath) {
			continue
		}
		count, line, ok := exportedSurface(aliasPath)
		if !ok || count > 0 {
			continue
		}
		violations = append(violations, core.Violation{
			File:              aliasRel,
			Line:              line,
			Rule:              ruleNonEmptyAlias,
			Message:           fmt.Sprintf("domain %q has an app package but %s exposes no exported alias surface", domainName, aliasName),
			Fix:               "re-export at least one service, constructor, type, const, or var from the domain alias file",
			DefaultSeverity:   r.cfg.severity,
			EffectiveSeverity: r.cfg.severity,
		})
	}
	return violations
}

func (r *NonEmptyAlias) shouldRequireSurface(domainDir string, arch core.Architecture) bool {
	if r.cfg.requirePlaceholderAliases {
		return true
	}
	appDir := arch.Layout.AppDir
	if appDir == "" {
		appDir = "app"
	}
	return hasNonTestGoFiles(filepath.Join(domainDir, filepath.FromSlash(appDir)))
}

func exportedSurface(path string) (int, int, bool) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		return 0, 0, false
	}
	line := fset.Position(file.Package).Line
	var count int
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if d.Name.IsExported() {
				count++
			}
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					if s.Name.IsExported() {
						count++
					}
				case *ast.ValueSpec:
					for _, name := range s.Names {
						if name.IsExported() {
							count++
						}
					}
				}
			}
		}
	}
	return count, line, true
}

func aliasFileName(arch core.Architecture) string {
	if arch.Naming.AliasFileName != "" {
		return arch.Naming.AliasFileName
	}
	return "alias.go"
}

func hasDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func hasFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func hasNonTestGoFiles(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || filepath.Ext(name) != ".go" || len(name) >= len("_test.go") && name[len(name)-len("_test.go"):] == "_test.go" {
			continue
		}
		return true
	}
	return false
}

var _ core.Rule = (*NonEmptyAlias)(nil)
