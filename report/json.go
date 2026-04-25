package report

import (
	"encoding/json"
	"io"
	"sort"
	"strings"

	"github.com/NamhaeSusan/go-arch-guard/core"
)

// jsonReportSchema versions the JSON report shape. v2 renamed `severity` to
// `effectiveSeverity` and added `defaultSeverity`; consumers should branch on
// this string before parsing severity fields.
const jsonReportSchema = "go-arch-guard.report.v2"

// JSONSummary captures aggregate counts for a machine-readable report.
type JSONSummary struct {
	Total    int      `json:"total"`
	Errors   int      `json:"errors"`
	Warnings int      `json:"warnings"`
	Files    int      `json:"files"`
	Rules    []string `json:"rules"`
}

// JSONViolation is a JSON-friendly view of a core.Violation. Every field is
// always emitted (no `omitempty`) so consumers see one stable shape per
// violation regardless of severity or violation source — meta.* violations
// without a file or line still emit `"file": ""` and `"line": 0` rather than
// dropping the keys, which would force parsers to handle two object shapes.
type JSONViolation struct {
	File              string `json:"file"`
	Line              int    `json:"line"`
	Rule              string `json:"rule"`
	Message           string `json:"message"`
	Fix               string `json:"fix"`
	EffectiveSeverity string `json:"effectiveSeverity"`
	DefaultSeverity   string `json:"defaultSeverity"`
}

// JSONReport is a stable machine-readable report for automation and AI agents.
type JSONReport struct {
	Schema     string          `json:"schema"`
	Summary    JSONSummary     `json:"summary"`
	Violations []JSONViolation `json:"violations"`
}

// BuildJSONReport converts violations into a machine-readable report. The
// violations slice is sorted by (File, Line, Rule, Message) on a defensive
// copy before serialization so the JSON output is byte-stable regardless of
// the caller's input order. core.Run already sorts the same way, but direct
// callers (custom runners, third-party rule pipelines) cannot be assumed to
// sort, and unstable JSON breaks every downstream comparison and snapshot.
func BuildJSONReport(violations []core.Violation) JSONReport {
	report := JSONReport{
		Schema:     jsonReportSchema,
		Violations: make([]JSONViolation, 0, len(violations)),
	}

	sorted := make([]core.Violation, len(violations))
	copy(sorted, violations)
	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].File != sorted[j].File {
			return sorted[i].File < sorted[j].File
		}
		if sorted[i].Line != sorted[j].Line {
			return sorted[i].Line < sorted[j].Line
		}
		if sorted[i].Rule != sorted[j].Rule {
			return sorted[i].Rule < sorted[j].Rule
		}
		return sorted[i].Message < sorted[j].Message
	})

	files := make(map[string]struct{})
	ruleSet := make(map[string]struct{})

	for _, v := range sorted {
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
