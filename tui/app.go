package tui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"golang.org/x/tools/go/packages"
)

// Run launches the TUI application for the given packages.
func Run(pkgs []*packages.Package, module string) error {
	importedBy := BuildImportedByMap(pkgs)
	tree := BuildTree(pkgs, module)
	detail := NewDetailPanel(importedBy, module)

	tree.SetSelectedFunc(func(node *tview.TreeNode) {
		node.SetExpanded(!node.IsExpanded())
	})

	tree.SetChangedFunc(func(node *tview.TreeNode) {
		ref, ok := node.GetReference().(*PkgNode)
		if !ok {
			return
		}
		detail.Update(ref)
	})

	tree.SetBorder(true).
		SetTitle(" Package Tree ").
		SetBorderColor(tcell.ColorGray)

	flex := tview.NewFlex().
		AddItem(tree, 0, 1, true).
		AddItem(detail.View(), 0, 1, false)

	app := tview.NewApplication().SetRoot(flex, true)
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'q' {
			app.Stop()
			return nil
		}
		return event
	})

	return app.Run()
}
