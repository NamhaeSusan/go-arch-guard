package core

import "maps"

// RuleSet is an immutable collection of rules plus a set of violation IDs
// to filter out at the runner. With and Without return copies, so chaining
// is safe and the original RuleSet is never mutated.
type RuleSet struct {
	rules            []Rule
	skipViolationIDs map[string]bool
}

// NewRuleSet seeds a RuleSet with the given rules. Equivalent to
// RuleSet{}.With(rules...).
func NewRuleSet(rules ...Rule) RuleSet {
	return RuleSet{}.With(rules...)
}

// With returns a new RuleSet with the given rules appended. nil rules are
// silently dropped so callers can compose conditional rule lists like
// rs.With(maybeRule()) without an explicit nil check at every site.
func (rs RuleSet) With(rules ...Rule) RuleSet {
	out := rs.copy()
	for _, r := range rules {
		if r == nil {
			continue
		}
		out.rules = append(out.rules, r)
	}
	return out
}

// Without returns a new RuleSet whose runner filters out violations whose
// Rule field matches any of the given violation-level IDs (e.g.
// "isolation.cross-domain"). IDs not present in the active rule set are
// rejected by Run; see WithSeverityOverride for the same guarantee.
func (rs RuleSet) Without(violationIDs ...string) RuleSet {
	out := rs.copy()
	if out.skipViolationIDs == nil {
		out.skipViolationIDs = make(map[string]bool, len(violationIDs))
	}
	for _, id := range violationIDs {
		out.skipViolationIDs[id] = true
	}
	return out
}

// Rules returns the rules in registration order. The returned slice is a
// copy; callers may not mutate it.
func (rs RuleSet) Rules() []Rule {
	out := make([]Rule, len(rs.rules))
	copy(out, rs.rules)
	return out
}

// IsViolationSkipped reports whether Without(...) was called for id.
func (rs RuleSet) IsViolationSkipped(id string) bool {
	return rs.skipViolationIDs[id]
}

func (rs RuleSet) copy() RuleSet {
	rules := make([]Rule, len(rs.rules))
	copy(rules, rs.rules)
	var skip map[string]bool
	if len(rs.skipViolationIDs) > 0 {
		skip = make(map[string]bool, len(rs.skipViolationIDs))
		maps.Copy(skip, rs.skipViolationIDs)
	}
	return RuleSet{rules: rules, skipViolationIDs: skip}
}
