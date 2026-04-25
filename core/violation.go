package core

import "fmt"

// Violation is a single rule failure emitted by Rule.Check. The Rule field
// is the violation-level ID (e.g. "isolation.cross-domain"), not the
// rule-type ID. DefaultSeverity records what the rule declared in its
// ViolationSpec; EffectiveSeverity is what callers see after construction-
// and runtime-level overrides have been applied by Run.
type Violation struct {
	File              string
	Line              int
	Rule              string
	Message           string
	Fix               string
	DefaultSeverity   Severity
	EffectiveSeverity Severity
}

func (v Violation) String() string {
	fileStr := v.File
	if v.Line > 0 {
		fileStr = fmt.Sprintf("%s:%d", v.File, v.Line)
	}
	return fmt.Sprintf("[%s] violation: %s (file: %s, rule: %s, fix: %s)",
		v.EffectiveSeverity, v.Message, fileStr, v.Rule, v.Fix)
}
