package tui

import (
	"fmt"
	"strings"

	"github.com/NamhaeSusan/go-arch-guard/rules"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"golang.org/x/tools/go/packages"
)

// Run launches the TUI application for the given packages.
func Run(pkgs []*packages.Package, module, root string) error {
	violations := BuildViolationIndex(pkgs, module, root)
	importedBy := BuildImportedByMap(pkgs)
	metrics := BuildMetricsIndex(pkgs, module)
	tree := BuildTree(pkgs, module, violations)
	detail := NewDetailPanel(importedBy, violations, metrics, module)

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
		SetTitle(" Package Tree (/ search, q quit) ").
		SetBorderColor(tcell.ColorGray)

	// Search input.
	searchInput := tview.NewInputField().
		SetLabel(" / ").
		SetFieldWidth(0).
		SetFieldBackgroundColor(tcell.ColorBlack)

	// Status bar.
	errCount, warnCount := 0, 0
	for _, viols := range violations {
		for _, v := range viols {
			if v.Severity == rules.Error {
				errCount++
			} else {
				warnCount++
			}
		}
	}
	status := tview.NewTextView().SetDynamicColors(true)
	status.SetText(fmt.Sprintf(" [green]%d[white] pkgs  [red]%d[white] errors  [yellow]%d[white] warnings  [gray]│  / search  Tab panel  q quit  ↑↓ navigate",
		len(pkgs), errCount, warnCount))

	// Layout.
	mainFlex := tview.NewFlex().
		AddItem(tree, 0, 1, true).
		AddItem(detail.View(), 0, 1, false)

	searchRow := tview.NewFlex().
		AddItem(searchInput, 0, 1, false)

	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(status, 1, 0, false).
		AddItem(mainFlex, 0, 1, true).
		AddItem(searchRow, 0, 0, false)

	app := tview.NewApplication().SetRoot(layout, true)
	searchVisible := false

	filterTree := func(query string) {
		r := tree.GetRoot()
		if r == nil {
			return
		}
		filterNode(r, strings.ToLower(query))
		app.Draw()
	}

	searchInput.SetChangedFunc(func(text string) {
		filterTree(text)
	})

	searchInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape || key == tcell.KeyEnter {
			searchVisible = false
			layout.ResizeItem(searchRow, 0, 0)
			app.SetFocus(tree)
			if key == tcell.KeyEscape {
				searchInput.SetText("")
				filterTree("")
			}
		}
	})

	detailFocused := false
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if searchVisible {
			return event
		}
		if event.Key() == tcell.KeyTab {
			detailFocused = !detailFocused
			if detailFocused {
				app.SetFocus(detail.View())
			} else {
				app.SetFocus(tree)
			}
			return nil
		}
		switch event.Rune() {
		case 'q':
			app.Stop()
			return nil
		case '/':
			searchVisible = true
			layout.ResizeItem(searchRow, 1, 0)
			app.SetFocus(searchInput)
			return nil
		}
		return event
	})

	return app.Run()
}

func filterNode(node *tview.TreeNode, query string) bool {
	if query == "" {
		node.SetExpanded(true)
		for _, child := range node.GetChildren() {
			filterNode(child, query)
		}
		return true
	}

	ref, ok := node.GetReference().(*PkgNode)
	name := ""
	if ok {
		name = strings.ToLower(ref.RelPath)
	}

	selfMatch := strings.Contains(name, query)
	childMatch := false
	for _, child := range node.GetChildren() {
		if filterNode(child, query) {
			childMatch = true
		}
	}

	visible := selfMatch || childMatch
	node.SetExpanded(visible)
	return visible
}
