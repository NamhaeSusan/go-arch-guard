package core

import "testing"

type auditRule struct {
	spec RuleSpec
	id   string
}

func (r auditRule) Spec() RuleSpec             { return r.spec }
func (r auditRule) Check(*Context) []Violation { return []Violation{{Rule: r.id}} }
func TestUnknownViolationFailsClosed(t *testing.T) {
	v := Run(NewContext(nil, "", "", Architecture{}, nil), NewRuleSet(auditRule{RuleSpec{ID: "demo"}, "typo"}))
	if len(v) != 1 || v[0].EffectiveSeverity != Error {
		t.Fatalf("got %+v", v)
	}
}
func TestInvalidSeverityRejected(t *testing.T) {
	for _, where := range []string{"rule", "violation", "override"} {
		t.Run(where, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("expected invalid-severity panic")
				}
			}()
			spec := RuleSpec{ID: "demo"}
			var opts []RunOption
			switch where {
			case "rule":
				spec.DefaultSeverity = Severity(2)
			case "violation":
				spec.Violations = []ViolationSpec{{ID: "demo", DefaultSeverity: Severity(-1)}}
			case "override":
				opts = []RunOption{WithSeverityOverride("demo", Severity(2))}
			}
			Run(NewContext(nil, "", "", Architecture{}, nil), NewRuleSet(auditRule{spec, "demo"}), opts...)
		})
	}
}
