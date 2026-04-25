package report_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/report"
)

func TestBuildJSONReport(t *testing.T) {
	violations := []core.Violation{
		{
			File:              "internal/domain/order/app/service.go",
			Line:              12,
			Rule:              "isolation.cross-domain",
			Message:           "bad import",
			Fix:               "move orchestration",
			DefaultSeverity:   core.Error,
			EffectiveSeverity: core.Error,
		},
		{
			File:              "internal/domain/order/app/service.go",
			Rule:              "blast.high-coupling",
			Message:           "too central",
			Fix:               "extract boundary",
			DefaultSeverity:   core.Error,
			EffectiveSeverity: core.Warning,
		},
	}

	got := report.BuildJSONReport(violations)
	if got.Schema != "go-arch-guard.report.v2" {
		t.Fatalf("unexpected schema marker: %q", got.Schema)
	}
	if got.Summary.Total != 2 || got.Summary.Errors != 1 || got.Summary.Warnings != 1 {
		t.Fatalf("unexpected summary counts: %+v", got.Summary)
	}
	if got.Summary.Files != 1 {
		t.Fatalf("expected one unique file, got %d", got.Summary.Files)
	}
	if len(got.Summary.Rules) != 2 || got.Summary.Rules[0] != "blast.high-coupling" || got.Summary.Rules[1] != "isolation.cross-domain" {
		t.Fatalf("unexpected rules summary: %+v", got.Summary.Rules)
	}
	// Output is sorted by (File, Line, Rule, Message). Both violations share
	// the same File, so the Line=0 entry (blast.high-coupling) sorts ahead
	// of the Line=12 entry (isolation.cross-domain).
	if got.Violations[0].Rule != "blast.high-coupling" || got.Violations[1].Rule != "isolation.cross-domain" {
		t.Fatalf("violations not sorted: %+v", got.Violations)
	}
	if got.Violations[0].EffectiveSeverity != "warning" || got.Violations[1].EffectiveSeverity != "error" {
		t.Fatalf("unexpected effective severities: %+v", got.Violations)
	}
	if got.Violations[0].DefaultSeverity != "error" || got.Violations[1].DefaultSeverity != "error" {
		t.Fatalf("unexpected severities: %+v", got.Violations)
	}
}

func TestMarshalJSONReport(t *testing.T) {
	violations := []core.Violation{{
		Rule:              "test.rule",
		Message:           "bad",
		DefaultSeverity:   core.Error,
		EffectiveSeverity: core.Error,
	}}
	data, err := report.MarshalJSONReport(violations)
	if err != nil {
		t.Fatalf("MarshalJSONReport() error = %v", err)
	}

	var decoded report.JSONReport
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json output must decode: %v\n%s", err, data)
	}
	if decoded.Schema != "go-arch-guard.report.v2" || decoded.Summary.Total != 1 || decoded.Violations[0].EffectiveSeverity != "error" || decoded.Violations[0].DefaultSeverity != "error" {
		t.Fatalf("unexpected decoded report: %+v", decoded)
	}
}

func TestBuildJSONReport_NilViolations(t *testing.T) {
	got := report.BuildJSONReport(nil)
	if got.Summary.Total != 0 || got.Summary.Errors != 0 || got.Summary.Warnings != 0 {
		t.Fatalf("nil input must produce zero counts: %+v", got.Summary)
	}
	data, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	if !bytes.Contains(data, []byte(`"violations":[]`)) {
		t.Fatalf("nil violations must marshal as [] not null: %s", data)
	}
}

func TestBuildJSONReport_MetaViolationStableShape(t *testing.T) {
	// Meta violations (e.g. meta.no-matching-packages, meta.rule-panic)
	// often have empty File and zero Line. The JSON output must still emit
	// the file/line/fix keys with zero values so consumers see one stable
	// shape per violation. omitempty on these fields would force parsers
	// to handle two object shapes.
	violations := []core.Violation{{
		Rule:              "meta.rule-panic",
		Message:           `rule "fake.bombs" panicked: boom`,
		DefaultSeverity:   core.Error,
		EffectiveSeverity: core.Error,
	}}
	data, err := report.MarshalJSONReport(violations)
	if err != nil {
		t.Fatalf("MarshalJSONReport() error = %v", err)
	}
	for _, frag := range []string{`"file": ""`, `"line": 0`, `"fix": ""`, `"rule": "meta.rule-panic"`} {
		if !bytes.Contains(data, []byte(frag)) {
			t.Errorf("meta violation JSON missing %q\n%s", frag, data)
		}
	}
}

func TestBuildJSONReport_SortsRegardlessOfInputOrder(t *testing.T) {
	// Regression guard: BuildJSONReport must produce byte-stable output for
	// equivalent inputs. core.Run already sorts, but BuildJSONReport is
	// public and accepts any caller-provided slice.
	a := []core.Violation{
		{File: "z.go", Line: 1, Rule: "r2", Message: "m1", DefaultSeverity: core.Error, EffectiveSeverity: core.Error},
		{File: "a.go", Line: 5, Rule: "r1", Message: "m1", DefaultSeverity: core.Error, EffectiveSeverity: core.Error},
		{File: "a.go", Line: 5, Rule: "r1", Message: "m0", DefaultSeverity: core.Error, EffectiveSeverity: core.Error},
	}
	b := []core.Violation{
		{File: "a.go", Line: 5, Rule: "r1", Message: "m0", DefaultSeverity: core.Error, EffectiveSeverity: core.Error},
		{File: "z.go", Line: 1, Rule: "r2", Message: "m1", DefaultSeverity: core.Error, EffectiveSeverity: core.Error},
		{File: "a.go", Line: 5, Rule: "r1", Message: "m1", DefaultSeverity: core.Error, EffectiveSeverity: core.Error},
	}

	dataA, err := report.MarshalJSONReport(a)
	if err != nil {
		t.Fatal(err)
	}
	dataB, err := report.MarshalJSONReport(b)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(dataA, dataB) {
		t.Fatalf("BuildJSONReport must be order-insensitive\nA:\n%s\nB:\n%s", dataA, dataB)
	}
}

func TestBuildJSONReport_DoesNotMutateInput(t *testing.T) {
	// BuildJSONReport sorts on a defensive copy. The caller's slice must
	// not be reordered.
	violations := []core.Violation{
		{File: "z.go", Rule: "r2", Message: "m1", DefaultSeverity: core.Error, EffectiveSeverity: core.Error},
		{File: "a.go", Rule: "r1", Message: "m0", DefaultSeverity: core.Error, EffectiveSeverity: core.Error},
	}
	_ = report.BuildJSONReport(violations)
	if violations[0].File != "z.go" || violations[1].File != "a.go" {
		t.Fatalf("input slice mutated: %+v", violations)
	}
}

func TestWriteJSONReport(t *testing.T) {
	var buf bytes.Buffer
	violations := []core.Violation{{
		Rule:              "test.rule",
		Message:           "warn",
		DefaultSeverity:   core.Error,
		EffectiveSeverity: core.Warning,
	}}
	if err := report.WriteJSONReport(&buf, violations); err != nil {
		t.Fatalf("WriteJSONReport() error = %v", err)
	}
	if buf.Len() == 0 {
		t.Fatal("expected JSON output")
	}

	var decoded report.JSONReport
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("written JSON must decode: %v\n%s", err, buf.String())
	}
	if decoded.Summary.Warnings != 1 || decoded.Violations[0].EffectiveSeverity != "warning" || decoded.Violations[0].DefaultSeverity != "error" {
		t.Fatalf("unexpected written report: %+v", decoded)
	}
}
