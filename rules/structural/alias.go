package structural

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/core/analysisutil"
)

const (
	ruleAlias           = "structural.alias"
	aliasExists         = "structure.domain-alias-exists"
	aliasPackage        = "structure.domain-alias-package"
	aliasExclusive      = "structure.domain-alias-exclusive"
	aliasNoInterface    = "structure.domain-alias-no-interface"
	aliasContractExport = "structure.domain-alias-contract-reexport"
)

type Option func(any)

func WithSeverity(s core.Severity) Option {
	return func(rule any) {
		switch r := rule.(type) {
		case *Alias:
			r.severity = s
		case *Placement:
			r.severity = s
		case *BannedPackage:
			r.severity = s
		case *ModelRequired:
			r.severity = s
		case *InternalTopLevel:
			r.severity = s
		}
	}
}

type Alias struct {
	severity core.Severity
}

func NewAlias(opts ...Option) *Alias {
	r := &Alias{severity: core.Error}
	applyOptions(r, opts)
	return r
}

func (r *Alias) Spec() core.RuleSpec {
	return withSeverity(core.RuleSpec{
		ID:              ruleAlias,
		Description:     "domain roots must expose only an alias file and avoid contract re-exports",
		DefaultSeverity: r.severity,
		Violations: []core.ViolationSpec{
			{ID: aliasExists, Description: "domain root is missing its alias file"},
			{ID: aliasPackage, Description: "alias file package name does not match the domain root"},
			{ID: aliasExclusive, Description: "domain root contains public Go files outside the alias file"},
			{ID: aliasNoInterface, Description: "alias file declares or re-exports an interface"},
			{ID: aliasContractExport, Description: "alias file re-exports a contract sublayer type"},
		},
	}, r.severity)
}

func (r *Alias) Check(ctx *core.Context) []core.Violation {
	if ctx == nil {
		return nil
	}
	arch := ctx.Arch()
	if arch.Layout.DomainDir == "" || !arch.Structure.RequireAlias {
		return nil
	}

	domainDir := filepath.Join(ctx.Root(), "internal", filepath.FromSlash(arch.Layout.DomainDir))
	entries, err := os.ReadDir(domainDir)
	if err != nil {
		return nil
	}

	var violations []core.Violation
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		domainName := entry.Name()
		relPath := filepath.ToSlash(filepath.Join("internal", arch.Layout.DomainDir, domainName))
		if ctx.IsExcluded(relPath + "/") {
			continue
		}
		rootDir := filepath.Join(domainDir, domainName)
		aliasName := aliasFileName(arch)
		aliasPath := filepath.Join(rootDir, aliasName)
		aliasRel := relPath + "/" + aliasName

		if _, err := os.Stat(aliasPath); err != nil {
			violations = append(violations, violation(r.severity, aliasExists, relPath+"/",
				`domain root "`+domainName+`" must define `+aliasName,
				"add "+aliasName+" as the single public surface file for the domain root package"))
			continue
		}

		violations = append(violations, r.checkAliasPackage(aliasPath, aliasRel, aliasName, domainName)...)
		violations = append(violations, r.checkAliasOnly(ctx, rootDir, relPath, aliasName, domainName)...)
		violations = append(violations, r.checkAliasTypes(aliasPath, aliasRel, aliasName, arch)...)
	}
	return violations
}

func (r *Alias) checkAliasPackage(aliasPath, aliasRel, aliasName, domainName string) []core.Violation {
	file, err := parser.ParseFile(token.NewFileSet(), aliasPath, nil, parser.PackageClauseOnly)
	if err != nil || file.Name.Name == domainName {
		return nil
	}
	return []core.Violation{violation(r.severity, aliasPackage, aliasRel,
		aliasName+` package name must match domain root "`+domainName+`"`,
		`set "package `+domainName+`" in `+aliasName)}
}

func (r *Alias) checkAliasOnly(ctx *core.Context, rootDir, relPath, aliasName, domainName string) []core.Violation {
	entries, err := os.ReadDir(rootDir)
	if err != nil {
		return nil
	}
	var violations []core.Violation
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") || name == aliasName {
			continue
		}
		rel := relPath + "/" + name
		if ctx.IsExcluded(rel) {
			continue
		}
		violations = append(violations, violation(r.severity, aliasExclusive, rel,
			`domain root "`+domainName+`" must expose its public API from `+aliasName+` only`,
			`move "`+name+`" into a sub-package or merge the public API into `+aliasName))
	}
	return violations
}

func (r *Alias) checkAliasTypes(aliasPath, aliasRel, aliasName string, arch core.Architecture) []core.Violation {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, aliasPath, nil, 0)
	if err != nil {
		return nil
	}
	var violations []core.Violation
	for _, info := range inspectTypeSpecs(file, fset) {
		if info.isInterface {
			v := violation(r.severity, aliasNoInterface, aliasRel,
				aliasName+` re-exports interface "`+info.name+`" - suspected cross-domain dependency; use `+arch.Layout.OrchestrationDir+`/ instead`,
				"move cross-domain coordination to "+arch.Layout.OrchestrationDir+"/handler/ or "+arch.Layout.OrchestrationDir+"/")
			v.Line = info.line
			violations = append(violations, v)
		}
		if src := analysisutil.MatchContractSublayer(arch.Layers, info.aliasFrom); src != "" {
			v := violation(r.severity, aliasContractExport, aliasRel,
				aliasName+` re-exports "`+info.name+`" from `+src+` - suspected cross-domain dependency; use `+arch.Layout.OrchestrationDir+`/ instead`,
				"move cross-domain coordination to "+arch.Layout.OrchestrationDir+"/handler/ or "+arch.Layout.OrchestrationDir+"/")
			v.Line = info.line
			violations = append(violations, v)
		}
	}
	return violations
}

type typeSpecInfo struct {
	name        string
	line        int
	isInterface bool
	aliasFrom   string
}

func inspectTypeSpecs(file *ast.File, fset *token.FileSet) []typeSpecInfo {
	var result []typeSpecInfo
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range gen.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			info := typeSpecInfo{
				name: typeSpec.Name.Name,
				line: fset.Position(typeSpec.Name.Pos()).Line,
			}
			if _, ok := typeSpec.Type.(*ast.InterfaceType); ok {
				info.isInterface = true
			}
			if typeSpec.Assign != 0 {
				if sel, ok := typeSpec.Type.(*ast.SelectorExpr); ok {
					if ident, ok := sel.X.(*ast.Ident); ok {
						info.aliasFrom = analysisutil.ResolveIdentImportPath(file, ident.Name)
					}
				}
			}
			if info.isInterface || info.aliasFrom != "" {
				result = append(result, info)
			}
		}
	}
	return result
}

func applyOptions(rule any, opts []Option) {
	for _, opt := range opts {
		opt(rule)
	}
}

func withSeverity(spec core.RuleSpec, severity core.Severity) core.RuleSpec {
	spec.DefaultSeverity = severity
	for i := range spec.Violations {
		spec.Violations[i].DefaultSeverity = severity
	}
	return spec
}

func violation(severity core.Severity, rule, file, message, fix string) core.Violation {
	return core.Violation{
		File:              file,
		Rule:              rule,
		Message:           message,
		Fix:               fix,
		DefaultSeverity:   severity,
		EffectiveSeverity: severity,
	}
}

func aliasFileName(arch core.Architecture) string {
	if arch.Naming.AliasFileName != "" {
		return arch.Naming.AliasFileName
	}
	return "alias.go"
}

func hasNonTestGoFiles(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go") {
			return true
		}
	}
	return false
}
