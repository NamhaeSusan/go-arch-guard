package structural

import (
	"go/ast"
	"path/filepath"
	"sort"
	"strings"

	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/core/analysisutil"
	"golang.org/x/tools/go/packages"
)

type RepoFileInterface struct {
	severity     core.Severity
	portSuffixes []string
}

var defaultRepoPortSuffixes = []string{"Repository", "Repo"}

func NewRepoFileInterface(opts ...Option) *RepoFileInterface {
	cfg := newConfig(opts, core.Error)
	return &RepoFileInterface{
		severity:     cfg.severity,
		portSuffixes: cfg.repoPortSuffixes,
	}
}

func (r *RepoFileInterface) suffixes() []string {
	if len(r.portSuffixes) == 0 {
		return defaultRepoPortSuffixes
	}
	return r.portSuffixes
}

func (r *RepoFileInterface) isRepoPortName(name string) bool {
	for _, suf := range r.suffixes() {
		if strings.HasSuffix(name, suf) {
			return true
		}
	}
	return false
}

func (r *RepoFileInterface) Spec() core.RuleSpec {
	return core.RuleSpec{
		ID:              "structural.repo-file-interface",
		Description:     "repository port interface placement and filename conventions",
		DefaultSeverity: r.severity,
		Violations: []core.ViolationSpec{
			{ID: "structural.repo-file-interface-missing", Description: "repo file must contain matching interface", DefaultSeverity: r.severity},
			{ID: "structural.repo-file-extra-interface", Description: "repo file must define one interface", DefaultSeverity: r.severity},
			{ID: "structural.interface-placement", Description: "repository-port interfaces must live in the port layer", DefaultSeverity: r.severity},
		},
	}
}

func (r *RepoFileInterface) Check(ctx *core.Context) []core.Violation {
	layers := ctx.Arch().Layers
	if !analysisutil.HasPortSublayer(layers) {
		return []core.Violation{metaRuleDisabledByConfig("structural.repo-file-interface",
			"LayerModel.PortLayers is empty; repo-file interface enforcement requires at least one port sublayer",
			"declare PortLayers in your LayerModel (e.g. []string{\"core/repo\"}), or remove structural.NewRepoFileInterface() from your RuleSet")}
	}
	var violations []core.Violation
	for _, pkg := range ctx.Pkgs() {
		if portLayer := analysisutil.MatchPortSublayer(layers, pkg.PkgPath); portLayer != "" {
			violations = append(violations, r.checkPortPackage(ctx, pkg, portLayer)...)
			continue
		}
		violations = append(violations, r.checkInterfacePlacement(ctx, pkg)...)
	}
	return violations
}

func (r *RepoFileInterface) checkPortPackage(ctx *core.Context, pkg *packages.Package, portLayer string) []core.Violation {
	var violations []core.Violation
	portDisplay := layerDisplay(portLayer)
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
				"structural.repo-file-interface-missing",
				`file "`+base+`" in `+portDisplay+`/ must contain interface "`+expected+`"`,
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
			"structural.repo-file-extra-interface",
			`file "`+base+`" in `+portDisplay+`/ must define only "`+expected+`", found extra: `+strings.Join(extra, ", "),
			"move each extra interface to its own file (e.g. "+analysisutil.PascalToSnake(extra[0])+".go)",
		))
	}
	return violations
}

func (r *RepoFileInterface) checkInterfacePlacement(ctx *core.Context, pkg *packages.Package) []core.Violation {
	arch := ctx.Arch()
	sublayer, ok := domainSublayerForPackage(ctx.Module(), arch, pkg.PkgPath)
	if !ok || analysisutil.IsPortSublayer(arch.Layers, sublayer) {
		return nil
	}
	repoName := analysisutil.PortSublayerName(arch.Layers)
	repoDisplay := layerDisplay(repoName)
	currentDisplay := layerDisplay(sublayer)
	var violations []core.Violation
	for _, file := range pkg.Syntax {
		filePath := analysisutil.RelativePathForPackage(pkg, pkg.Fset.Position(file.Pos()).Filename)
		if ctx.IsExcluded(filePath) {
			continue
		}
		for _, info := range analysisutil.InspectTypeSpecs(file, pkg.Fset) {
			if info.IsInterface && r.isRepoPortName(info.Name) {
				violations = append(violations, r.violation(
					filePath,
					info.Line,
					"structural.interface-placement",
					`interface "`+info.Name+`" matches repository-port naming and must be defined in `+repoDisplay+`/, not in `+currentDisplay+`/`,
					"move to "+repoDisplay+"/, or rename if it's a consumer-defined interface",
				))
			}
			if info.AliasFrom != "" && isPortPackage(ctx.Module(), arch, info.AliasFrom) {
				violations = append(violations, r.violation(
					filePath,
					info.Line,
					"structural.interface-placement",
					`type alias "`+info.Name+`" re-exports interface from `+repoDisplay+`/ - repository-port contracts must stay in `+repoDisplay+`/`,
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

func isPortPackage(module string, arch core.Architecture, pkgPath string) bool {
	sublayer, ok := domainSublayerForPackage(module, arch, pkgPath)
	return ok && analysisutil.IsPortSublayer(arch.Layers, sublayer)
}

func domainSublayerForPackage(module string, arch core.Architecture, pkgPath string) (string, bool) {
	module = strings.TrimSuffix(module, "/")
	if module == "" || arch.Layout.InternalRoot == "" || arch.Layout.DomainDir == "" {
		return "", false
	}
	internalPrefix := module + "/" + arch.Layout.InternalRoot + "/"
	rel, ok := strings.CutPrefix(pkgPath, internalPrefix)
	if !ok {
		return "", false
	}
	domainPrefix := arch.Layout.DomainDir + "/"
	domainRel, ok := strings.CutPrefix(rel, domainPrefix)
	if !ok {
		return "", false
	}
	_, layerRel, ok := strings.Cut(domainRel, "/")
	if !ok || layerRel == "" {
		return "", false
	}
	if sublayer, ok := knownSublayerForRel(arch.Layers.Sublayers, layerRel); ok {
		return sublayer, true
	}
	return "", false
}

func knownSublayerForRel(sublayers []string, layerRel string) (string, bool) {
	var best string
	for _, sublayer := range sublayers {
		if layerRel == sublayer || strings.HasPrefix(layerRel, sublayer+"/") {
			if len(sublayer) > len(best) {
				best = sublayer
			}
		}
	}
	return best, best != ""
}

func layerDisplay(sublayer string) string {
	return strings.TrimSuffix(sublayer, "/")
}

var _ core.Rule = (*RepoFileInterface)(nil)
