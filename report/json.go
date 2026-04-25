package report

import (
	"encoding/json"
	"io"
	"sort"
	"strings"

	"github.com/NamhaeSusan/go-arch-guard/core"
)

const jsonReportSchema = "go-arch-guard.report.v1"

// JSONSummary captures aggregate counts for a machine-readable report.
type JSONSummary struct {
	Total    int      `json:"total"`
	Errors   int      `json:"errors"`
	Warnings int      `json:"warnings"`
	Files    int      `json:"files"`
	Rules    []string `json:"rules"`
}

// JSONViolation is a JSON-friendly view of a core.Violation.
type JSONViolation struct {
	File              string `json:"file,omitempty"`
	Line              int    `json:"line,omitempty"`
	Rule              string `json:"rule"`
	Message           string `json:"message"`
	Fix               string `json:"fix,omitempty"`
	EffectiveSeverity string `json:"effectiveSeverity"`
	DefaultSeverity   string `json:"defaultSeverity"`
}

// JSONReport is a stable machine-readable report for automation and AI agents.
type JSONReport struct {
	Schema     string          `json:"schema"`
	Summary    JSONSummary     `json:"summary"`
	Violations []JSONViolation `json:"violations"`
}

// BuildJSONReport converts violations into a machine-readable report.
func BuildJSONReport(violations []core.Violation) JSONReport {
	report := JSONReport{
		Schema:     jsonReportSchema,
		Violations: make([]JSONViolation, 0, len(violations)),
	}
	files := make(map[string]struct{})
	ruleSet := make(map[string]struct{})

	for _, v := range violations {
		report.Violations = append(report.Violations, JSONViolation{
			File:              v.File,
			Line:              v.Line,
			Rule:              v.Rule,
			Message:           v.Message,
			Fix:               v.Fix,
			EffectiveSeverity: strings.ToLower(v.EffectiveSeverity.String()),
			DefaultSeverity:   strings.ToLower(v.DefaultSeverity.String()),
		})
		report.Summary.Total++
		if v.EffectiveSeverity == core.Error {
			report.Summary.Errors++
		} else {
			report.Summary.Warnings++
		}
		if v.File != "" {
			files[v.File] = struct{}{}
		}
		if v.Rule != "" {
			ruleSet[v.Rule] = struct{}{}
		}
	}

	report.Summary.Files = len(files)
	report.Summary.Rules = make([]string, 0, len(ruleSet))
	for rule := range ruleSet {
		report.Summary.Rules = append(report.Summary.Rules, rule)
	}
	sort.Strings(report.Summary.Rules)
	return report
}

// MarshalJSONReport marshals a machine-readable report with stable indentation.
func MarshalJSONReport(violations []core.Violation) ([]byte, error) {
	return json.MarshalIndent(BuildJSONReport(violations), "", "  ")
}

// WriteJSONReport writes a machine-readable report with stable indentation.
// Unlike MarshalJSONReport, the output includes a trailing newline.
func WriteJSONReport(w io.Writer, violations []core.Violation) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(BuildJSONReport(violations))
}
