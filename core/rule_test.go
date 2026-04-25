package core

// fakeRule is a deterministic test double used by ruleset_test.go,
// runner_test.go, and example_test.go.
type fakeRule struct {
	spec       RuleSpec
	violations []Violation
}

func (r *fakeRule) Spec() RuleSpec { return r.spec }

func (r *fakeRule) Check(ctx *Context) []Violation {
	out := make([]Violation, len(r.violations))
	copy(out, r.violations)
	return out
}

// Compile-time assertion — kept here so the test build fails fast if Rule
// is removed or renamed.
var _ Rule = (*fakeRule)(nil)
