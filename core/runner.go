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
//     a buggy rule should not crash the entire run. Emitted IDs starting
//     with "meta." are exempt from this check: any rule may emit
//     meta.* violations to surface environmental issues (e.g.
//     "meta.no-matching-packages" when the project module cannot be
//     resolved) without declaring them in its catalog.
//   - Rules execute serially in registration order. Rule.Check must be
//     pure; future runners may parallelize.
//   - Effective severity precedence (highest wins):
//     1. WithSeverityOverride(violationID, ...)
//     2. RuleSpec.Violations[i].DefaultSeverity for matching ID
//     3. Warning, when the violation ID starts with "meta." (environmental
//     meta.* violations should never block builds by accident)
//     4. RuleSpec.DefaultSeverity
//     5. Error
//   - Violations are deduped by Rule field for any Rule starting with "meta.".
//   - Final violations are sorted by (File, Line, Rule, Message) for
//     deterministic output regardless of map iteration order inside rules.
func Run(ctx *Context, rules RuleSet, opts ...RunOption) []Violation {
	if err := ctx.Arch().Validate(); err != nil {
		panic(fmt.Sprintf("core.Run: %v", err))
	}
	if err := validateRuleSet(rules); err != nil {
		panic(fmt.Sprintf("core.Run: %v", err))
	}

	o := newRunOpts(opts...)
	known := knownViolationIDs(rules)
	checkUnknown := func(label string, ids []string) {
		var unknown []string
		for _, id := range ids {
			// meta.* IDs are not declared in any rule's catalog (they're
			// emergency emits, see Check loop below). Callers may legitimately
			// filter or override them, so allow them through unconditionally.
			if !known[id] && !strings.HasPrefix(id, "meta.") {
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

		// panicked is intentionally discarded; the meta.rule-panic
		// violation itself carries the signal through the standard
		// severity / dedup pipeline below.
		emitted, _ := safeCheck(r, ctx)
		for _, v := range emitted {
			// meta.* IDs are emergency emit by any rule (e.g. "meta.no-matching-packages"
			// when the project module cannot be resolved) and are not part of any
			// rule's catalog. Allow them through unconditionally.
			if !ruleKnown[v.Rule] && !strings.HasPrefix(v.Rule, "meta.") {
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
				// meta.* IDs are environmental warnings (e.g.
				// meta.no-matching-packages). Default them to Warning so a
				// dependency rule that defaults to Error doesn't accidentally
				// promote a configuration mismatch into a hard failure.
				if strings.HasPrefix(v.Rule, "meta.") {
					declared = Warning
				} else {
					declared = ruleDefault
				}
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

// validateRuleSet checks for duplicate rule-type IDs across the set, and
// duplicate violation IDs within any single rule's Violations catalog.
// Both are silent foot-guns: dups collapse into a map so filters and
// overrides cannot disambiguate them.
func validateRuleSet(rs RuleSet) error {
	var errs []string

	seenRule := make(map[string]bool)
	for _, r := range rs.Rules() {
		spec := r.Spec()
		if spec.ID == "" {
			continue
		}
		if seenRule[spec.ID] {
			errs = append(errs, fmt.Sprintf("RuleSet contains duplicate rule-type ID %q", spec.ID))
		}
		seenRule[spec.ID] = true

		seenViolation := make(map[string]bool, len(spec.Violations))
		for _, v := range spec.Violations {
			if seenViolation[v.ID] {
				errs = append(errs, fmt.Sprintf("rule %q declares duplicate violation ID %q in Spec().Violations", spec.ID, v.ID))
			}
			seenViolation[v.ID] = true
		}
	}

	if len(errs) == 0 {
		return nil
	}
	sort.Strings(errs)
	return fmt.Errorf("%s", strings.Join(errs, "; "))
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

// safeCheck invokes rule.Check inside a deferred recover. A panic from a
// rule is converted into a single meta.rule-panic violation so a buggy
// rule does not crash the entire run; other rules continue to execute.
// The returned violation goes through the runner's normal meta.* path:
// default severity is Warning (so a buggy rule does not block the build),
// dedup collapses repeated panics from the same rule, and callers can
// promote it to Error with WithSeverityOverride("meta.rule-panic", Error).
func safeCheck(rule Rule, ctx *Context) (violations []Violation, panicked bool) {
	defer func() {
		if rec := recover(); rec != nil {
			ruleID := "<unknown>"
			func() {
				defer func() { _ = recover() }()
				ruleID = rule.Spec().ID
			}()
			violations = []Violation{{
				Rule:    "meta.rule-panic",
				Message: fmt.Sprintf("rule %q panicked: %v", ruleID, rec),
				Fix:     "report the panic to the rule author; other rules continue to run",
			}}
			panicked = true
		}
	}()
	return rule.Check(ctx), false
}

// dedupeMetaViolations collapses duplicate meta.* violations. The dedup key
// is the (Rule, Message) pair, not just Rule alone — two rules can emit the
// same meta ID with different messages (e.g. "module not determined" vs
// "module does not match any loaded package") and both signals should reach
// the user. Exact duplicates are still collapsed.
func dedupeMetaViolations(in []Violation) []Violation {
	type metaKey struct {
		Rule    string
		Message string
	}
	seen := make(map[metaKey]bool)
	out := in[:0]
	for _, v := range in {
		if strings.HasPrefix(v.Rule, "meta.") {
			key := metaKey{Rule: v.Rule, Message: v.Message}
			if seen[key] {
				continue
			}
			seen[key] = true
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
