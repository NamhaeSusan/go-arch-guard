package rules

import (
	"golang.org/x/tools/go/packages"
)

// CheckTxBoundary gates where transactions may start and prevents transaction
// types from leaking into function signatures outside allowed layers.
// Fully opt-in: returns nil when TxBoundaryConfig has neither StartSymbols
// nor Types configured.
func CheckTxBoundary(
	pkgs []*packages.Package,
	projectModule string,
	projectRoot string,
	opts ...Option,
) []Violation {
	cfg := NewConfig(opts...)
	tc := cfg.TxBoundary
	if len(tc.StartSymbols) == 0 && len(tc.Types) == 0 {
		return nil
	}
	allowed := tc.AllowedLayers
	if len(allowed) == 0 {
		allowed = []string{"app"}
	}
	m := cfg.model()

	var violations []Violation

	if len(tc.StartSymbols) > 0 {
		violations = append(violations,
			checkForbiddenCallsByLayer(pkgs, projectModule, projectRoot, m, cfg,
				[]forbiddenCallRule{{
					Symbols:       tc.StartSymbols,
					AllowedLayers: allowed,
					RuleName:      "tx.start-outside-allowed-layer",
					Message:       "transaction must not start here; allowed layers: %v",
					Fix:           "move the transaction-starting call into an allowed layer: %v",
				}})...,
		)
	}

	if len(tc.Types) > 0 {
		violations = append(violations,
			checkTypeInSignature(pkgs, projectModule, projectRoot, m, cfg,
				tc.Types, allowed,
				"tx.type-in-signature",
				"tx type %q must not appear in function signature outside allowed layers: %v",
				"keep %q confined to allowed layers: %v",
			)...,
		)
	}

	return violations
}
