package rules

import (
	"go/ast"
	"go/types"
	"strings"

	"github.com/NamhaeSusan/go-arch-guard/core/analysisutil"
	"golang.org/x/tools/go/packages"
)

// checkTypeInSignature walks all FuncDecl params and results in internal
// packages and emits a violation when a base type (after stripping
// pointer/slice/array/map/chan wrappers) matches one of typeNames and
// the function's package layer is outside allowedLayers.
//
// message/fix are typed callbacks receiving (typeID, allowedLayers).
func checkTypeInSignature(
	pkgs []*packages.Package,
	projectModule, projectRoot string,
	m Model,
	cfg Config,
	typeNames []string,
	allowedLayers []string,
	ruleName string,
	message, fix func(typeID string, allowed []string) string,
) []Violation {
	if len(typeNames) == 0 {
		return nil
	}
	projectModule = resolveModule(pkgs, projectModule)
	projectRoot = resolveRoot(pkgs, projectRoot)
	internalPrefix := projectModule + "/internal/"

	wanted := map[string]bool{}
	for _, n := range typeNames {
		wanted[n] = true
	}
	allowed := map[string]bool{}
	for _, l := range allowedLayers {
		allowed[l] = true
	}

	var violations []Violation
	for _, pkg := range pkgs {
		if isExcludedPackage(cfg, pkg.PkgPath, projectModule) {
			continue
		}
		if !strings.HasPrefix(pkg.PkgPath, internalPrefix) {
			continue
		}
		layer := layerOfPackage(m, pkg.PkgPath, internalPrefix)
		if layer == "" || allowed[layer] {
			continue
		}
		if pkg.TypesInfo == nil {
			continue
		}
		for _, file := range pkg.Syntax {
			for _, decl := range file.Decls {
				fd, ok := decl.(*ast.FuncDecl)
				if !ok || fd.Type == nil {
					continue
				}
				checkFieldList(pkg, fd.Type.Params, wanted, projectRoot, cfg, ruleName, message, fix, allowedLayers, &violations)
				checkFieldList(pkg, fd.Type.Results, wanted, projectRoot, cfg, ruleName, message, fix, allowedLayers, &violations)
			}
		}
	}
	return violations
}

func checkFieldList(
	pkg *packages.Package,
	fields *ast.FieldList,
	wanted map[string]bool,
	projectRoot string,
	cfg Config,
	ruleName string,
	message, fix func(typeID string, allowed []string) string,
	allowedLayers []string,
	out *[]Violation,
) {
	if fields == nil {
		return
	}
	for _, f := range fields.List {
		t := pkg.TypesInfo.TypeOf(f.Type)
		if t == nil {
			continue
		}
		id := namedQualifiedName(stripWrappers(t))
		if id == "" || !wanted[id] {
			continue
		}
		pos := pkg.Fset.Position(f.Pos())
		relFile := relPathFromRoot(projectRoot, pos.Filename)
		*out = append(*out, Violation{
			File:              relFile,
			Line:              pos.Line,
			Rule:              ruleName,
			Message:           message(id, allowedLayers),
			Fix:               fix(id, allowedLayers),
			DefaultSeverity:   cfg.Sev,
			EffectiveSeverity: cfg.Sev,
		})
	}
}

func stripWrappers(t types.Type) types.Type {
	return analysisutil.StripWrappers(t)
}

func namedQualifiedName(t types.Type) string {
	return analysisutil.NamedQualifiedName(t)
}
