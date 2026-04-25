package interfaces

import (
	"fmt"
	"go/ast"
	"go/types"
	"sort"
	"strings"

	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/core/analysisutil"
	"golang.org/x/tools/go/packages"
)

type Pattern struct {
	cfg ruleConfig
}

func NewPattern(opts ...Option) *Pattern {
	return &Pattern{cfg: newConfig(opts, core.Error)}
}

func (r *Pattern) Spec() core.RuleSpec {
	return core.RuleSpec{
		ID:              "interfaces.pattern",
		Description:     "enforce interface package and constructor conventions",
		DefaultSeverity: r.cfg.severity,
		Violations: []core.ViolationSpec{
			{ID: "interface.constructor-name", DefaultSeverity: r.cfg.severity},
			{ID: "interface.constructor-returns-interface", DefaultSeverity: r.cfg.severity},
			{ID: "interface.exported-impl", DefaultSeverity: r.cfg.severity},
			{ID: "interface.too-many-methods", DefaultSeverity: r.cfg.severity},
			{ID: "interface.single-per-package", DefaultSeverity: r.cfg.severity},
		},
	}
}

func (r *Pattern) Check(ctx *core.Context) []core.Violation {
	if ctx == nil {
		return nil
	}
	pkgs := ctx.Pkgs()
	projectModule := analysisutil.ResolveModuleFromContext(ctx, "")
	arch := ctx.Arch()
	if !hasInternalPackages(pkgs, projectModule, arch.Layout.InternalRoot) {
		return []core.Violation{metaLayoutNotSupported("interfaces.pattern", projectModule)}
	}

	var violations []core.Violation
	for _, pkg := range pkgs {
		if isExcludedInterfacePatternPkg(arch, pkg) {
			continue
		}

		violations = append(violations, r.checkExportedImpl(pkg)...)

		ifaces := collectExportedInterfacesFromPkg(pkg)
		if len(ifaces) == 0 {
			continue
		}

		violations = append(violations, r.checkSingleInterfacePerPackage(pkg, ifaces)...)
		violations = append(violations, r.checkTooManyMethods(pkg, ifaces)...)
		violations = append(violations, r.checkConstructorName(pkg)...)
		violations = append(violations, r.checkConstructorReturnsInterface(pkg, ifaces)...)
	}
	return violations
}

func isExcludedInterfacePatternPkg(arch core.Architecture, pkg *packages.Package) bool {
	parts := strings.Split(pkg.PkgPath, "/")
	internalIdx := -1
	for i, p := range parts {
		if p == arch.Layout.InternalRoot {
			internalIdx = i
			break
		}
	}
	if internalIdx < 0 || internalIdx >= len(parts)-1 {
		return true
	}

	after := parts[internalIdx+1:]
	if arch.Layout.SharedDir != "" && after[0] == arch.Layout.SharedDir {
		return true
	}

	exclude := arch.Structure.InterfacePatternExclude
	if arch.Layout.DomainDir == "" {
		return exclude[after[0]]
	}

	if after[0] != arch.Layout.DomainDir || len(after) < 3 {
		return true
	}
	sublayer := after[2]
	if exclude[sublayer] {
		return true
	}
	if len(after) >= 4 {
		return exclude[sublayer+"/"+after[3]]
	}
	return false
}

func collectExportedInterfacesFromPkg(pkg *packages.Package) map[string]*ast.InterfaceType {
	result := make(map[string]*ast.InterfaceType)
	for _, file := range pkg.Syntax {
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
				if iface, ok := ts.Type.(*ast.InterfaceType); ok {
					result[ts.Name.Name] = iface
				}
			}
		}
	}
	return result
}

func (r *Pattern) checkSingleInterfacePerPackage(pkg *packages.Package, ifaces map[string]*ast.InterfaceType) []core.Violation {
	if len(ifaces) <= 1 {
		return nil
	}
	names := make([]string, 0, len(ifaces))
	for name := range ifaces {
		names = append(names, name)
	}
	sort.Strings(names)
	return []core.Violation{r.violation(pkg, 0, "interface.single-per-package",
		fmt.Sprintf("package has %d exported interfaces (%s), expected at most 1", len(ifaces), strings.Join(names, ", ")),
		"split into separate packages, one interface each")}
}

