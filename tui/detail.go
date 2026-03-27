package tui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// DetailPanel renders import information for a selected package.
type DetailPanel struct {
	view       *tview.TextView
	importedBy map[string][]string
	module     string
}

// NewDetailPanel creates a detail panel with the given reverse-import map.
func NewDetailPanel(importedBy map[string][]string, module string) *DetailPanel {
	view := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true)
	view.SetBorder(true).
		SetTitle(" Dependencies ").
		SetBorderColor(tcell.ColorGray)

	return &DetailPanel{view: view, importedBy: importedBy, module: module}
}

// View returns the underlying tview.TextView.
func (d *DetailPanel) View() *tview.TextView {
	return d.view
}

// Update refreshes the detail panel for the selected package node.
func (d *DetailPanel) Update(node *PkgNode) {
	d.view.Clear()
	if node == nil || !node.IsLeaf {
		d.view.SetText("[gray]Select a package to view dependencies")
		return
	}

	var b strings.Builder

	fmt.Fprintf(&b, "[white::b]%s\n\n", node.RelPath)

	// Imports section.
	b.WriteString("[dodgerblue::b]Imports:\n")
	if len(node.Imports) == 0 {
		b.WriteString("[gray]  (none)\n")
	}
	for _, imp := range node.Imports {
		color := "gray"
		display := imp
		if rel, ok := strings.CutPrefix(imp, d.module+"/"); ok {
			color = "green"
			display = rel
		}
		fmt.Fprintf(&b, "[%s]  • %s\n", color, display)
	}

	// Imported by section.
	b.WriteString("\n[yellow::b]Imported by:\n")
	refs := d.importedBy[node.FullPath]
	if len(refs) == 0 {
		b.WriteString("[gray]  (none)\n")
	}
	for _, ref := range refs {
		display := ref
		if rel, ok := strings.CutPrefix(ref, d.module+"/"); ok {
			display = rel
		}
		fmt.Fprintf(&b, "[green]  • %s\n", display)
	}

	d.view.SetText(b.String())
}
