package goarchguard_test

import (
	"fmt"
	"testing"

	"github.com/kimtaeyun/go-arch-guard/analyzer"
	"github.com/kimtaeyun/go-arch-guard/report"
	"github.com/kimtaeyun/go-arch-guard/rules"
)

func Example() {
	// Load packages from your project
	pkgs, err := analyzer.Load("testdata/valid", "internal/...")
	if err != nil {
		fmt.Println(err)
		return
	}

	// Check all rules
	var violations []rules.Violation
	violations = append(violations, rules.CheckDependency(pkgs, "github.com/kimtaeyun/testproject", "testdata/valid")...)
	violations = append(violations, rules.CheckNaming(pkgs)...)
	violations = append(violations, rules.CheckStructure("testdata/valid")...)

	fmt.Printf("violations: %d\n", len(violations))
	// Output: violations: 0
}

func ExampleWithSeverity() {
	// Use Warning severity for gradual adoption
	pkgs, _ := analyzer.Load("testdata/invalid", "internal/...")

	violations := rules.CheckDependency(pkgs, "github.com/kimtaeyun/testproject-invalid", "testdata/invalid",
		rules.WithSeverity(rules.Warning))

	// AssertNoViolations passes with warnings — only errors fail tests
	t := &testing.T{}
	report.AssertNoViolations(t, violations)

	fmt.Printf("warnings found: %d, test failed: %v\n", len(violations), t.Failed())
	// Output: warnings found: 3, test failed: false
}