func (r *Pattern) checkTooManyMethods(pkg *packages.Package, ifaces map[string]*ast.InterfaceType) []core.Violation {
	if r.cfg.maxMethods <= 0 {
		return nil
	}

	names := make([]string, 0, len(ifaces))
	for name := range ifaces {
		names = append(names, name)
	}
	sort.Strings(names)

	var violations []core.Violation
	for _, name := range names {
		iface := ifaces[name]
		count := iface.Methods.NumFields()
		if count <= r.cfg.maxMethods {
			continue
		}
		pos := pkg.Fset.Position(iface.Pos())
		violations = append(violations, core.Violation{
			File:              analysisutil.RelativePathForPackage(pkg, pos.Filename),
			Line:              pos.Line,
			Rule:              "interface.too-many-methods",
			Message:           fmt.Sprintf("interface %q has %d methods, expected at most %d", name, count, r.cfg.maxMethods),
			Fix:               "split the interface by consumer needs",
			DefaultSeverity:   r.cfg.severity,
			EffectiveSeverity: r.cfg.severity,
		})
	}
	return violations
}

func (r *Pattern) checkExportedImpl(pkg *packages.Package) []core.Violation {
	if pkg.Types == nil {
		return nil
	}
	structs := collectExportedStructs(pkg)
	if len(structs) == 0 {
		return nil
	}

	scope := pkg.Types.Scope()
	typedIfaces := make(map[string]*types.Interface)
	for _, name := range scope.Names() {
		if !ast.IsExported(name) {
			continue
		}
		if iface := lookupInterface(scope, name); iface != nil && iface.NumMethods() > 0 {
			typedIfaces[name] = iface
		}
	}
	if len(typedIfaces) == 0 {
		return nil
	}

	structNames := make([]string, 0, len(structs))
	for name := range structs {
		structNames = append(structNames, name)
	}
	sort.Strings(structNames)
	ifaceNames := make([]string, 0, len(typedIfaces))
	for name := range typedIfaces {
		ifaceNames = append(ifaceNames, name)
	}
	sort.Strings(ifaceNames)

	var violations []core.Violation
	for _, structName := range structNames {
		obj := scope.Lookup(structName)
		if obj == nil {
			continue
		}
		named, ok := types.Unalias(obj.Type()).(*types.Named)
		if !ok {
			continue
		}
		ptrType := types.NewPointer(named)
		for _, ifaceName := range ifaceNames {
			iface := typedIfaces[ifaceName]
			if !types.Implements(named, iface) && !types.Implements(ptrType, iface) {
				continue
			}
			violations = append(violations, r.violation(pkg, 0, "interface.exported-impl",
				fmt.Sprintf("type %q is exported but implements interface %q; make it unexported", structName, ifaceName),
				fmt.Sprintf("rename to %q", strings.ToLower(structName[:1])+structName[1:])))
		}
	}
	return violations
}

func collectExportedStructs(pkg *packages.Package) map[string]bool {
	result := make(map[string]bool)
	for _, file := range pkg.Syntax {
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
				if _, ok := ts.Type.(*ast.StructType); ok {
					result[ts.Name.Name] = true
				}
			}
		}
	}
	return result
}

func lookupInterface(scope *types.Scope, name string) *types.Interface {
	obj := scope.Lookup(name)
	if obj == nil {
		return nil
	}
	t := types.Unalias(obj.Type())
	if iface, ok := t.Underlying().(*types.Interface); ok {
		return iface
	}
	return nil
}

func (r *Pattern) checkConstructorName(pkg *packages.Package) []core.Violation {
	var violations []core.Violation
	for _, file := range pkg.Syntax {
		for _, decl := range file.Decls {
			fd, ok := decl.(*ast.FuncDecl)
			if !ok || fd.Recv != nil || !fd.Name.IsExported() {
				continue
			}
			name := fd.Name.Name
			if strings.HasPrefix(name, "New") && name != "New" {
				violations = append(violations, r.violation(pkg, 0, "interface.constructor-name",
					fmt.Sprintf("constructor %q must be named \"New\"; NewXxx variants are not allowed", name),
					"rename to \"New\""))
			}
		}
	}
	return violations
}

