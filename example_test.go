package goarchguard_test

import (
	"fmt"

	"github.com/NamhaeSusan/go-arch-guard/analyzer"
	"github.com/NamhaeSusan/go-arch-guard/rules"
)

func Example() {
	pkgs, err := analyzer.Load("testdata/valid", "internal/...", "cmd/...")
	if err != nil {
		fmt.Println(err)
		return
	}

	var violations []rules.Violation
	violations = append(violations, rules.CheckDomainIsolation(pkgs, "github.com/kimtaeyun/testproject-dc", "testdata/valid")...)
	violations = append(violations, rules.CheckLayerDirection(pkgs, "github.com/kimtaeyun/testproject-dc", "testdata/valid")...)
	violations = append(violations, rules.CheckNaming(pkgs)...)

	fmt.Printf("violations: %d\n", len(violations))
	// Output: violations: 0
}
