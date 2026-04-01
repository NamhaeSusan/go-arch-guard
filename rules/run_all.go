package rules

import "golang.org/x/tools/go/packages"

// RunAll executes the recommended built-in rule set and returns the merged
// violations in a stable order. Empty module/root values are auto-extracted
// from the loaded packages the same way the lower-level import rules behave.
func RunAll(pkgs []*packages.Package, projectModule string, projectRoot string, opts ...Option) []Violation {
	projectModule = resolveModule(pkgs, projectModule)
	projectRoot = resolveRoot(pkgs, projectRoot)

	var violations []Violation
	violations = append(violations, CheckDomainIsolation(pkgs, projectModule, projectRoot, opts...)...)
	violations = append(violations, CheckLayerDirection(pkgs, projectModule, projectRoot, opts...)...)
	violations = append(violations, CheckNaming(pkgs, opts...)...)
	violations = append(violations, CheckStructure(projectRoot, opts...)...)
	violations = append(violations, AnalyzeBlastRadius(pkgs, projectModule, projectRoot, opts...)...)
	return violations
}
