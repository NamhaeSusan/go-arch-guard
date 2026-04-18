package rules

import (
	"fmt"
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/packages"
)

// checkTypeInSignature walks all FuncDecl params and results in internal
// packages and emits a violation when a base type (after stripping
// pointer/slice/array/map/chan wrappers) matches one of typeNames and
// the function's package layer is outside allowedLayers.
//
// message/fix are fmt templates formatted with (typeID string, allowedLayers []string).
func checkTypeInSignature(
	pkgs []*packages.Package,
	projectModule, projectRoot string,
	m Model,
	cfg Config,
	typeNames []string,
	allowedLayers []string,
	ruleName, message, fix string,
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
	ruleName, message, fix string,
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
			File:     relFile,
			Line:     pos.Line,
			Rule:     ruleName,
			Message:  fmt.Sprintf(message, id, allowedLayers),
			Fix:      fmt.Sprintf(fix, id, allowedLayers),
			Severity: cfg.Sev,
		})
	}
}

func stripWrappers(t types.Type) types.Type {
	for {
		switch x := t.(type) {
		case *types.Pointer:
			t = x.Elem()
		case *types.Slice:
			t = x.Elem()
		case *types.Array:
			t = x.Elem()
		case *types.Map:
			t = x.Elem()
		case *types.Chan:
			t = x.Elem()
		default:
			return t
		}
	}
}

func namedQualifiedName(t types.Type) string {
	named, ok := t.(*types.Named)
	if !ok {
		return ""
	}
	obj := named.Obj()
	if obj == nil || obj.Pkg() == nil {
		return ""
	}
	return obj.Pkg().Path() + "." + obj.Name()
}
