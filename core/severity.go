// Package core defines the rule interface and runtime for go-arch-guard.
//
// External rule authors implement the Rule interface and pass instances
// into a RuleSet, which is executed by Run against a Context.
package core

// Severity classifies how loud a Violation is. Error blocks builds; Warning
// is advisory only.
type Severity int

const (
	Error Severity = iota
	Warning
)

func (s Severity) String() string {
	if s == Warning {
		return "WARNING"
	}
	return "ERROR"
}
