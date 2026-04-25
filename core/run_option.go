package core

// RunOption configures a single Run invocation.
type RunOption func(*runOpts)

type runOpts struct {
	severityOverrides map[string]Severity
}

func newRunOpts(opts ...RunOption) *runOpts {
	o := &runOpts{severityOverrides: make(map[string]Severity)}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

func (o *runOpts) severityFor(violationID string) (Severity, bool) {
	s, ok := o.severityOverrides[violationID]
	return s, ok
}

// WithSeverityOverride sets the effective severity for a violation-level ID.
// When multiple overrides target the same ID, the last one passed to Run
// wins.
//
// Example:
//
//	core.Run(ctx, rules,
//	    core.WithSeverityOverride("isolation.cross-domain", core.Warning))
func WithSeverityOverride(violationID string, s Severity) RunOption {
	return func(o *runOpts) {
		o.severityOverrides[violationID] = s
	}
}
