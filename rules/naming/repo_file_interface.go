package naming

import (
	"go/ast"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/core/analysisutil"
	"golang.org/x/tools/go/packages"
)

type RepoFileInterface struct {
	severity core.Severity
}

func NewRepoFileInterface(opts ...Option) *RepoFileInterface {
	cfg := newConfig(opts, core.Error)
	return &RepoFileInterface{severity: cfg.severity}
}

func (r *RepoFileInterface) Spec() core.RuleSpec {
	return core.RuleSpec{
		ID:              "naming.repo-file-interface",
		Description:     "repository port interface placement and filename conventions",
		DefaultSeverity: r.severity,
		Violations: []core.ViolationSpec{
			{ID: "structure.repo-file-interface", Description: "repo file must contain matching interface", DefaultSeverity: r.severity},
			{ID: "structure.repo-file-extra-interface", Description: "repo file must define one interface", DefaultSeverity: r.severity},
			{ID: "structure.interface-placement", Description: "repository-port interfaces must live in the port layer", DefaultSeverity: r.severity},
		},
	}
}

func (r *RepoFileInterface) Check(ctx *core.Context) []core.Violation {
	layers := ctx.Arch().Layers
	if !analysisutil.HasPortSublayer(layers) {
		return nil
	}
	var violations []core.Violation
	for _, pkg := range ctx.Pkgs() {
		if analysisutil.MatchPortSublayer(layers, pkg.PkgPath) != "" {
			violations = append(violations, r.checkPortPackage(ctx, pkg)...)
			continue
		}
		violations = append(violations, r.checkInterfacePlacement(ctx, pkg)...)
	}
	return violations
}

func (r *RepoFileInterface) checkPortPackage(ctx *core.Context, pkg *packages.Package) []core.Violation {
	var violations []core.Violation
	for _, file := range pkg.Syntax {
		filename := pkg.Fset.Position(file.Pos()).Filename
		base := filepath.Base(filename)
		if strings.HasSuffix(base, "_test.go") {
			continue
		}
		relPath := analysisutil.RelativePathForPackage(pkg, filename)
		if ctx.IsExcluded(relPath) {
			continue
		}
		expected := analysisutil.SnakeToPascal(strings.TrimSuffix(base, ".go"))
		ifaces := collectInterfacesFromFile(file)
		if _, ok := ifaces[expected]; !ok {
			violations = append(violations, r.violation(
				relPath,
				0,
				"structure.repo-file-interface",
				`file "`+base+`" in repo/ must contain interface "`+expected+`"`,
				`add "type `+expected+` interface { ... }" or rename the file`,
			))
		}
		if len(ifaces) <= 1 {
			continue
		}
		extra := make([]string, 0, len(ifaces)-1)
		for name := range ifaces {
			if name != expected {
				extra = append(extra, name)
			}
		}
		sort.Strings(extra)
		violations = append(violations, r.violation(
			relPath,
			0,
			"structure.repo-file-extra-interface",
			`file "`+base+`" in repo/ must define only "`+expected+`", found extra: `+strings.Join(extra, ", "),
			"move each extra interface to its own file (e.g. "+pascalToSnake(extra[0])+".go)",
		))
	}
	return violations
}

func (r *RepoFileInterface) checkInterfacePlacement(ctx *core.Context, pkg *packages.Package) []core.Violation {
	arch := ctx.Arch()
	if !arch.Structure.RequireAlias || !isDomainPackage(arch, pkg.PkgPath) {
		return nil
	}
	repoName := analysisutil.PortSublayerName(arch.Layers)
	var violations []core.Violation
	for _, file := range pkg.Syntax {
		filePath := analysisutil.RelativePathForPackage(pkg, pkg.Fset.Position(file.Pos()).Filename)
		if ctx.IsExcluded(filePath) {
			continue
		}
		for _, info := range analysisutil.InspectTypeSpecs(file, pkg.Fset) {
			if info.IsInterface && isRepoPortName(info.Name) {
				violations = append(violations, r.violation(
					filePath,
					info.Line,
					"structure.interface-placement",
					`interface "`+info.Name+`" matches repository-port naming and must be defined in `+repoName+`/, not in `+path.Base(path.Dir(pkg.PkgPath))+`/`,
					"move to "+repoName+"/, or rename if it's a consumer-defined interface",
				))
			}
			if info.AliasFrom != "" && analysisutil.MatchPortSublayer(arch.Layers, info.AliasFrom) != "" {
				violations = append(violations, r.violation(
					filePath,
					info.Line,
					"structure.interface-placement",
					`type alias "`+info.Name+`" re-exports interface from `+repoName+` - suspected cross-domain dependency; use `+arch.Layout.OrchestrationDir+`/ instead`,
					"remove alias and move cross-domain coordination to "+arch.Layout.OrchestrationDir+"/",
				))
			}
		}
	}
	return violations
}

func (r *RepoFileInterface) violation(file string, line int, rule, message, fix string) core.Violation {
	return core.Violation{
		File:              file,
		Line:              line,
		Rule:              rule,
		Message:           message,
		Fix:               fix,
		DefaultSeverity:   r.severity,
		EffectiveSeverity: r.severity,
	}
}

func collectInterfacesFromFile(file *ast.File) map[string]*ast.InterfaceType {
	result := make(map[string]*ast.InterfaceType)
	for _, decl := range file.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range gd.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			if iface, ok := ts.Type.(*ast.InterfaceType); ok {
				result[ts.Name.Name] = iface
			}
		}
	}
	return result
}

func isDomainPackage(arch core.Architecture, pkgPath string) bool {
	return arch.Layout.DomainDir != "" && strings.Contains(pkgPath, "/internal/"+arch.Layout.DomainDir+"/")
}

func isRepoPortName(name string) bool {
	return strings.HasSuffix(name, "Repository") || strings.HasSuffix(name, "Repo")
}

var _ core.Rule = (*RepoFileInterface)(nil)
