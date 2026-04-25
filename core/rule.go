package core

// Rule is the contract every architecture rule satisfies. Implementations
// MUST be pure: Check must not mutate Context and must be safe to call
// multiple times with the same Context. Run executes rules serially, but
// the contract is purity so future runners may parallelize.
type Rule interface {
	// Spec returns the rule's metadata. Spec MAY reflect construction-time
	// configuration (e.g. severity passed via WithSeverity); it is not
	// required to be a pure function of the type.
	Spec() RuleSpec

	// Check inspects the context and returns zero or more violations.
	// The returned slice is owned by the caller; rules must not retain it.
	Check(ctx *Context) []Violation
}
