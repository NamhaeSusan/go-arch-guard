package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/NamhaeSusan/go-arch-guard/rules"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// DetailPanel renders import information for a selected package.
type DetailPanel struct {
	view       *tview.TextView
	importedBy map[string][]string
	violations ViolationIndex
	metrics    MetricsIndex
	module     string
}

// NewDetailPanel creates a detail panel with the given data.
func NewDetailPanel(importedBy map[string][]string, violations ViolationIndex, metrics MetricsIndex, module string) *DetailPanel {
	view := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true)
	view.SetBorder(true).
		SetTitle(" Dependencies ").
		SetBorderColor(tcell.ColorGray)

	return &DetailPanel{
		view:       view,
		importedBy: importedBy,
		violations: violations,
		metrics:    metrics,
		module:     module,
	}
}

// View returns the underlying tview.TextView.
func (d *DetailPanel) View() *tview.TextView {
	return d.view
}

// Update refreshes the detail panel for the selected package node.
func (d *DetailPanel) Update(node *PkgNode) {
	d.view.Clear()
	if node == nil {
		d.view.SetText("[gray]Select a package to view details")
		return
	}

	if node.IsLeaf {
		d.renderLeaf(node)
	} else {
		d.renderGroup(node)
	}
}

func (d *DetailPanel) renderLeaf(node *PkgNode) {
	var b strings.Builder

	fmt.Fprintf(&b, "[white::b]%s\n", node.RelPath)

	// Metrics section.
	if m, ok := d.metrics[node.FullPath]; ok {
		fmt.Fprintf(&b, "[gray]Ca:%d  Ce:%d  Instability:%.2f  Transitive:%d\n", m.Ca, m.Ce, m.Instability, m.TransitiveDependents)
	}
	b.WriteString("\n")

	// Violations section.
	d.writeViolations(&b, node.RelPath)

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

func (d *DetailPanel) renderGroup(node *PkgNode) {
	var b strings.Builder

	fmt.Fprintf(&b, "[white::b]%s\n\n", node.RelPath)

	// Collect all violations under this path.
	var allViols []rules.Violation
	d.violations.walkPath(node.RelPath, func(viols []rules.Violation) {
		allViols = append(allViols, viols...)
	})

	if len(allViols) == 0 {
		b.WriteString("[green]No violations under this path")
		d.view.SetText(b.String())
		return
	}

	// Count by severity.
	errors, warnings := 0, 0
	for _, vp := range allViols {
		if vp.EffectiveSeverity == rules.Error {
			errors++
		} else {
			warnings++
		}
	}

	// Summary.
	if errors > 0 {
		fmt.Fprintf(&b, "[red::b]%d error(s)", errors)
	}
	if warnings > 0 {
		if errors > 0 {
			b.WriteString("[white]  ")
		}
		fmt.Fprintf(&b, "[yellow::b]%d warning(s)", warnings)
	}
	b.WriteString("\n\n")

	// Sort: errors first, then warnings. Within each group, sort by file.
	sort.Slice(allViols, func(i, j int) bool {
		if allViols[i].EffectiveSeverity != allViols[j].EffectiveSeverity {
			return allViols[i].EffectiveSeverity < allViols[j].EffectiveSeverity // Error(0) before Warning(1)
		}
		return allViols[i].File < allViols[j].File
	})

	// Render errors section.
	if errors > 0 {
		b.WriteString("[red::b]── Errors ──\n")
		for _, vp := range allViols {
			if vp.EffectiveSeverity != rules.Error {
				break
			}
			fmt.Fprintf(&b, "[red]  ✗ [%s] %s\n", vp.Rule, vp.File)
			fmt.Fprintf(&b, "[gray]    %s\n", vp.Message)
			if vp.Fix != "" {
				fmt.Fprintf(&b, "[darkgray]    fix: %s\n", vp.Fix)
			}
		}
		b.WriteString("\n")
	}

	// Render warnings section.
	if warnings > 0 {
		b.WriteString("[yellow::b]── Warnings ──\n")
		for _, vp := range allViols {
			if vp.EffectiveSeverity != rules.Warning {
				continue
			}
			fmt.Fprintf(&b, "[yellow]  ⚠ [%s] %s\n", vp.Rule, vp.File)
			fmt.Fprintf(&b, "[gray]    %s\n", vp.Message)
			if vp.Fix != "" {
				fmt.Fprintf(&b, "[darkgray]    fix: %s\n", vp.Fix)
			}
		}
		b.WriteString("\n")
	}

	d.view.SetText(b.String())
}

func (d *DetailPanel) writeViolations(b *strings.Builder, relPath string) {
	viols, ok := d.violations[relPath]
	if !ok || len(viols) == 0 {
		return
	}
	sort.Slice(viols, func(i, j int) bool {
		if viols[i].EffectiveSeverity != viols[j].EffectiveSeverity {
			return viols[i].EffectiveSeverity < viols[j].EffectiveSeverity
		}
		return viols[i].File < viols[j].File
	})
	fmt.Fprintf(b, "[red::b]Violations (%d):\n", len(viols))
	for _, v := range viols {
		color := "red"
		sev := "ERR"
		if v.EffectiveSeverity == rules.Warning {
			color = "yellow"
			sev = "WARN"
		}
		fmt.Fprintf(b, "[%s]  [%s] %s\n", color, sev, v.Rule)
		fmt.Fprintf(b, "[gray]    %s\n", v.Message)
		if v.Fix != "" {
			fmt.Fprintf(b, "[darkgray]    fix: %s\n", v.Fix)
		}
	}
	b.WriteString("\n")
}
