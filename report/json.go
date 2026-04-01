package report

import (
	"encoding/json"
	"io"
	"sort"
	"strings"

	"github.com/NamhaeSusan/go-arch-guard/rules"
)

const jsonReportSchema = "go-arch-guard.report.v1"

// JSONSummary captures aggregate counts for a machine-readable report.
type JSONSummary struct {
	Total    int      `json:"total"`
	Errors   int      `json:"errors"`
	Warnings int      `json:"warnings"`
	Files    int      `json:"files"`
	Rules    []string `json:"rules,omitempty"`
}

// JSONViolation is a JSON-friendly view of a rules.Violation.
type JSONViolation struct {
	File     string `json:"file,omitempty"`
	Line     int    `json:"line,omitempty"`
	Rule     string `json:"rule"`
	Message  string `json:"message"`
	Fix      string `json:"fix,omitempty"`
	Severity string `json:"severity"`
}

// JSONReport is a stable machine-readable report for automation and AI agents.
type JSONReport struct {
	Schema     string          `json:"schema"`
	Summary    JSONSummary     `json:"summary"`
	Violations []JSONViolation `json:"violations"`
}

// BuildJSONReport converts violations into a machine-readable report.
func BuildJSONReport(violations []rules.Violation) JSONReport {
	report := JSONReport{
		Schema:     jsonReportSchema,
		Violations: make([]JSONViolation, 0, len(violations)),
	}
	files := make(map[string]struct{})
	ruleSet := make(map[string]struct{})

	for _, v := range violations {
		report.Violations = append(report.Violations, JSONViolation{
			File:     v.File,
			Line:     v.Line,
			Rule:     v.Rule,
			Message:  v.Message,
			Fix:      v.Fix,
			Severity: strings.ToLower(v.Severity.String()),
		})
		report.Summary.Total++
		if v.Severity == rules.Error {
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
	for rule := range ruleSet {
		report.Summary.Rules = append(report.Summary.Rules, rule)
	}
	sort.Strings(report.Summary.Rules)
	return report
}

// MarshalJSONReport marshals a machine-readable report with stable indentation.
func MarshalJSONReport(violations []rules.Violation) ([]byte, error) {
	return json.MarshalIndent(BuildJSONReport(violations), "", "  ")
}

// WriteJSONReport writes a machine-readable report with stable indentation.
func WriteJSONReport(w io.Writer, violations []rules.Violation) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(BuildJSONReport(violations))
}