func (r *Pattern) checkConstructorReturnsInterface(pkg *packages.Package, ifaces map[string]*ast.InterfaceType) []core.Violation {
	var violations []core.Violation
	for _, file := range pkg.Syntax {
		for _, decl := range file.Decls {
			fd, ok := decl.(*ast.FuncDecl)
			if !ok || fd.Recv != nil || fd.Name.Name != "New" {
				continue
			}
			if fd.Type.Results == nil || len(fd.Type.Results.List) == 0 {
				continue
			}

			firstRet := fd.Type.Results.List[0].Type
			if ident, ok := firstRet.(*ast.Ident); ok && ifaces[ident.Name] != nil {
				continue
			}

			fix := "return an interface type"
			if len(ifaces) == 1 {
				for ifaceName := range ifaces {
					fix = fmt.Sprintf("return %s instead", ifaceName)
				}
			}
			violations = append(violations, r.violation(pkg, 0, "interface.constructor-returns-interface",
				fmt.Sprintf("New() returns %s, should return an interface", formatTypeExpr(firstRet)),
				fix))
		}
	}
	return violations
}

func formatTypeExpr(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.StarExpr:
		return "*" + formatTypeExpr(e.X)
	case *ast.SelectorExpr:
		return formatTypeExpr(e.X) + "." + e.Sel.Name
	case *ast.ArrayType:
		return "[]" + formatTypeExpr(e.Elt)
	case *ast.MapType:
		return "map[" + formatTypeExpr(e.Key) + "]" + formatTypeExpr(e.Value)
	case *ast.ChanType:
		switch e.Dir {
		case ast.SEND:
			return "chan<- " + formatTypeExpr(e.Value)
		case ast.RECV:
			return "<-chan " + formatTypeExpr(e.Value)
		default:
			return "chan " + formatTypeExpr(e.Value)
		}
	case *ast.FuncType:
		return "func" + formatFieldList(e.Params) + formatFuncResults(e.Results)
	case *ast.IndexExpr:
		return formatTypeExpr(e.X) + "[" + formatTypeExpr(e.Index) + "]"
	case *ast.IndexListExpr:
		parts := make([]string, 0, len(e.Indices))
		for _, idx := range e.Indices {
			parts = append(parts, formatTypeExpr(idx))
		}
		return formatTypeExpr(e.X) + "[" + strings.Join(parts, ", ") + "]"
	default:
		return "unknown"
	}
}

func formatFieldList(fields *ast.FieldList) string {
	if fields == nil {
		return "()"
	}
	parts := make([]string, 0, len(fields.List))
	for _, f := range fields.List {
		parts = append(parts, formatTypeExpr(f.Type))
	}
	return "(" + strings.Join(parts, ", ") + ")"
}

func formatFuncResults(fields *ast.FieldList) string {
	if fields == nil || len(fields.List) == 0 {
		return ""
	}
	if len(fields.List) == 1 && len(fields.List[0].Names) == 0 {
		return " " + formatTypeExpr(fields.List[0].Type)
	}
	return " " + formatFieldList(fields)
}

func (r *Pattern) violation(pkg *packages.Package, line int, id, message, fix string) core.Violation {
	return core.Violation{
		File:              packageFile(pkg),
		Line:              line,
		Rule:              id,
		Message:           message,
		Fix:               fix,
		DefaultSeverity:   r.cfg.severity,
		EffectiveSeverity: r.cfg.severity,
	}
}

func packageFile(pkg *packages.Package) string {
	if len(pkg.GoFiles) > 0 {
		return analysisutil.RelativePathForPackage(pkg, pkg.GoFiles[0])
	}
	return pkg.PkgPath
}

var _ core.Rule = (*Pattern)(nil)
