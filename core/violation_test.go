package core

import "testing"

func TestViolationStringWithLine(t *testing.T) {
	v := Violation{
		File:              "internal/handler/foo.go",
		Line:              42,
		Rule:              "isolation.cross-domain",
		Message:           "must not import sibling",
		Fix:               "move to orchestration",
		DefaultSeverity:   Error,
		EffectiveSeverity: Warning,
	}
	got := v.String()
	want := "[WARNING] violation: must not import sibling (file: internal/handler/foo.go:42, rule: isolation.cross-domain, fix: move to orchestration)"
	if got != want {
		t.Errorf("Violation.String() =\n  %q\nwant:\n  %q", got, want)
	}
}

func TestViolationStringWithoutLine(t *testing.T) {
	v := Violation{
		File:              "internal/",
		Rule:              "meta.no-matching-packages",
		Message:           "no packages matched",
		EffectiveSeverity: Error,
	}
	if got := v.String(); got != "[ERROR] violation: no packages matched (file: internal/, rule: meta.no-matching-packages, fix: )" {
		t.Errorf("unexpected: %q", got)
	}
}
