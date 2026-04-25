package interfaces

import (
	"fmt"
	"go/ast"
	"go/token"
	"sort"
	"strings"

	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/core/analysisutil"
	"golang.org/x/tools/go/packages"
)

type Container struct {
	cfg config
}

func NewContainer(opts ...Option) *Container {
	cfg := config{severity: core.Warning}
	for _, opt := range opts {
		opt(&cfg)
	}
	return &Container{cfg: cfg}
}

func (r *Container) Spec() core.RuleSpec {
	return core.RuleSpec{
		ID:              "interfaces.container",
		Description:     "detect interfaces used only as struct field containers",
		DefaultSeverity: r.cfg.severity,
		Violations: []core.ViolationSpec{
			{ID: "interface.container-only", DefaultSeverity: r.cfg.severity},
		},
	}
}

func (r *Container) Check(ctx *core.Context) []core.Violation {
	if ctx == nil {
		return nil
	}

	var violations []core.Violation
	for _, pkg := range ctx.Pkgs() {
		violations = append(violations, r.checkPackage(pkg)...)
	}
	return violations
}

func (r *Container) checkPackage(pkg *packages.Package) []core.Violation {
	ifaces := collectNonTestInterfaces(pkg)
	if len(ifaces) == 0 {
		return nil
	}

	usage := make(map[string]*usageCounts, len(ifaces))
	for name := range ifaces {
		usage[name] = &usageCounts{}
	}

	for _, file := range pkg.Syntax {
		if isTestFile(pkg, file) {
			continue
		}
		ast.Inspect(file, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.StructType:
				if node.Fields == nil {
					return true
				}
				for _, f := range node.Fields.List {
					if len(f.Names) == 0 {
						continue
					}
					countTypeRefs(f.Type, ifaces, usage, fieldUsage)
				}
			case *ast.FuncDecl:
				if node.Type == nil {
					return true
				}
				countFieldListRefs(node.Type.Params, ifaces, usage, paramUsage)
				countFieldListRefs(node.Type.Results, ifaces, usage, returnUsage)
			}
			return true
		})
	}

	flagged := make([]string, 0)
	for name, c := range usage {
		if c.field > 0 && c.param == 0 && c.ret == 0 {
			flagged = append(flagged, name)
		}
	}
	sort.Strings(flagged)

	violations := make([]core.Violation, 0, len(flagged))
	for _, name := range flagged {
		pos := ifaces[name]
		file := analysisutil.RelativePathForPackage(pkg, pos.Filename)
		violations = append(violations, core.Violation{
			File:              file,
			Line:              pos.Line,
			Rule:              "interface.container-only",
			Message:           fmt.Sprintf("interface %q is only used as a struct field type, never as a function parameter or return type", name),
			Fix:               "use the concrete type or pass this interface as a constructor parameter",
			DefaultSeverity:   r.cfg.severity,
			EffectiveSeverity: r.cfg.severity,
		})
	}
	return violations
}

type usageCounts struct {
	field int
	param int
	ret   int
}

type usageKind int

const (
	fieldUsage usageKind = iota
	paramUsage
	returnUsage
)

func countFieldListRefs(fields *ast.FieldList, ifaces map[string]token.Position, usage map[string]*usageCounts, kind usageKind) {
	if fields == nil {
		return
	}
	for _, f := range fields.List {
		countTypeRefs(f.Type, ifaces, usage, kind)
	}
}

func countTypeRefs(expr ast.Expr, ifaces map[string]token.Position, usage map[string]*usageCounts, kind usageKind) {
	switch e := expr.(type) {
	case *ast.Ident:
		c, ok := usage[e.Name]
		if !ok {
			return
		}
		switch kind {
		case fieldUsage:
			c.field++
		case paramUsage:
			c.param++
		case returnUsage:
			c.ret++
		}
	case *ast.StarExpr:
		countTypeRefs(e.X, ifaces, usage, kind)
	case *ast.ArrayType:
		countTypeRefs(e.Elt, ifaces, usage, kind)
	case *ast.MapType:
		countTypeRefs(e.Key, ifaces, usage, kind)
		countTypeRefs(e.Value, ifaces, usage, kind)
	case *ast.ChanType:
		countTypeRefs(e.Value, ifaces, usage, kind)
	case *ast.FuncType:
		countFieldListRefs(e.Params, ifaces, usage, kind)
		countFieldListRefs(e.Results, ifaces, usage, kind)
	}
}

func collectNonTestInterfaces(pkg *packages.Package) map[string]token.Position {
	result := make(map[string]token.Position)
	for _, file := range pkg.Syntax {
		if isTestFile(pkg, file) {
			continue
		}
		for _, decl := range file.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			for _, spec := range gd.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok || ts.Assign != 0 {
					continue
				}
				if _, ok := ts.Type.(*ast.InterfaceType); ok {
					result[ts.Name.Name] = pkg.Fset.Position(ts.Name.Pos())
				}
			}
		}
	}
	return result
}

func isTestFile(pkg *packages.Package, file *ast.File) bool {
	pos := pkg.Fset.Position(file.Pos())
	return strings.HasSuffix(pos.Filename, "_test.go")
}

var _ core.Rule = (*Container)(nil)
