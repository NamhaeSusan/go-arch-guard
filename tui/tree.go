package tui

import (
	"sort"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"golang.org/x/tools/go/packages"
)

// PkgNode holds metadata for a tree node.
type PkgNode struct {
	RelPath       string
	IsLeaf        bool
	Imports       []string
	FullPath      string
	HasViolations bool
}

// BuildTree creates a tview.TreeView from loaded packages.
func BuildTree(pkgs []*packages.Package, module string, violations ViolationIndex) *tview.TreeView {
	root := tview.NewTreeNode("📦 " + module).SetColor(tcell.ColorWhite)
	root.SetReference(&PkgNode{RelPath: "", IsLeaf: false})

	// Collect relative paths and build import map.
	type pkgInfo struct {
		relPath  string
		imports  []string
		fullPath string
	}
	var infos []pkgInfo
	for _, pkg := range pkgs {
		rel := strings.TrimPrefix(pkg.PkgPath, module+"/")
		if rel == pkg.PkgPath {
			continue
		}
		var imports []string
		for impPath := range pkg.Imports {
			imports = append(imports, impPath)
		}
		sort.Strings(imports)
		infos = append(infos, pkgInfo{relPath: rel, imports: imports, fullPath: pkg.PkgPath})
	}
	sort.Slice(infos, func(i, j int) bool { return infos[i].relPath < infos[j].relPath })

	// Build tree nodes from paths.
	nodeMap := map[string]*tview.TreeNode{"": root}
	for _, info := range infos {
		parts := strings.Split(info.relPath, "/")
		for depth := 1; depth <= len(parts); depth++ {
			key := strings.Join(parts[:depth], "/")
			if _, exists := nodeMap[key]; exists {
				continue
			}
			parentKey := ""
			if depth > 1 {
				parentKey = strings.Join(parts[:depth-1], "/")
			}
			parent := nodeMap[parentKey]

			isLeaf := depth == len(parts)
			name := parts[depth-1]
			sev := violations.Severity(key)
			color, prefix := severityStyle(sev)
			node := tview.NewTreeNode(prefix + name).
				SetColor(color).
				SetSelectable(true).
				SetExpanded(depth <= 2)
			pn := &PkgNode{
				RelPath:       key,
				IsLeaf:        isLeaf,
				HasViolations: sev != sevNone,
			}
			if isLeaf {
				pn.Imports = info.imports
				pn.FullPath = info.fullPath
			}
			node.SetReference(pn)

			parent.AddChild(node)
			nodeMap[key] = node
		}
	}

	tree := tview.NewTreeView().
		SetRoot(root).
		SetCurrentNode(root)

	return tree
}

func severityStyle(sev severity) (tcell.Color, string) {
	switch sev {
	case sevError:
		return tcell.ColorRed, "✗ "
	case sevWarning:
		return tcell.ColorYellow, "⚠ "
	default:
		return tcell.ColorGreen, ""
	}
}

// BuildImportedByMap creates a reverse lookup: package path → list of packages that import it.
func BuildImportedByMap(pkgs []*packages.Package) map[string][]string {
	importedBy := make(map[string][]string)
	for _, pkg := range pkgs {
		for impPath := range pkg.Imports {
			importedBy[impPath] = append(importedBy[impPath], pkg.PkgPath)
		}
	}
	for k := range importedBy {
		sort.Strings(importedBy[k])
	}
	return importedBy
}
