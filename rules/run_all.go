package rules

import "golang.org/x/tools/go/packages"

// RunAll executes the recommended built-in rule set and returns violations
// in rule-execution order (domain isolation, layer direction, naming, structure,
// blast radius). Empty module/root values are auto-extracted from the loaded
// packages the same way the lower-level import rules behave.
func RunAll(pkgs []*packages.Package, projectModule string, projectRoot string, opts ...Option) []Violation {
	// Pre-resolve so CheckStructure (which only takes root, not pkgs) gets the
	// auto-extracted value. The other functions will re-resolve as a no-op.
	projectModule = resolveModule(pkgs, projectModule)
	projectRoot = resolveRoot(pkgs, projectRoot)

	// Apply defaults: limit repo interface methods for models with repo sublayers.
	cfg := NewConfig(opts...)
	if cfg.MaxRepoInterfaceMethods == 0 && hasPortSublayer(cfg.model()) {
		opts = append([]Option{WithMaxRepoInterfaceMethods(10)}, opts...)
	}

	var violations []Violation
	violations = append(violations, CheckDomainIsolation(pkgs, projectModule, projectRoot, opts...)...)
	violations = append(violations, CheckLayerDirection(pkgs, projectModule, projectRoot, opts...)...)
	violations = append(violations, CheckNaming(pkgs, opts...)...)
	violations = append(violations, CheckStructure(projectRoot, opts...)...)
	violations = append(violations, CheckTypePatterns(pkgs, opts...)...)
	violations = append(violations, CheckInterfacePattern(pkgs, opts...)...)
	violations = append(violations, CheckContainerInterface(pkgs, opts...)...)
	violations = append(violations, CheckCrossDomainAnonymous(pkgs, opts...)...)
	violations = append(violations, AnalyzeBlastRadius(pkgs, projectModule, projectRoot, opts...)...)
	return deduplicateMetaViolations(violations)
}
