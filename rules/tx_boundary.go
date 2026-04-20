package rules

import (
	"fmt"

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
	scope := scanScope{
		enforceUnclassified: tc.EnforceUnclassified,
		enforceCmdRoot:      tc.EnforceCmdRoot,
	}

	var violations []Violation

	if len(tc.StartSymbols) > 0 {
		violations = append(violations,
			checkForbiddenCallsByLayer(pkgs, projectModule, projectRoot, m, cfg, scope,
				[]forbiddenCallRule{{
					Symbols:       tc.StartSymbols,
					AllowedLayers: allowed,
					RuleName:      "tx.start-outside-allowed-layer",
					Message: func(layer string, allowed []string) string {
						return fmt.Sprintf("transaction must not start in layer %q; allowed layers: %v", layer, allowed)
					},
					Fix: func(layer string, allowed []string) string {
						return fmt.Sprintf("move the transaction-starting call out of %q into an allowed layer: %v", layer, allowed)
					},
				}})...,
		)
	}

	if len(tc.Types) > 0 {
		violations = append(violations,
			checkTypeInSignature(pkgs, projectModule, projectRoot, m, cfg, scope,
				tc.Types, allowed,
				"tx.type-in-signature",
				func(typeID string, allowed []string) string {
					return fmt.Sprintf("tx type %q must not appear in function signature outside allowed layers: %v", typeID, allowed)
				},
				func(typeID string, allowed []string) string {
					return fmt.Sprintf("keep %q confined to allowed layers: %v", typeID, allowed)
				},
			)...,
		)
	}

	return violations
}
