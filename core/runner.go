package core

import (
	"fmt"
	"sort"
	"strings"
)

// Run executes the RuleSet against the Context and returns violations.
//
// Contract:
//
//   - Architecture is validated before any rule runs; an invalid
//     Architecture panics (presets MAY call Validate at construction time
//     to fail earlier).
//   - Run rejects unknown violation IDs passed to RuleSet.Without or
//     WithSeverityOverride by panicking — these are caller-side errors
//     and must be surfaced loudly. The set of known IDs is the union
//     of every rule's RuleSpec.ViolationIDs() plus each rule's RuleSpec.ID
//     itself. Single-ID rules with empty Violations declare implicitly
//     through their RuleSpec.ID.
//   - For each Violation a rule emits, Run validates Violation.Rule
//     against THAT rule's own ID set. Unknown IDs (rule-author bugs) are
//     replaced with "meta.unknown-violation-id" rather than panicked —
//     a buggy rule should not crash the entire run.
//   - Rules execute serially in registration order. Rule.Check must be
//     pure; future runners may parallelize.
//   - Effective severity precedence (highest wins):
//     1. WithSeverityOverride(violationID, ...)
//     2. RuleSpec.Violations[i].DefaultSeverity for matching ID
//     3. RuleSpec.DefaultSeverity
//     4. Error
//   - Violations are deduped by Rule field for any Rule starting with "meta.".
//   - Final violations are sorted by (File, Line, Rule, Message) for
//     deterministic output regardless of map iteration order inside rules.
func Run(ctx *Context, rules RuleSet, opts ...RunOption) []Violation {
	if err := ctx.Arch().Validate(); err != nil {
		panic(fmt.Sprintf("core.Run: %v", err))
	}

	o := newRunOpts(opts...)
	known := knownViolationIDs(rules)
	checkUnknown := func(label string, ids []string) {
		var unknown []string
		for _, id := range ids {
			if !known[id] {
				unknown = append(unknown, id)
			}
		}
		if len(unknown) > 0 {
			sort.Strings(unknown)
			panic(fmt.Sprintf("core.Run: %s references unknown violation IDs: %s", label, strings.Join(unknown, ", ")))
		}
	}
	checkUnknown("RuleSet.Without", sortedKeys(rules.skipViolationIDs))
	checkUnknown("WithSeverityOverride", sortedKeys(o.severityOverrides))

	var out []Violation
	for _, r := range rules.Rules() {
		spec := r.Spec()
		violationDefaults := defaultsByID(spec)
		ruleDefault := spec.DefaultSeverity
		ruleKnown := ruleKnownIDs(spec)

		for _, v := range r.Check(ctx) {
			if !ruleKnown[v.Rule] {
				v = Violation{
					File:    v.File,
					Line:    v.Line,
					Rule:    "meta.unknown-violation-id",
					Message: fmt.Sprintf("rule %q emitted undeclared violation ID %q", spec.ID, v.Rule),
					Fix:     fmt.Sprintf("declare %q in %s.Spec().Violations or fix the typo in Check()", v.Rule, spec.ID),
				}
			}
			if rules.IsViolationSkipped(v.Rule) {
				continue
			}
			declared, ok := violationDefaults[v.Rule]
			if !ok {
				declared = ruleDefault
			}
			effective := declared
			if override, ok := o.severityFor(v.Rule); ok {
				effective = override
			}
			v.DefaultSeverity = declared
			v.EffectiveSeverity = effective
			out = append(out, v)
		}
	}

	out = dedupeMetaViolations(out)

	sort.SliceStable(out, func(i, j int) bool {
		if out[i].File != out[j].File {
			return out[i].File < out[j].File
		}
		if out[i].Line != out[j].Line {
			return out[i].Line < out[j].Line
		}
		if out[i].Rule != out[j].Rule {
			return out[i].Rule < out[j].Rule
		}
		return out[i].Message < out[j].Message
	})

	return out
}

func knownViolationIDs(rs RuleSet) map[string]bool {
	known := make(map[string]bool)
	for _, r := range rs.Rules() {
		for id := range ruleKnownIDs(r.Spec()) {
			known[id] = true
		}
	}
	return known
}

// ruleKnownIDs returns the set of violation IDs the rule may legitimately
// emit: its RuleSpec.ID (for single-ID rules) plus every ViolationSpec.ID.
func ruleKnownIDs(spec RuleSpec) map[string]bool {
	out := make(map[string]bool, 1+len(spec.Violations))
	if spec.ID != "" {
		out[spec.ID] = true
	}
	for _, v := range spec.Violations {
		out[v.ID] = true
	}
	return out
}

func defaultsByID(spec RuleSpec) map[string]Severity {
	if len(spec.Violations) == 0 {
		return nil
	}
	out := make(map[string]Severity, len(spec.Violations))
	for _, v := range spec.Violations {
		out[v.ID] = v.DefaultSeverity
	}
	return out
}

func dedupeMetaViolations(in []Violation) []Violation {
	seen := make(map[string]bool)
	out := in[:0]
	for _, v := range in {
		if strings.HasPrefix(v.Rule, "meta.") {
			if seen[v.Rule] {
				continue
			}
			seen[v.Rule] = true
		}
		out = append(out, v)
	}
	return out
}

func sortedKeys[V any](m map[string]V) []string {
	if len(m) == 0 {
		return nil
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
