package core

// RuleSpec is the static-ish metadata describing a rule type. A rule type
// emits Violations whose Rule field matches one of the IDs in Violations.
//
// Single-ID rules may leave Violations empty and reuse RuleSpec.ID as the
// violation ID; multi-ID rules MUST populate Violations so callers can
// discover and override individual sub-IDs.
type RuleSpec struct {
	ID              string          // rule-type ID, e.g. "dependency.isolation"
	Description     string          // single-line summary used by --list-rules
	DefaultSeverity Severity        // fallback for emitted violations not listed in Violations
	Violations      []ViolationSpec // declarative violation-ID catalog
}

// ViolationSpec describes one violation ID emitted by a rule type.
type ViolationSpec struct {
	ID              string
	Description     string
	DefaultSeverity Severity
}

// ViolationIDs returns the IDs declared in spec.Violations, in declaration
// order. For single-ID rules with an empty Violations slice, returns nil.
func (s RuleSpec) ViolationIDs() []string {
	if len(s.Violations) == 0 {
		return nil
	}
	ids := make([]string, len(s.Violations))
	for i, v := range s.Violations {
		ids[i] = v.ID
	}
	return ids
}
